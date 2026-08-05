package collect

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// Engine runs a collect job for one source.
type Engine struct {
	collectRepo  repository.CollectRepository
	videoRepo    repository.VideoRepository
	categoryRepo repository.CategoryRepository
	metaRepo     repository.MetadataRepository
	playRepo     repository.PlayRepository
	fetcher      *Fetcher
	log          *zap.Logger
}

// NewEngine creates an Engine.
func NewEngine(
	collectRepo repository.CollectRepository,
	videoRepo repository.VideoRepository,
	categoryRepo repository.CategoryRepository,
	metaRepo repository.MetadataRepository,
	playRepo repository.PlayRepository,
	log *zap.Logger,
) *Engine {
	if log == nil {
		log = zap.NewNop()
	}
	return &Engine{
		collectRepo:  collectRepo,
		videoRepo:    videoRepo,
		categoryRepo: categoryRepo,
		metaRepo:     metaRepo,
		playRepo:     playRepo,
		fetcher:      NewFetcher(log),
		log:          log,
	}
}

// Result is the summary of one collect run.
type Result struct {
	CollectCount int
	HasError     bool
	Message      string
}

// Run executes collection for a source (all pages until page_count or vod_time out of range).
// dataRange filters by time (today/last1d/last3d/last1w/last1m/all).
// logID is the collect_logs row ID for incremental count updates.
func (e *Engine) Run(ctx context.Context, source *model.CollectSources, dataRange string, logID uint32) Result {
	res := Result{}

	collector, err := newCollector(source, e.fetcher, e.log)
	if err != nil {
		res.HasError = true
		res.Message = err.Error()
		return res
	}

	maps, err := e.collectRepo.ListCategories(ctx, source.ID)
	if err != nil {
		e.log.Error("collect: list categories failed", zap.Uint32("source_id", source.ID), zap.Error(err))
		res.HasError = true
		res.Message = err.Error()
		return res
	}
	catMap := map[uint32]uint32{}
	for _, m := range maps {
		if m.ExternalCategoryID == 0 {
			continue
		}
		catMap[m.ExternalCategoryID] = m.CategoryID
	}

	cats, err := e.categoryRepo.List(ctx, false)
	if err != nil {
		e.log.Error("collect: list all categories for parent map failed", zap.Uint32("source_id", source.ID), zap.Error(err))
		res.HasError = true
		res.Message = err.Error()
		return res
	}
	parentMap := map[uint32]uint32{}
	for _, c := range cats {
		parentMap[c.ID] = c.ParentID
	}

	cutoffTime := utils.DataRangeCutoff(dataRange)

	// Phase 1: collect all ids from list pages
	var allIDs []uint32
	pageNo := 1
	maxPages := 50
	for pageNo <= maxPages {
		if err := ctx.Err(); err != nil {
			res.Message = "采集已取消"
			return res
		}
		listPage, err := collector.FetchListPage(ctx, source, pageNo, dataRange)
		if err != nil {
			e.log.Error("collect: fetch list failed", zap.Uint32("source_id", source.ID), zap.Int("page", pageNo), zap.Error(err))
			res.HasError = true
			res.Message = fmt.Sprintf("拉取列表第%d页失败: %v", pageNo, err)
			return res
		}
		if len(listPage.IDs) == 0 {
			break
		}

		// check time for time range filtering: stop if the last item is before cutoff
		if cutoffTime != nil && len(listPage.Times) > 0 {
			lastTime := listPage.Times[len(listPage.Times)-1]
			if lastTime != "" && isTimeBefore(lastTime, cutoffTime) {
				// still add IDs whose time is within range
				for i, t := range listPage.Times {
					if t != "" && isTimeBefore(t, cutoffTime) {
						break
					}
					allIDs = append(allIDs, listPage.IDs[i])
				}
				break
			}
		}

		allIDs = append(allIDs, listPage.IDs...)

		if pageNo >= listPage.PageCount || pageNo >= maxPages {
			break
		}
		pageNo++
	}

	if len(allIDs) == 0 {
		return res
	}

	// Phase 2: fetch details in batches of 25 and process
	batchSize := 25
	for i := 0; i < len(allIDs); i += batchSize {
		if err := ctx.Err(); err != nil {
			res.Message = "采集已取消"
			return res
		}
		end := i + batchSize
		if end > len(allIDs) {
			end = len(allIDs)
		}
		batch := allIDs[i:end]

		page, err := collector.FetchDetail(ctx, source, batch)
		if err != nil {
			e.log.Error("collect: fetch detail failed", zap.Uint32("source_id", source.ID), zap.Int("batch_start", i), zap.Int("batch_end", end), zap.Error(err))
			res.HasError = true
			res.Message = fmt.Sprintf("拉取详情失败(batch %d-%d): %v", i, end, err)
			return res
		}

		stopCollection := false
		pageCollected := 0
		for _, item := range page.Items {
			if err := ctx.Err(); err != nil {
				res.Message = "采集已取消"
				stopCollection = true
				break
			}

			// time filtering: if item time is before cutoff, stop all subsequent processing
			if cutoffTime != nil && item.VodTime != "" {
				if isTimeBefore(item.VodTime, cutoffTime) {
					stopCollection = true
					break
				}
			}

			if err := e.upsertItem(ctx, source, catMap, parentMap, item); err != nil {
				e.log.Warn("collect item failed",
					zap.Uint32("source_id", source.ID),
					zap.String("title", item.Title),
					zap.Error(err),
				)
				continue
			}
			pageCollected++
		}

		if logID > 0 && pageCollected > 0 {
			if err := e.collectRepo.IncrementLogCount(ctx, logID, pageCollected); err != nil {
				e.log.Warn("increment log count failed", zap.Error(err))
			}
		}
		res.CollectCount += pageCollected

		if stopCollection {
			break
		}
	}
	return res
}

// isTimeBefore checks if a time string (format "2006-01-02 15:04:05" or "2006-01-02") is before cutoff.
func isTimeBefore(v string, cutoff *time.Time) bool {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(v), time.Local)
	if err != nil {
		// try date-only format
		t, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(v), time.Local)
		if err != nil {
			return false
		}
	}
	return t.Before(*cutoff)
}

// extractHost parses a URL and returns its host (authority, e.g. "example.com" or "example.com:8080").
// Returns empty string for empty input, URLs without scheme, or parse errors.
// Used for detecting remote domain changes between collected data and stored data.
func extractHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}

func (e *Engine) upsertItem(ctx context.Context, source *model.CollectSources, catMap map[uint32]uint32, parentMap map[uint32]uint32, item Item) error {
	if item.ExternalCategoryID == 0 {
		return fmt.Errorf("外部分类ID无效")
	}
	categoryID, ok := catMap[item.ExternalCategoryID]
	if !ok {
		// 未绑定分类
		return nil
	}
	if item.Title == "" {
		return fmt.Errorf("标题为空")
	}

	existing, err := e.collectRepo.FindVideoByTitleYear(ctx, item.Title, item.Year)
	if err != nil {
		return err
	}

	serialStatus := parseSerialStatus(item.Remarks)

	// Resolve target category fields based on whether the bound system category is
	// a top-level (一级) or second-level (二级) category.
	catID, parentCatID := resolveCategoryFields(categoryID, parentMap)

	// Cross-source: existing video was collected by a different source.
	// Supplement empty basic fields (no cover, no associations, no PublishStatus) + upsert episodes.
	if existing != nil && existing.CollectSourceID != 0 && existing.CollectSourceID != source.ID {
		return e.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
			vRepo := e.videoRepo.WithTx(tx)
			pRepo := e.playRepo.WithTx(tx)
			applySupplementFields(existing, item, catID, parentCatID, serialStatus, false)
			if updateErr := vRepo.Update(ctx, existing); updateErr != nil {
				return updateErr
			}
			return e.upsertEpisodes(ctx, pRepo, source, existing.ID, item)
		})
	}

	// Same source or manually created (CollectSourceID==0): ensure metadata names
	directorIDs, err := e.ensureDirectors(ctx, item.Directors)
	if err != nil {
		return err
	}
	actorIDs, err := e.ensureActors(ctx, item.Actors)
	if err != nil {
		return err
	}
	tagIDs, err := e.ensureTags(ctx, item.Tags)
	if err != nil {
		return err
	}

	// Same-source domain migration: detect remote host changes and batch-migrate
	// all historical records of this source. Non-fatal: warn on failure.
	if existing != nil {
		if err := e.migrateDomainIfChanged(ctx, source, existing, item); err != nil {
			e.log.Warn("collect: domain migration failed",
				zap.Uint32("source_id", source.ID),
				zap.String("title", item.Title),
				zap.Error(err),
			)
		}
	}

	return e.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		vRepo := e.videoRepo.WithTx(tx)
		pRepo := e.playRepo.WithTx(tx)
		var videoID uint32
		if existing != nil {
			videoID = existing.ID
			// Manually created (CollectSourceID==0): claim it for this source
			if existing.CollectSourceID == 0 {
				existing.CollectSourceID = source.ID
			}
			// Same source: supplement empty basic fields + override cover (capture domain/path changes)
			applySupplementFields(existing, item, catID, parentCatID, serialStatus, true)
			if err := vRepo.Update(ctx, existing); err != nil {
				return err
			}
		} else {
			v := &model.Videos{
				Title:            item.Title,
				Subtitle:         item.Subtitle,
				Description:      item.Description,
				CategoryID:       catID,
				ParentCategoryID: parentCatID,
				PublishStatus:    constant.PublishStatusOnline,
				SerialStatus:     serialStatus,
				CoverImage:       item.Cover,
				PosterImage:      "",
				Year:             utils.Int32ToUint32(item.Year),
				Region:           item.Region,
				Language:         item.Language,
				Duration:         utils.Int32ToUint32(item.Duration),
				ReleaseDate:      strings.TrimSpace(item.ReleaseDate),
				CollectSourceID:  source.ID,
			}
			if v.SerialStatus == 0 {
				v.SerialStatus = constant.SerialStatusFinished
			}
			if err := vRepo.Create(ctx, v); err != nil {
				return err
			}
			videoID = v.ID
		}

		if err := vRepo.ReplaceDirectors(ctx, videoID, directorIDs); err != nil {
			return err
		}
		actorRels := make([]model.VideoActors, 0, len(actorIDs))
		for _, id := range actorIDs {
			actorRels = append(actorRels, model.VideoActors{VideoID: videoID, ActorID: id})
		}
		if err := vRepo.ReplaceActors(ctx, videoID, actorRels); err != nil {
			return err
		}
		if err := vRepo.ReplaceTags(ctx, videoID, tagIDs); err != nil {
			return err
		}

		// episodes into bound play source
		for _, ep := range item.Episodes {
			if ep.URL == "" {
				continue
			}
			num := ep.Number
			if num <= 0 {
				num = 1
			}
			format := ep.Format
			if format == "" {
				format = constant.PlayFormatHLS
			}
			existingEp, err := pRepo.GetEpisodeByKey(ctx, videoID, source.PlaySourceID, num)
			if err != nil {
				return err
			}
			if existingEp != nil {
				existingEp.Title = ep.Title
				existingEp.PlayURL = ep.URL
				existingEp.Quality = ep.Quality
				existingEp.Format = format
				existingEp.Status = constant.StatusEnabled
				if err := pRepo.UpdateEpisode(ctx, existingEp); err != nil {
					return err
				}
				continue
			}
			m := &model.PlayEpisodes{
				SourceID:      source.PlaySourceID,
				VideoID:       videoID,
				EpisodeNumber: uint32(num),
				Title:         ep.Title,
				PlayURL:       ep.URL,
				Quality:       ep.Quality,
				Format:        format,
				Status:        constant.StatusEnabled,
			}
			if err := pRepo.CreateEpisode(ctx, m); err != nil {
				return err
			}
		}
		return nil
	})
}

// resolveCategoryFields determines the video's category_id and parent_category_id
// based on the bound system category. If the bound category is a top-level category
// (ParentID==0, 一级分类), it is written to parent_category_id and category_id is left
// empty. If it is a second-level category (二级分类), it is written to category_id and
// its parent is written to parent_category_id.
func resolveCategoryFields(boundCategoryID uint32, parentMap map[uint32]uint32) (categoryID, parentCategoryID uint32) {
	parentID := parentMap[boundCategoryID]
	if parentID == 0 {
		// 一级分类：写入 parent_category_id
		return 0, boundCategoryID
	}
	// 二级分类：写入 category_id，父分类写入 parent_category_id
	return boundCategoryID, parentID
}

// applySupplementFields fills empty/zero basic fields of an existing video with values
// from the collected item (supplement semantics: never overwrite non-empty values).
// When updateCover is true (same-source), the cover image is overridden with the new
// value to capture remote domain/path changes; when false (cross-source) the cover is
// left untouched. PublishStatus is never modified here — it is a system-managed field.
// catID/parentCatID are the pre-resolved target category fields (see resolveCategoryFields).
func applySupplementFields(v *model.Videos, item Item, catID, parentCatID uint32, serialStatus uint8, updateCover bool) {
	if v.Subtitle == "" && item.Subtitle != "" {
		v.Subtitle = item.Subtitle
	}
	if v.Description == "" && item.Description != "" {
		v.Description = item.Description
	}
	if updateCover && item.Cover != "" {
		v.CoverImage = item.Cover
	}
	if v.Year == 0 && item.Year > 0 {
		v.Year = uint32(item.Year)
	}
	if v.Region == "" && item.Region != "" {
		v.Region = item.Region
	}
	if v.Language == "" && item.Language != "" {
		v.Language = item.Language
	}
	if v.Duration == 0 && item.Duration > 0 {
		v.Duration = uint32(item.Duration)
	}
	if v.ReleaseDate == "" {
		if rd := strings.TrimSpace(item.ReleaseDate); rd != "" {
			v.ReleaseDate = rd
		}
	}
	if v.SerialStatus == 0 && serialStatus > 0 {
		v.SerialStatus = serialStatus
	}
	if v.CategoryID == 0 {
		v.CategoryID = catID
	}
	if v.ParentCategoryID == 0 {
		v.ParentCategoryID = parentCatID
	}
}

// migrateDomainIfChanged detects remote domain changes between stored data and newly
// collected data, then batch-migrates all historical records of the same source.
// Cover image domain and play URL domain are detected independently. This enables
// updating all old data to a new domain by collecting just one item.
func (e *Engine) migrateDomainIfChanged(ctx context.Context, source *model.CollectSources, existing *model.Videos, item Item) error {
	// Cover image domain migration
	if existing.CoverImage != "" && item.Cover != "" {
		oldHost := extractHost(existing.CoverImage)
		newHost := extractHost(item.Cover)
		if oldHost != "" && newHost != "" && oldHost != newHost {
			n, err := e.videoRepo.UpdateCoverDomainByCollectSource(ctx, source.ID, oldHost, newHost)
			if err != nil {
				return fmt.Errorf("migrate cover domain: %w", err)
			}
			e.log.Info("collect: migrated cover domain",
				zap.Uint32("source_id", source.ID),
				zap.String("old_host", oldHost),
				zap.String("new_host", newHost),
				zap.Int("affected", n),
			)
		}
	}

	// Play URL domain migration (independent detection)
	if len(item.Episodes) == 0 || source.PlaySourceID == 0 {
		return nil
	}
	var newPlayURL string
	for _, ep := range item.Episodes {
		if strings.TrimSpace(ep.URL) != "" {
			newPlayURL = ep.URL
			break
		}
	}
	if newPlayURL == "" {
		return nil
	}
	newHost := extractHost(newPlayURL)
	if newHost == "" {
		return nil
	}
	eps, _, err := e.playRepo.ListEpisodes(ctx, existing.ID, source.PlaySourceID, 0, 1)
	if err != nil {
		return fmt.Errorf("migrate play url domain: list episodes: %w", err)
	}
	if len(eps) == 0 {
		return nil
	}
	oldHost := extractHost(eps[0].PlayURL)
	if oldHost == "" || oldHost == newHost {
		return nil
	}
	n, err := e.playRepo.UpdatePlayURLDomainBySource(ctx, source.PlaySourceID, oldHost, newHost)
	if err != nil {
		return fmt.Errorf("migrate play url domain: %w", err)
	}
	e.log.Info("collect: migrated play url domain",
		zap.Uint32("play_source_id", source.PlaySourceID),
		zap.String("old_host", oldHost),
		zap.String("new_host", newHost),
		zap.Int("affected", n),
	)
	return nil
}

func (e *Engine) ensureDirectors(ctx context.Context, names []string) ([]uint32, error) {
	ids := make([]uint32, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		// list + create pattern via metadata repo methods
		item, err := e.metaRepo.GetDirectorByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if item == nil {
			item = &model.Directors{Name: name}
			if err := e.metaRepo.CreateDirector(ctx, item); err != nil {
				// race: re-get
				got, getErr := e.metaRepo.GetDirectorByName(ctx, name)
				if getErr != nil {
					return nil, getErr
				}
				if got == nil {
					return nil, err
				}
				item = got
			}
		}
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (e *Engine) ensureActors(ctx context.Context, names []string) ([]uint32, error) {
	ids := make([]uint32, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		item, err := e.metaRepo.GetActorByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if item == nil {
			item = &model.Actors{Name: name}
			if err := e.metaRepo.CreateActor(ctx, item); err != nil {
				got, getErr := e.metaRepo.GetActorByName(ctx, name)
				if getErr != nil {
					return nil, getErr
				}
				if got == nil {
					return nil, err
				}
				item = got
			}
		}
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (e *Engine) ensureTags(ctx context.Context, names []string) ([]uint32, error) {
	ids := make([]uint32, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		item, err := e.metaRepo.GetTagByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if item == nil {
			item = &model.Tags{Name: name}
			if err := e.metaRepo.CreateTag(ctx, item); err != nil {
				got, getErr := e.metaRepo.GetTagByName(ctx, name)
				if getErr != nil {
					return nil, getErr
				}
				if got == nil {
					return nil, err
				}
				item = got
			}
		}
		ids = append(ids, item.ID)
	}
	return ids, nil
}

// parseSerialStatus parses vod_remarks to determine serial status.
// Common remarks: "完结", "已完结", "完结中" → Finished; "更新至", "连载中" → Ongoing; "预告" → Upcoming.
func parseSerialStatus(remarks string) uint8 {
	remarks = strings.TrimSpace(remarks)
	if remarks == "" {
		return 0
	}
	switch {
	case strings.Contains(remarks, "完结"):
		return constant.SerialStatusFinished
	case strings.Contains(remarks, "更新") || strings.Contains(remarks, "连载") || (strings.Contains(remarks, "第") && strings.Contains(remarks, "集")):
		return constant.SerialStatusOngoing
	case strings.Contains(remarks, "预告") || strings.Contains(remarks, "即将"):
		return constant.SerialStatusUpcoming
	default:
		return 0
	}
}

// upsertEpisodes only adds/updates play episodes for a video collected by a different source.
// It does not modify the video record itself.
func (e *Engine) upsertEpisodes(ctx context.Context, pRepo repository.PlayRepository, source *model.CollectSources, videoID uint32, item Item) error {
	for _, ep := range item.Episodes {
		if ep.URL == "" {
			continue
		}
		num := ep.Number
		if num <= 0 {
			num = 1
		}
		format := ep.Format
		if format == "" {
			format = constant.PlayFormatHLS
		}
		existingEp, err := pRepo.GetEpisodeByKey(ctx, videoID, source.PlaySourceID, num)
		if err != nil {
			return err
		}
		if existingEp != nil {
			existingEp.Title = ep.Title
			existingEp.PlayURL = ep.URL
			existingEp.Quality = ep.Quality
			existingEp.Format = format
			existingEp.Status = constant.StatusEnabled
			if err := pRepo.UpdateEpisode(ctx, existingEp); err != nil {
				return err
			}
			continue
		}
		m := &model.PlayEpisodes{
			SourceID:      source.PlaySourceID,
			VideoID:       videoID,
			EpisodeNumber: uint32(num),
			Title:         ep.Title,
			PlayURL:       ep.URL,
			Quality:       ep.Quality,
			Format:        format,
			Status:        constant.StatusEnabled,
		}
		if err := pRepo.CreateEpisode(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

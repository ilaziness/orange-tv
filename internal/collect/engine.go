package collect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
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
func (e *Engine) Run(ctx context.Context, source *model.CollectSources, dataRange string, logID uint64) Result {
	res := Result{}

	if source.Type != constant.CollectTypeAppleCMS {
		res.HasError = true
		res.Message = "仅支持苹果CMS采集源"
		return res
	}

	maps, err := e.collectRepo.ListCategories(ctx, int64(source.ID))
	if err != nil {
		e.log.Error("collect: list categories failed", zap.Int64("source_id", int64(source.ID)), zap.Error(err))
		res.HasError = true
		res.Message = err.Error()
		return res
	}
	catMap := map[int64]int64{}
	for _, m := range maps {
		if m.ExternalCategoryID == 0 {
			continue
		}
		catMap[int64(m.ExternalCategoryID)] = int64(m.CategoryID)
	}

	cats, err := e.categoryRepo.List(ctx, false)
	if err != nil {
		e.log.Error("collect: list all categories for parent map failed", zap.Int64("source_id", int64(source.ID)), zap.Error(err))
		res.HasError = true
		res.Message = err.Error()
		return res
	}
	parentMap := map[uint64]uint64{}
	for _, c := range cats {
		parentMap[c.ID] = c.ParentID
	}

	cutoffTime := dataRangeCutoff(dataRange)

	// Phase 1: collect all vod_ids from list pages
	var allVodIDs []int64
	pageNo := 1
	maxPages := 50
	for pageNo <= maxPages {
		if err := ctx.Err(); err != nil {
			res.Message = "采集已取消"
			return res
		}
		body, err := e.fetcher.FetchList(ctx, source.CollectURL, source.APIKey, pageNo, dataRange)
		if err != nil {
			e.log.Error("collect: fetch list failed", zap.Int64("source_id", int64(source.ID)), zap.Int("page", pageNo), zap.Error(err))
			res.HasError = true
			res.Message = fmt.Sprintf("拉取列表第%d页失败: %v", pageNo, err)
			return res
		}
		listPage, err := ParseAppleCMSList(body)
		if err != nil {
			e.log.Error("collect: parse list failed", zap.Int64("source_id", int64(source.ID)), zap.Int("page", pageNo), zap.Error(err))
			res.HasError = true
			res.Message = fmt.Sprintf("解析列表第%d页失败: %v", pageNo, err)
			return res
		}
		if len(listPage.VodIDs) == 0 {
			break
		}

		// check vod_time for time range filtering: stop if the last item is before cutoff
		if cutoffTime != nil && len(listPage.VodTimes) > 0 {
			lastTime := listPage.VodTimes[len(listPage.VodTimes)-1]
			if lastTime != "" && isVodTimeBefore(lastTime, cutoffTime) {
				// still add IDs whose vod_time is within range
				for i, t := range listPage.VodTimes {
					if t != "" && isVodTimeBefore(t, cutoffTime) {
						break
					}
					allVodIDs = append(allVodIDs, listPage.VodIDs[i])
				}
				break
			}
		}

		allVodIDs = append(allVodIDs, listPage.VodIDs...)

		if pageNo >= listPage.PageCount || pageNo >= maxPages {
			break
		}
		pageNo++
	}

	if len(allVodIDs) == 0 {
		return res
	}

	// Phase 2: fetch details in batches of 25 and process
	batchSize := 25
	for i := 0; i < len(allVodIDs); i += batchSize {
		if err := ctx.Err(); err != nil {
			res.Message = "采集已取消"
			return res
		}
		end := i + batchSize
		if end > len(allVodIDs) {
			end = len(allVodIDs)
		}
		batch := allVodIDs[i:end]

		body, err := e.fetcher.FetchDetail(ctx, source.CollectURL, source.APIKey, batch)
		if err != nil {
			e.log.Error("collect: fetch detail failed", zap.Int64("source_id", int64(source.ID)), zap.Int("batch_start", i), zap.Int("batch_end", end), zap.Error(err))
			res.HasError = true
			res.Message = fmt.Sprintf("拉取详情失败(batch %d-%d): %v", i, end, err)
			return res
		}
		page, err := ParseAppleCMSDetail(body)
		if err != nil {
			e.log.Error("collect: parse detail failed", zap.Int64("source_id", int64(source.ID)), zap.Int("batch_start", i), zap.Int("batch_end", end), zap.Error(err))
			res.HasError = true
			res.Message = fmt.Sprintf("解析详情失败(batch %d-%d): %v", i, end, err)
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

			// vod_time filtering: if item time is before cutoff, stop all subsequent processing
			if cutoffTime != nil && item.VodTime != "" {
				if isVodTimeBefore(item.VodTime, cutoffTime) {
					stopCollection = true
					break
				}
			}

			if err := e.upsertItem(ctx, source, catMap, parentMap, item); err != nil {
				e.log.Warn("collect item failed",
					zap.Int64("source_id", int64(source.ID)),
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

// dataRangeCutoff returns the earliest time for the given data range.
// Returns nil for "all" or empty (no filtering).
func dataRangeCutoff(dataRange string) *time.Time {
	now := time.Now()
	switch strings.TrimSpace(dataRange) {
	case "today":
		t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return &t
	case "last1d":
		t := now.AddDate(0, 0, -1)
		return &t
	case "last3d":
		t := now.AddDate(0, 0, -3)
		return &t
	case "last1w":
		t := now.AddDate(0, 0, -7)
		return &t
	case "last1m":
		t := now.AddDate(0, -1, 0)
		return &t
	default:
		return nil
	}
}

// isVodTimeBefore checks if vod_time (format "2006-01-02 15:04:05") is before cutoff.
func isVodTimeBefore(vodTime string, cutoff *time.Time) bool {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(vodTime), time.Local)
	if err != nil {
		// try date-only format
		t, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(vodTime), time.Local)
		if err != nil {
			return false
		}
	}
	return t.Before(*cutoff)
}

func (e *Engine) upsertItem(ctx context.Context, source *model.CollectSources, catMap map[int64]int64, parentMap map[uint64]uint64, item Item) error {
	if item.ExternalCategoryID <= 0 {
		return fmt.Errorf("外部分类ID无效")
	}
	categoryID, ok := catMap[item.ExternalCategoryID]
	if !ok {
		return fmt.Errorf("未映射分类: %d", item.ExternalCategoryID)
	}
	if item.Title == "" {
		return fmt.Errorf("标题为空")
	}

	existing, err := e.collectRepo.FindVideoByTitleYear(ctx, item.Title, item.Year)
	if err != nil {
		return err
	}

	// If video was collected by a different source, only add play episodes (no metadata creation)
	if existing != nil && existing.CollectSourceID != 0 && existing.CollectSourceID != uint64(source.ID) {
		return e.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
			pRepo := e.playRepo.WithTx(tx)
			return e.upsertEpisodes(ctx, pRepo, source, existing.ID, item)
		})
	}

	// ensure metadata names (only for same-source updates or new videos)
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

	serialStatus := parseSerialStatus(item.Remarks)

	return e.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		vRepo := e.videoRepo.WithTx(tx)
		pRepo := e.playRepo.WithTx(tx)
		var videoID uint64
		if existing != nil {
			videoID = existing.ID
			// Same source or manually created (CollectSourceID==0): claim it
			if existing.CollectSourceID == 0 {
				existing.CollectSourceID = uint64(source.ID)
			}
			// Same source: update video fields
			existing.Subtitle = item.Subtitle
			if item.Description != "" {
				existing.Description = &item.Description
			}
			if item.Cover != "" {
				existing.CoverImage = item.Cover
			}
			if item.Year > 0 {
				existing.Year = uint32(item.Year)
			}
			if item.Region != "" {
				existing.Region = item.Region
			}
			if item.Language != "" {
				existing.Language = item.Language
			}
			if item.Duration > 0 {
				existing.Duration = uint32(item.Duration)
			}
			if rd := releaseDatePtr(item.ReleaseDate); rd != nil {
				existing.ReleaseDate = rd
			}
			if serialStatus > 0 {
				existing.SerialStatus = serialStatus
			}
			existing.CategoryID = uint64(categoryID)
			existing.ParentCategoryID = parentMap[uint64(categoryID)]
			if existing.PublishStatus == 0 {
				existing.PublishStatus = constant.PublishStatusOnline
			}
			if err := vRepo.Update(ctx, existing); err != nil {
				return err
			}
		} else {
			desc := item.Description
			var descPtr *string
			if desc != "" {
				descPtr = &desc
			}
			v := &model.Videos{
				Title:            item.Title,
				Subtitle:         item.Subtitle,
				Description:      descPtr,
				CategoryID:       uint64(categoryID),
				ParentCategoryID: parentMap[uint64(categoryID)],
				PublishStatus:    constant.PublishStatusOnline,
				SerialStatus:     serialStatus,
				CoverImage:       item.Cover,
				PosterImage:      "",
				Year:             uint32(item.Year),
				Region:           item.Region,
				Language:         item.Language,
				Duration:         uint32(item.Duration),
				ReleaseDate:      releaseDatePtr(item.ReleaseDate),
				CollectSourceID:  uint64(source.ID),
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
			existingEp, err := pRepo.GetEpisodeByKey(ctx, int64(videoID), int64(source.PlaySourceID), num)
			if err != nil {
				return err
			}
			if existingEp != nil {
				existingEp.Title = ep.Title
				existingEp.PlayURL = ep.URL
				existingEp.Quality = ep.Quality
				existingEp.Format = format
				existingEp.Status = constant.StatusEnabled
				if existingEp.DeletedAt != nil {
					if err := pRepo.RestoreAndUpdateEpisode(ctx, existingEp); err != nil {
						return err
					}
				} else {
					if err := pRepo.UpdateEpisode(ctx, existingEp); err != nil {
						return err
					}
				}
				continue
			}
			m := &model.PlayEpisodes{
				SourceID:      uint64(source.PlaySourceID),
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

func (e *Engine) ensureDirectors(ctx context.Context, names []string) ([]uint64, error) {
	ids := make([]uint64, 0, len(names))
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

func (e *Engine) ensureActors(ctx context.Context, names []string) ([]uint64, error) {
	ids := make([]uint64, 0, len(names))
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

func (e *Engine) ensureTags(ctx context.Context, names []string) ([]uint64, error) {
	ids := make([]uint64, 0, len(names))
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

// releaseDatePtr trims the raw collect release_date string and returns a pointer
// (nil when empty). The value is stored as-is to preserve the original source format.
func releaseDatePtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
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
func (e *Engine) upsertEpisodes(ctx context.Context, pRepo repository.PlayRepository, source *model.CollectSources, videoID uint64, item Item) error {
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
		existingEp, err := pRepo.GetEpisodeByKey(ctx, int64(videoID), int64(source.PlaySourceID), num)
		if err != nil {
			return err
		}
		if existingEp != nil {
			existingEp.Title = ep.Title
			existingEp.PlayURL = ep.URL
			existingEp.Quality = ep.Quality
			existingEp.Format = format
			existingEp.Status = constant.StatusEnabled
			if existingEp.DeletedAt != nil {
				if err := pRepo.RestoreAndUpdateEpisode(ctx, existingEp); err != nil {
					return err
				}
			} else {
				if err := pRepo.UpdateEpisode(ctx, existingEp); err != nil {
					return err
				}
			}
			continue
		}
		m := &model.PlayEpisodes{
			SourceID:      uint64(source.PlaySourceID),
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

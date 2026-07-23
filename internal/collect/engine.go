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
	return &Engine{
		collectRepo:  collectRepo,
		videoRepo:    videoRepo,
		categoryRepo: categoryRepo,
		metaRepo:     metaRepo,
		playRepo:     playRepo,
		fetcher:      NewFetcher(),
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

	maps, err := e.collectRepo.ListCategories(ctx, int64(source.ID))
	if err != nil {
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

	isApple := source.Type == uint8(constant.CollectTypeAppleCMS)
	pageNo := 1
	maxPages := 50
	cutoffTime := dataRangeCutoff(dataRange)

	for pageNo <= maxPages {
		if err := ctx.Err(); err != nil {
			res.Message = "采集已取消"
			break
		}
		body, err := e.fetcher.FetchPage(ctx, source.CollectURL, source.APIKey, pageNo, isApple, dataRange)
		if err != nil {
			res.HasError = true
			res.Message = fmt.Sprintf("拉取第%d页失败: %v", pageNo, err)
			break
		}
		var page *Page
		if isApple {
			page, err = ParseAppleCMS(body)
		} else {
			page, err = ParseDefaultJSON(body)
		}
		if err != nil {
			res.HasError = true
			res.Message = fmt.Sprintf("解析第%d页失败: %v", pageNo, err)
			break
		}
		if len(page.Items) == 0 {
			break
		}

		pageCollected := 0
		stopCollection := false
		for _, item := range page.Items {
			if err := ctx.Err(); err != nil {
				res.Message = "采集已取消"
				stopCollection = true
				break
			}

			// vod_time filtering: if item time is before cutoff, stop all subsequent pages
			if cutoffTime != nil && item.VodTime != "" {
				if isVodTimeBefore(item.VodTime, cutoffTime) {
					stopCollection = true
					break
				}
			}

			if err := e.upsertItem(ctx, source, catMap, item); err != nil {
				if e.log != nil {
					e.log.Warn("collect item failed",
						zap.Int64("source_id", int64(source.ID)),
						zap.String("title", item.Title),
						zap.Error(err),
					)
				}
				continue
			}
			pageCollected++
		}

		// increment log count for this page
		if logID > 0 && pageCollected > 0 {
			if err := e.collectRepo.IncrementLogCount(ctx, logID, pageCollected); err != nil && e.log != nil {
				e.log.Warn("increment log count failed", zap.Error(err))
			}
		}
		res.CollectCount += pageCollected

		if stopCollection {
			break
		}
		if pageNo >= page.PageCount || pageNo >= maxPages {
			break
		}
		pageNo++
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

func (e *Engine) upsertItem(ctx context.Context, source *model.CollectSources, catMap map[int64]int64, item Item) error {
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

	// ensure metadata names
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

	return e.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		vRepo := e.videoRepo.WithTx(tx)
		pRepo := e.playRepo.WithTx(tx)
		var videoID uint64
		if existing != nil {
			videoID = existing.ID
			existing.Subtitle = item.Subtitle
			if item.Description != "" {
				existing.Description = &item.Description
			}
			if item.Cover != "" {
				existing.CoverImage = item.Cover
			}
			if item.Poster != "" {
				existing.PosterImage = item.Poster
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
			if item.Rating > 0 {
				existing.Rating = item.Rating
			}
			if rd := parseReleaseDate(item.ReleaseDate); rd != nil {
				existing.ReleaseDate = rd
			}
			existing.CategoryID = uint64(categoryID)
			if existing.PublishStatus == 0 {
				existing.PublishStatus = uint8(constant.PublishStatusOnline)
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
				Title:         item.Title,
				Subtitle:      item.Subtitle,
				Description:   descPtr,
				CategoryID:    uint64(categoryID),
				PublishStatus: uint8(constant.PublishStatusOnline),
				SerialStatus:  uint8(constant.SerialStatusFinished),
				CoverImage:    item.Cover,
				PosterImage:   firstNonEmpty(item.Poster, item.Cover),
				Year:          uint32(item.Year),
				Region:        item.Region,
				Language:      item.Language,
				Duration:      uint32(item.Duration),
				Rating:        item.Rating,
				ReleaseDate:   parseReleaseDate(item.ReleaseDate),
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
				existingEp.Status = uint8(constant.StatusEnabled)
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
				Status:        uint8(constant.StatusEnabled),
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

func parseReleaseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", "2006-1-2", time.RFC3339} {
		if tm, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &tm
		}
	}
	return nil
}

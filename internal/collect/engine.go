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
	Total   int
	Success int
	Failed  int
	Message string
}

// Run executes collection for a source (all pages until page_count).
func (e *Engine) Run(ctx context.Context, source *model.CollectSources) Result {
	start := time.Now()
	res := Result{}
	maps, err := e.collectRepo.ListCategories(ctx, source.ID)
	if err != nil {
		res.Message = err.Error()
		return res
	}
	catMap := map[string]int64{}
	for _, m := range maps {
		catMap[strings.TrimSpace(m.ExternalCategory)] = m.CategoryID
	}
	if len(catMap) == 0 {
		res.Message = "请先配置分类映射"
		return res
	}
	if source.PlaySourceID <= 0 {
		res.Message = "采集源未绑定播放源"
		return res
	}
	playSrc, err := e.playRepo.GetSource(ctx, source.PlaySourceID)
	if err != nil || playSrc == nil {
		res.Message = "绑定的播放源不存在"
		return res
	}

	isApple := source.Type == constant.CollectTypeAppleCMS
	pageNo := 1
	maxPages := 50 // safety cap per run
	cancelled := false
	for pageNo <= maxPages {
		if err := ctx.Err(); err != nil {
			res.Message = "采集已取消"
			break
		}
		body, err := e.fetcher.FetchPage(ctx, source.CollectURL, source.APIKey, pageNo, isApple)
		if err != nil {
			res.Failed++
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
			res.Failed++
			res.Message = fmt.Sprintf("解析第%d页失败: %v", pageNo, err)
			break
		}
		if len(page.Items) == 0 {
			break
		}
		for _, item := range page.Items {
			if err := ctx.Err(); err != nil {
				res.Message = "采集已取消"
				cancelled = true
				break
			}
			res.Total++
			if err := e.upsertItem(ctx, source, catMap, item); err != nil {
				res.Failed++
				if e.log != nil {
					e.log.Warn("collect item failed",
						zap.Int64("source_id", source.ID),
						zap.String("title", item.Title),
						zap.Error(err),
					)
				}
				continue
			}
			res.Success++
		}
		if cancelled {
			break
		}
		if pageNo >= page.PageCount || pageNo >= maxPages {
			break
		}
		pageNo++
	}
	_ = start
	return res
}

func (e *Engine) upsertItem(ctx context.Context, source *model.CollectSources, catMap map[string]int64, item Item) error {
	categoryID, ok := catMap[item.CategoryKey]
	if !ok {
		// try numeric/name fallback already in map keys only
		return fmt.Errorf("未映射分类: %s", item.CategoryKey)
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
		var videoID int64
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
				existing.Year = item.Year
			}
			if item.Region != "" {
				existing.Region = item.Region
			}
			if item.Language != "" {
				existing.Language = item.Language
			}
			if item.Duration > 0 {
				existing.Duration = item.Duration
			}
			if item.Rating > 0 {
				existing.Rating = item.Rating
			}
			if rd := parseReleaseDate(item.ReleaseDate); rd != nil {
				existing.ReleaseDate = rd
			}
			existing.CategoryID = categoryID
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
				Title:         item.Title,
				Subtitle:      item.Subtitle,
				Description:   descPtr,
				CategoryID:    categoryID,
				PublishStatus: constant.PublishStatusOnline,
				SerialStatus:  constant.SerialStatusFinished,
				CoverImage:    item.Cover,
				PosterImage:   firstNonEmpty(item.Poster, item.Cover),
				Year:          item.Year,
				Region:        item.Region,
				Language:      item.Language,
				Duration:      item.Duration,
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
				SourceID:      source.PlaySourceID,
				VideoID:       videoID,
				EpisodeNumber: num,
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

func (e *Engine) ensureDirectors(ctx context.Context, names []string) ([]int64, error) {
	ids := make([]int64, 0, len(names))
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

func (e *Engine) ensureActors(ctx context.Context, names []string) ([]int64, error) {
	ids := make([]int64, 0, len(names))
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

func (e *Engine) ensureTags(ctx context.Context, names []string) ([]int64, error) {
	ids := make([]int64, 0, len(names))
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

package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/uptrace/bun"
)

// VideoService manages videos and associations.
type VideoService interface {
	List(ctx context.Context, req *dto.VideoListRequest) ([]shareddto.VideoListItem, int, error)
	Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error)
	Create(ctx context.Context, req *dto.CreateVideoRequest) (*shareddto.VideoDetailResponse, error)
	Update(ctx context.Context, id int64, req *dto.UpdateVideoRequest) (*shareddto.VideoDetailResponse, error)
	Delete(ctx context.Context, id int64) error
}

type videoService struct {
	videoRepo    repository.VideoRepository
	categoryRepo repository.CategoryRepository
	metaRepo     repository.MetadataRepository
	playRepo     repository.PlayRepository
	cache        cache.Cache
}

// NewVideoService creates a VideoService.
func NewVideoService(
	videoRepo repository.VideoRepository,
	categoryRepo repository.CategoryRepository,
	metaRepo repository.MetadataRepository,
	playRepo repository.PlayRepository,
	c cache.Cache,
) VideoService {
	if c == nil {
		c = cache.NewNopCache()
	}
	return &videoService{
		videoRepo:    videoRepo,
		categoryRepo: categoryRepo,
		metaRepo:     metaRepo,
		playRepo:     playRepo,
		cache:        c,
	}
}

func (s *videoService) List(ctx context.Context, req *dto.VideoListRequest) ([]shareddto.VideoListItem, int, error) {
	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		Keyword:       strings.TrimSpace(req.Keyword),
		CategoryID:    req.CategoryID,
		PublishStatus: req.PublishStatus,
		Year:          req.Year,
		Region:        strings.TrimSpace(req.Region),
		Language:      strings.TrimSpace(req.Language),
		Sort:          req.Sort,
		Offset:        req.GetOffset(),
		Limit:         req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapVideoList(items), total, nil
}

func (s *videoService) Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error) {
	return s.getDetail(ctx, id, false)
}

func (s *videoService) Create(ctx context.Context, req *dto.CreateVideoRequest) (*shareddto.VideoDetailResponse, error) {
	if err := s.ensureCategory(ctx, req.CategoryID); err != nil {
		return nil, err
	}
	if err := s.ensureDirectors(ctx, req.DirectorIDs); err != nil {
		return nil, err
	}
	if err := s.ensureActors(ctx, req.Actors); err != nil {
		return nil, err
	}
	if err := s.ensureTags(ctx, req.TagIDs); err != nil {
		return nil, err
	}

	publish := constant.PublishStatusOffline
	if req.PublishStatus != nil {
		publish = *req.PublishStatus
	}
	serial := constant.SerialStatusOngoing
	if req.SerialStatus != nil {
		serial = *req.SerialStatus
	}
	var desc *string
	if strings.TrimSpace(req.Description) != "" {
		d := strings.TrimSpace(req.Description)
		desc = &d
	}
	releaseDate, err := parseOptionalDate(req.ReleaseDate)
	if err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "上映日期格式无效，期望 YYYY-MM-DD")
	}

	video := &model.Videos{
		Title:         strings.TrimSpace(req.Title),
		Subtitle:      strings.TrimSpace(req.Subtitle),
		Description:   desc,
		CategoryID:    req.CategoryID,
		PublishStatus: publish,
		SerialStatus:  serial,
		CoverImage:    strings.TrimSpace(req.CoverImage),
		PosterImage:   strings.TrimSpace(req.PosterImage),
		Year:          req.Year,
		Region:        strings.TrimSpace(req.Region),
		Rating:        req.Rating,
		Duration:      req.Duration,
		Language:      strings.TrimSpace(req.Language),
		ReleaseDate:   releaseDate,
	}

	err = s.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := s.videoRepo.WithTx(tx)
		if err := txRepo.Create(ctx, video); err != nil {
			return err
		}
		if err := txRepo.ReplaceDirectors(ctx, video.ID, uniqueIDs(req.DirectorIDs)); err != nil {
			return err
		}
		if err := txRepo.ReplaceActors(ctx, video.ID, toActorRels(req.Actors)); err != nil {
			return err
		}
		return txRepo.ReplaceTags(ctx, video.ID, uniqueIDs(req.TagIDs))
	})
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.invalidateListCaches(ctx, video.ID)
	return s.getDetail(ctx, video.ID, false)
}

func (s *videoService) Update(ctx context.Context, id int64, req *dto.UpdateVideoRequest) (*shareddto.VideoDetailResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil {
		return nil, errcode.VideoNotFound
	}
	if req.CategoryID != nil {
		if err := s.ensureCategory(ctx, *req.CategoryID); err != nil {
			return nil, err
		}
		video.CategoryID = *req.CategoryID
	}
	if req.DirectorIDs != nil {
		if err := s.ensureDirectors(ctx, *req.DirectorIDs); err != nil {
			return nil, err
		}
	}
	if req.Actors != nil {
		if err := s.ensureActors(ctx, *req.Actors); err != nil {
			return nil, err
		}
	}
	if req.TagIDs != nil {
		if err := s.ensureTags(ctx, *req.TagIDs); err != nil {
			return nil, err
		}
	}

	if req.Title != nil {
		video.Title = strings.TrimSpace(*req.Title)
	}
	if req.Subtitle != nil {
		video.Subtitle = strings.TrimSpace(*req.Subtitle)
	}
	if req.Description != nil {
		d := strings.TrimSpace(*req.Description)
		video.Description = &d
	}
	if req.PublishStatus != nil {
		video.PublishStatus = *req.PublishStatus
	}
	if req.SerialStatus != nil {
		video.SerialStatus = *req.SerialStatus
	}
	if req.CoverImage != nil {
		video.CoverImage = strings.TrimSpace(*req.CoverImage)
	}
	if req.PosterImage != nil {
		video.PosterImage = strings.TrimSpace(*req.PosterImage)
	}
	if req.Year != nil {
		video.Year = *req.Year
	}
	if req.Region != nil {
		video.Region = strings.TrimSpace(*req.Region)
	}
	if req.Rating != nil {
		video.Rating = *req.Rating
	}
	if req.Duration != nil {
		video.Duration = *req.Duration
	}
	if req.Language != nil {
		video.Language = strings.TrimSpace(*req.Language)
	}
	if req.ReleaseDate != nil {
		rd, err := parseOptionalDate(*req.ReleaseDate)
		if err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "上映日期格式无效，期望 YYYY-MM-DD")
		}
		video.ReleaseDate = rd
	}

	err = s.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := s.videoRepo.WithTx(tx)
		if err := txRepo.Update(ctx, video); err != nil {
			return err
		}
		if req.DirectorIDs != nil {
			if err := txRepo.ReplaceDirectors(ctx, id, uniqueIDs(*req.DirectorIDs)); err != nil {
				return err
			}
		}
		if req.Actors != nil {
			if err := txRepo.ReplaceActors(ctx, id, toActorRels(*req.Actors)); err != nil {
				return err
			}
		}
		if req.TagIDs != nil {
			if err := txRepo.ReplaceTags(ctx, id, uniqueIDs(*req.TagIDs)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.invalidateListCaches(ctx, id)
	return s.getDetail(ctx, id, false)
}

func (s *videoService) Delete(ctx context.Context, id int64) error {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil {
		return errcode.VideoNotFound
	}
	// Clear association rows so directors/actors/tags are not blocked by soft-deleted videos.
	err = s.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := s.videoRepo.WithTx(tx)
		if err := txRepo.ReplaceDirectors(ctx, id, nil); err != nil {
			return err
		}
		if err := txRepo.ReplaceActors(ctx, id, nil); err != nil {
			return err
		}
		if err := txRepo.ReplaceTags(ctx, id, nil); err != nil {
			return err
		}
		return txRepo.SoftDelete(ctx, id)
	})
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.invalidateListCaches(ctx, id)
	return nil
}

func (s *videoService) invalidateListCaches(ctx context.Context, videoID int64) {
	// Best-effort: common homepage / default page keys + open list keys.
	// Full prefix delete is not available on all cache drivers.
	for _, sort := range []string{"", "id_desc", "rating_desc", "view_count_desc", "created_at_desc"} {
		for _, page := range []int{1, 2} {
			for _, limit := range []int{12, 20, 24} {
				_ = s.cache.Delete(ctx, fmt.Sprintf("video:list:0:%s:%d:%d", sort, page, limit))
				_ = s.cache.Delete(ctx, fmt.Sprintf("open:videos:list:default:%d:%d", page, limit))
				_ = s.cache.Delete(ctx, fmt.Sprintf("open:videos:list:apple_cms:%d:%d", page, limit))
			}
		}
	}
	_ = s.cache.Delete(ctx, "open:categories")
	if videoID > 0 {
		_ = s.cache.Delete(ctx, fmt.Sprintf("open:videos:detail:default:%d", videoID))
		_ = s.cache.Delete(ctx, fmt.Sprintf("open:videos:detail:apple_cms:%d", videoID))
	}
}

func (s *videoService) getDetail(ctx context.Context, id int64, clientOnly bool) (*shareddto.VideoDetailResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil {
		return nil, errcode.VideoNotFound
	}
	if clientOnly && video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}

	directorIDs, err := s.videoRepo.ListDirectorIDs(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	directors, err := s.metaRepo.GetDirectorsByIDs(ctx, directorIDs)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorRels, err := s.videoRepo.ListActorRels(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorIDs := make([]int64, 0, len(actorRels))
	for _, rel := range actorRels {
		actorIDs = append(actorIDs, rel.ActorID)
	}
	actors, err := s.metaRepo.GetActorsByIDs(ctx, actorIDs)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorName := map[int64]string{}
	for _, a := range actors {
		actorName[a.ID] = a.Name
	}
	tagIDs, err := s.videoRepo.ListTagIDs(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tags, err := s.metaRepo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	episodes, err := s.playRepo.ListEpisodesByVideo(ctx, id, clientOnly)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sources, err := s.playRepo.ListSources(ctx)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sourceMap := map[int64]model.PlaySources{}
	for _, src := range sources {
		if clientOnly && src.Status != constant.StatusEnabled {
			continue
		}
		sourceMap[src.ID] = src
	}

	groups := map[int64]*shareddto.VideoSourceGroup{}
	order := make([]int64, 0)
	for _, ep := range episodes {
		src, ok := sourceMap[ep.SourceID]
		if !ok {
			continue
		}
		g, ok := groups[ep.SourceID]
		if !ok {
			g = &shareddto.VideoSourceGroup{ID: src.ID, Name: src.Name, Episodes: []shareddto.VideoSourceEpisode{}}
			groups[ep.SourceID] = g
			order = append(order, ep.SourceID)
		}
		g.Episodes = append(g.Episodes, shareddto.VideoSourceEpisode{
			Episode: ep.EpisodeNumber,
			Title:   ep.Title,
			URL:     ep.PlayURL,
			Quality: ep.Quality,
			Format:  ep.Format,
		})
	}
	sourceGroups := make([]shareddto.VideoSourceGroup, 0, len(order))
	for _, sid := range order {
		sourceGroups = append(sourceGroups, *groups[sid])
	}

	dirItems := make([]shareddto.NamedItem, 0, len(directors))
	for _, d := range directors {
		dirItems = append(dirItems, shareddto.NamedItem{ID: d.ID, Name: d.Name})
	}
	actorItems := make([]shareddto.ActorItem, 0, len(actorRels))
	for _, rel := range actorRels {
		actorItems = append(actorItems, shareddto.ActorItem{
			ID:   rel.ActorID,
			Name: actorName[rel.ActorID],
			Role: rel.Role,
		})
	}
	tagItems := make([]shareddto.NamedItem, 0, len(tags))
	for _, t := range tags {
		tagItems = append(tagItems, shareddto.NamedItem{ID: t.ID, Name: t.Name})
	}

	desc := ""
	if video.Description != nil {
		desc = *video.Description
	}
	release := ""
	if video.ReleaseDate != nil {
		release = video.ReleaseDate.Format("2006-01-02")
	}

	resp := &shareddto.VideoDetailResponse{
		ID:           video.ID,
		Title:        video.Title,
		Subtitle:     video.Subtitle,
		Description:  desc,
		CategoryID:   video.CategoryID,
		SerialStatus: video.SerialStatus,
		Cover:        video.CoverImage,
		Poster:       video.PosterImage,
		Year:         video.Year,
		Region:       video.Region,
		Language:     video.Language,
		Duration:     video.Duration,
		ReleaseDate:  release,
		Rating:       video.Rating,
		ViewCount:    video.ViewCount,
		Directors:    dirItems,
		Actors:       actorItems,
		Tags:         tagItems,
		Sources:      sourceGroups,
	}
	if !clientOnly {
		resp.PublishStatus = video.PublishStatus
	}
	return resp, nil
}

func (s *videoService) ensureCategory(ctx context.Context, id int64) error {
	c, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.CategoryNotFound
	}
	return nil
}

func (s *videoService) ensureDirectors(ctx context.Context, ids []int64) error {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return nil
	}
	items, err := s.metaRepo.GetDirectorsByIDs(ctx, ids)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(items) != len(ids) {
		return errcode.DirectorNotFound
	}
	return nil
}

func (s *videoService) ensureActors(ctx context.Context, actors []dto.VideoActorInput) error {
	ids := make([]int64, 0, len(actors))
	seen := map[int64]struct{}{}
	for _, a := range actors {
		if _, ok := seen[a.ActorID]; ok {
			continue
		}
		seen[a.ActorID] = struct{}{}
		ids = append(ids, a.ActorID)
	}
	if len(ids) == 0 {
		return nil
	}
	items, err := s.metaRepo.GetActorsByIDs(ctx, ids)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(items) != len(ids) {
		return errcode.ActorNotFound
	}
	return nil
}

func (s *videoService) ensureTags(ctx context.Context, ids []int64) error {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return nil
	}
	items, err := s.metaRepo.GetTagsByIDs(ctx, ids)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(items) != len(ids) {
		return errcode.TagNotFound
	}
	return nil
}

func mapVideoList(items []model.Videos) []shareddto.VideoListItem {
	out := make([]shareddto.VideoListItem, 0, len(items))
	for _, v := range items {
		out = append(out, shareddto.VideoListItem{
			ID:            v.ID,
			Title:         v.Title,
			Subtitle:      v.Subtitle,
			Cover:         v.CoverImage,
			Poster:        v.PosterImage,
			Year:          v.Year,
			Region:        v.Region,
			Language:      v.Language,
			Rating:        v.Rating,
			CategoryID:    v.CategoryID,
			PublishStatus: v.PublishStatus,
			SerialStatus:  v.SerialStatus,
			Duration:      v.Duration,
			ViewCount:     v.ViewCount,
		})
	}
	return out
}

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func toActorRels(inputs []dto.VideoActorInput) []model.VideoActors {
	seen := map[int64]struct{}{}
	out := make([]model.VideoActors, 0, len(inputs))
	for _, in := range inputs {
		if in.ActorID <= 0 {
			continue
		}
		if _, ok := seen[in.ActorID]; ok {
			continue
		}
		seen[in.ActorID] = struct{}{}
		out = append(out, model.VideoActors{
			ActorID: in.ActorID,
			Role:    strings.TrimSpace(in.Role),
		})
	}
	return out
}

func parseOptionalDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

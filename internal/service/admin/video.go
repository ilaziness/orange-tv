package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// VideoService manages videos and associations.
type VideoService interface {
	List(ctx context.Context, req *dto.VideoListRequest) ([]dto.VideoListItem, int, error)
	Get(ctx context.Context, id uint32) (*dto.VideoDetailResponse, error)
	Create(ctx context.Context, req *dto.CreateVideoRequest) (*dto.VideoDetailResponse, error)
	Update(ctx context.Context, id uint32, req *dto.UpdateVideoRequest) (*dto.VideoDetailResponse, error)
	Delete(ctx context.Context, id uint32) error
}

type videoService struct {
	videoRepo    repository.VideoRepository
	categoryRepo repository.CategoryRepository
	metaRepo     repository.MetadataRepository
	playRepo     repository.PlayRepository
	cache        *cache.Manager
	log          *zap.Logger
}

// NewVideoService creates a VideoService.
func NewVideoService(
	videoRepo repository.VideoRepository,
	categoryRepo repository.CategoryRepository,
	metaRepo repository.MetadataRepository,
	playRepo repository.PlayRepository,
	c *cache.Manager,
	log *zap.Logger,
) VideoService {
	if log == nil {
		log = zap.NewNop()
	}
	return &videoService{
		videoRepo:    videoRepo,
		categoryRepo: categoryRepo,
		metaRepo:     metaRepo,
		playRepo:     playRepo,
		cache:        c,
		log:          log,
	}
}

func (s *videoService) List(ctx context.Context, req *dto.VideoListRequest) ([]dto.VideoListItem, int, error) {
	cats, err := s.categoryRepo.List(ctx, false)
	if err != nil {
		s.log.Error("video: list categories failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	catMap := make(map[uint32]*model.Categories, len(cats))
	for i := range cats {
		catMap[cats[i].ID] = &cats[i]
	}

	filter := repository.VideoListFilter{
		Keyword:       strings.TrimSpace(req.Keyword),
		PublishStatus: req.PublishStatus,
		Year:          req.Year,
		Region:        strings.TrimSpace(req.Region),
		Language:      strings.TrimSpace(req.Language),
		Sort:          "id_desc",
		DirectorID:    req.DirectorID,
		ActorID:       req.ActorID,
		TagID:         req.TagID,
		Offset:        req.GetOffset(),
		Limit:         req.GetLimit(),
	}

	if req.CategoryID > 0 {
		filter.CategoryIDs = collectDescendantIDs(catMap, req.CategoryID)
	}

	items, total, err := s.videoRepo.List(ctx, filter)
	if err != nil {
		s.log.Error("video: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapVideoList(items, catMap), total, nil
}

func (s *videoService) Get(ctx context.Context, id uint32) (*dto.VideoDetailResponse, error) {
	return s.getDetail(ctx, id, false)
}

func (s *videoService) Create(ctx context.Context, req *dto.CreateVideoRequest) (*dto.VideoDetailResponse, error) {
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
	desc := strings.TrimSpace(req.Description)
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
		Duration:      req.Duration,
		Language:      strings.TrimSpace(req.Language),
		ReleaseDate:   strings.TrimSpace(req.ReleaseDate),
	}

	err := s.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := s.videoRepo.WithTx(tx)
		if err := txRepo.Create(ctx, video); err != nil {
			return err
		}
		if err := txRepo.ReplaceDirectors(ctx, video.ID, utils.UniqueUint32IDs(req.DirectorIDs)); err != nil {
			return err
		}
		if err := txRepo.ReplaceActors(ctx, video.ID, toActorRels(req.Actors)); err != nil {
			return err
		}
		return txRepo.ReplaceTags(ctx, video.ID, utils.UniqueUint32IDs(req.TagIDs))
	})
	if err != nil {
		s.log.Error("video: create transaction failed", zap.String("title", video.Title), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateVideo(ctx)
	return s.getDetail(ctx, video.ID, false)
}

func (s *videoService) Update(ctx context.Context, id uint32, req *dto.UpdateVideoRequest) (*dto.VideoDetailResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("video: get by id for update failed", zap.Uint32("video_id", id), zap.Error(err))
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
		video.Description = strings.TrimSpace(*req.Description)
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
	if req.Duration != nil {
		video.Duration = *req.Duration
	}
	if req.Language != nil {
		video.Language = strings.TrimSpace(*req.Language)
	}
	if req.ReleaseDate != nil {
		video.ReleaseDate = strings.TrimSpace(*req.ReleaseDate)
	}

	err = s.videoRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := s.videoRepo.WithTx(tx)
		if err := txRepo.Update(ctx, video); err != nil {
			return err
		}
		if req.DirectorIDs != nil {
			if err := txRepo.ReplaceDirectors(ctx, id, utils.UniqueUint32IDs(*req.DirectorIDs)); err != nil {
				return err
			}
		}
		if req.Actors != nil {
			if err := txRepo.ReplaceActors(ctx, id, toActorRels(*req.Actors)); err != nil {
				return err
			}
		}
		if req.TagIDs != nil {
			if err := txRepo.ReplaceTags(ctx, id, utils.UniqueUint32IDs(*req.TagIDs)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("video: update transaction failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateVideo(ctx)
	return s.getDetail(ctx, id, false)
}

func (s *videoService) Delete(ctx context.Context, id uint32) error {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("video: get by id for delete failed", zap.Uint32("video_id", id), zap.Error(err))
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
		s.log.Error("video: delete transaction failed", zap.Uint32("video_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateVideo(ctx)
	return nil
}

func (s *videoService) getDetail(ctx context.Context, id uint32, clientOnly bool) (*dto.VideoDetailResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("video: get by id for detail failed", zap.Uint32("video_id", id), zap.Error(err))
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
		s.log.Error("video: list director ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	directors, err := s.metaRepo.GetDirectorsByIDs(ctx, directorIDs)
	if err != nil {
		s.log.Error("video: get directors by ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorRels, err := s.videoRepo.ListActorRels(ctx, id)
	if err != nil {
		s.log.Error("video: list actor rels failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorIDs := make([]uint32, 0, len(actorRels))
	for _, rel := range actorRels {
		actorIDs = append(actorIDs, rel.ActorID)
	}
	actors, err := s.metaRepo.GetActorsByIDs(ctx, actorIDs)
	if err != nil {
		s.log.Error("video: get actors by ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorName := map[uint32]string{}
	for _, a := range actors {
		actorName[a.ID] = a.Name
	}
	tagIDs, err := s.videoRepo.ListTagIDs(ctx, id)
	if err != nil {
		s.log.Error("video: list tag ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tags, err := s.metaRepo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		s.log.Error("video: get tags by ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	episodes, err := s.playRepo.ListEpisodesByVideo(ctx, id, clientOnly)
	if err != nil {
		s.log.Error("video: list episodes by video failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sources, err := s.playRepo.ListSources(ctx)
	if err != nil {
		s.log.Error("video: list sources for detail failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sourceMap := map[uint32]model.PlaySources{}
	for _, src := range sources {
		if clientOnly && src.Status != constant.StatusEnabled {
			continue
		}
		sourceMap[src.ID] = src
	}

	groups := map[uint32]*dto.VideoSourceGroup{}
	order := make([]uint32, 0)
	for _, ep := range episodes {
		src, ok := sourceMap[ep.SourceID]
		if !ok {
			continue
		}
		g, ok := groups[ep.SourceID]
		if !ok {
			g = &dto.VideoSourceGroup{ID: src.ID, Name: src.Name, Episodes: []dto.VideoSourceEpisode{}}
			groups[ep.SourceID] = g
			order = append(order, ep.SourceID)
		}
		g.Episodes = append(g.Episodes, dto.VideoSourceEpisode{
			ID:      ep.ID,
			Episode: ep.EpisodeNumber,
			Title:   ep.Title,
			URL:     ep.PlayURL,
			Quality: ep.Quality,
			Format:  ep.Format,
			Status:  ep.Status,
		})
	}
	sourceGroups := make([]dto.VideoSourceGroup, 0, len(order))
	for _, sid := range order {
		sourceGroups = append(sourceGroups, *groups[sid])
	}

	dirItems := make([]shareddto.NamedItem, 0, len(directors))
	for _, d := range directors {
		dirItems = append(dirItems, shareddto.NamedItem{ID: d.ID, Name: d.Name})
	}
	actorItems := make([]shareddto.NamedItem, 0, len(actorRels))
	for _, rel := range actorRels {
		actorItems = append(actorItems, shareddto.NamedItem{
			ID:   rel.ActorID,
			Name: actorName[rel.ActorID],
		})
	}
	tagItems := make([]shareddto.NamedItem, 0, len(tags))
	for _, t := range tags {
		tagItems = append(tagItems, shareddto.NamedItem{ID: t.ID, Name: t.Name})
	}

	resp := &dto.VideoDetailResponse{
		ID:           video.ID,
		Title:        video.Title,
		Subtitle:     video.Subtitle,
		Description:  video.Description,
		CategoryID:   video.CategoryID,
		SerialStatus: video.SerialStatus,
		Cover:        video.CoverImage,
		Poster:       video.PosterImage,
		Year:         video.Year,
		Region:       video.Region,
		Language:     video.Language,
		Duration:     video.Duration,
		ReleaseDate:  video.ReleaseDate,
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

func (s *videoService) ensureCategory(ctx context.Context, id uint32) error {
	c, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("video: ensure category get failed", zap.Uint32("category_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.CategoryNotFound
	}
	return nil
}

func (s *videoService) ensureDirectors(ctx context.Context, ids []uint32) error {
	ids = utils.UniqueUint32IDs(ids)
	if len(ids) == 0 {
		return nil
	}
	items, err := s.metaRepo.GetDirectorsByIDs(ctx, ids)
	if err != nil {
		s.log.Error("video: ensure directors get failed", zap.Any("director_ids", ids), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(items) != len(ids) {
		return errcode.DirectorNotFound
	}
	return nil
}

func (s *videoService) ensureActors(ctx context.Context, actors []dto.VideoActorInput) error {
	ids := make([]uint32, 0, len(actors))
	seen := map[uint32]struct{}{}
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
		s.log.Error("video: ensure actors get failed", zap.Any("actor_ids", ids), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(items) != len(ids) {
		return errcode.ActorNotFound
	}
	return nil
}

func (s *videoService) ensureTags(ctx context.Context, ids []uint32) error {
	ids = utils.UniqueUint32IDs(ids)
	if len(ids) == 0 {
		return nil
	}
	items, err := s.metaRepo.GetTagsByIDs(ctx, ids)
	if err != nil {
		s.log.Error("video: ensure tags get failed", zap.Any("tag_ids", ids), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(items) != len(ids) {
		return errcode.TagNotFound
	}
	return nil
}

func mapVideoList(items []model.Videos, catMap map[uint32]*model.Categories) []dto.VideoListItem {
	out := make([]dto.VideoListItem, 0, len(items))
	for _, v := range items {
		out = append(out, dto.VideoListItem{
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
			CategoryName:  resolveCategoryName(catMap, v.CategoryID),
			PublishStatus: v.PublishStatus,
			SerialStatus:  v.SerialStatus,
			Duration:      v.Duration,
			ViewCount:     v.ViewCount,
			CreatedAt:     utils.FormatTimeStr(v.CreatedAt),
			UpdatedAt:     utils.FormatTimeStr(v.UpdatedAt),
		})
	}
	return out
}

// resolveCategoryName builds a display string for a category. If the category
// has a parent, the format is "parent/child"; otherwise just the category name.
func resolveCategoryName(catMap map[uint32]*model.Categories, id uint32) string {
	cat, ok := catMap[id]
	if !ok {
		return ""
	}
	if cat.ParentID > 0 {
		if parent, ok := catMap[cat.ParentID]; ok {
			return parent.Name + "/" + cat.Name
		}
	}
	return cat.Name
}

// collectDescendantIDs gathers the given category ID and all its descendant IDs
// from the flat category map.
func collectDescendantIDs(catMap map[uint32]*model.Categories, id uint32) []uint32 {
	ids := []uint32{id}
	for _, cat := range catMap {
		if cat.ParentID == id {
			ids = append(ids, collectDescendantIDs(catMap, cat.ID)...)
		}
	}
	return ids
}

func toActorRels(inputs []dto.VideoActorInput) []model.VideoActors {
	seen := map[uint32]struct{}{}
	out := make([]model.VideoActors, 0, len(inputs))
	for _, in := range inputs {
		if in.ActorID == 0 {
			continue
		}
		if _, ok := seen[in.ActorID]; ok {
			continue
		}
		seen[in.ActorID] = struct{}{}
		out = append(out, model.VideoActors{
			ActorID: in.ActorID,
		})
	}
	return out
}

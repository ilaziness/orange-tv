package open

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	opendto "github.com/ilaziness/orange-tv/internal/dto/open"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// ResourceService serves third-party resource-station APIs.
type ResourceService interface {
	// Enabled reports whether third-party collect is enabled.
	Enabled(ctx context.Context) bool
	ListVideos(ctx context.Context, page, pageSize int) ([]opendto.VideoListItem, int, error)
	GetVideo(ctx context.Context, ids []int64) ([]opendto.VideoDetailItem, error)
	ListCategories(ctx context.Context) ([]opendto.CategoryItem, error)
}

type resourceService struct {
	settingsRepo repository.SettingsRepository
	cache        *cache.Manager
	videoRepo    repository.VideoRepository
	metaRepo     repository.MetadataRepository
	playRepo     repository.PlayRepository
	catRepo      repository.CategoryRepository
	log          *zap.Logger
}

// NewResourceService creates a ResourceService.
func NewResourceService(
	settingsRepo repository.SettingsRepository,
	videoRepo repository.VideoRepository,
	metaRepo repository.MetadataRepository,
	playRepo repository.PlayRepository,
	catRepo repository.CategoryRepository,
	c *cache.Manager,
	log *zap.Logger,
) ResourceService {
	if log == nil {
		log = zap.NewNop()
	}
	return &resourceService{
		settingsRepo: settingsRepo, videoRepo: videoRepo, metaRepo: metaRepo,
		playRepo: playRepo, catRepo: catRepo, cache: c, log: log,
	}
}

func (s *resourceService) Enabled(ctx context.Context) bool {
	m, err := s.loadAPIConfig(ctx)
	if err != nil {
		return false
	}
	return service.BoolVal(m, constant.SettingEnableThirdPartyCollect, true)
}

func (s *resourceService) loadAPIConfig(ctx context.Context) (map[string]model.SystemSettings, error) {
	if m, err := s.cache.GetSettingsByGroup(ctx, constant.SettingGroupAPI); err == nil && m != nil {
		return m, nil
	}
	m, err := s.settingsRepo.GetByGroup(ctx, constant.SettingGroupAPI)
	if err != nil {
		s.log.Error("open resource: load api config failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.cache.SetSettingsByGroup(ctx, constant.SettingGroupAPI, m)
	return m, nil
}

type openVideoListCache struct {
	Items []opendto.VideoListItem
	Total int
}

func (s *resourceService) ListVideos(ctx context.Context, page, pageSize int) ([]opendto.VideoListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	cacheKey := cache.OpenVideoListKey(page, pageSize)
	if v, err := s.cache.GetOpenVideoList(ctx, cacheKey); err == nil && v != nil {
		if cached, ok := v.(*openVideoListCache); ok {
			return cached.Items, cached.Total, nil
		}
	}

	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		OnlyOnline: true,
		Sort:       "id_desc",
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	})
	if err != nil {
		s.log.Error("open resource: list videos failed", zap.Int("page", page), zap.Int("page_size", pageSize), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}

	list := make([]opendto.VideoListItem, 0, len(items))
	for _, it := range items {
		list = append(list, mapDefaultListItem(&it))
	}
	_ = s.cache.SetOpenVideoList(ctx, cacheKey, &openVideoListCache{Items: list, Total: total})
	return list, total, nil
}

func (s *resourceService) GetVideo(ctx context.Context, ids []int64) ([]opendto.VideoDetailItem, error) {
	if len(ids) == 0 {
		return nil, errcode.ParamError
	}
	if len(ids) > 50 {
		return nil, errcode.WithMessage(errcode.ParamError, "最多支持 50 个视频 id")
	}
	u64IDs := make([]uint64, 0, len(ids))
	for _, id := range ids {
		u64IDs = append(u64IDs, uint64(id))
	}
	videos, err := s.videoRepo.GetByIDs(ctx, u64IDs)
	if err != nil {
		s.log.Error("open resource: get videos by ids failed", zap.Int("count", len(ids)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	list := make([]opendto.VideoDetailItem, 0, len(videos))
	for i := range videos {
		detail, err := s.buildDetail(ctx, &videos[i])
		if err != nil {
			return nil, err
		}
		list = append(list, mapDefaultDetail(detail))
	}

	return list, nil
}

func (s *resourceService) ListCategories(ctx context.Context) ([]opendto.CategoryItem, error) {
	if cached, err := s.cache.GetOpenCategories(ctx); err == nil && cached != nil {
		return cached, nil
	}
	items, err := s.catRepo.List(ctx, true)
	if err != nil {
		s.log.Error("open resource: list categories failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]opendto.CategoryItem, 0, len(items))
	for _, c := range items {
		out = append(out, opendto.CategoryItem{
			ID:       c.ID,
			Name:     c.Name,
			ParentID: c.ParentID,
		})
	}
	_ = s.cache.SetOpenCategories(ctx, out)
	return out, nil
}

type detailBundle struct {
	Video     *model.Videos
	Directors []model.Directors
	Actors    []shareddto.NamedItem
	Tags      []model.Tags
	Sources   []opendto.VideoSource
}

func (s *resourceService) buildDetail(ctx context.Context, video *model.Videos) (*detailBundle, error) {
	directorIDs, err := s.videoRepo.ListDirectorIDs(ctx, video.ID)
	if err != nil {
		s.log.Error("open resource: load detail list director ids failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	directors, err := s.metaRepo.GetDirectorsByIDs(ctx, directorIDs)
	if err != nil {
		s.log.Error("open resource: load detail get directors failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorRels, err := s.videoRepo.ListActorRels(ctx, video.ID)
	if err != nil {
		s.log.Error("open resource: load detail list actor rels failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorIDs := make([]uint64, 0, len(actorRels))
	for _, rel := range actorRels {
		actorIDs = append(actorIDs, rel.ActorID)
	}
	actors, err := s.metaRepo.GetActorsByIDs(ctx, actorIDs)
	if err != nil {
		s.log.Error("open resource: load detail get actors failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorName := map[uint64]string{}
	for _, a := range actors {
		actorName[a.ID] = a.Name
	}
	actorItems := make([]shareddto.NamedItem, 0, len(actorRels))
	for _, rel := range actorRels {
		actorItems = append(actorItems, shareddto.NamedItem{ID: rel.ActorID, Name: actorName[rel.ActorID]})
	}
	tagIDs, err := s.videoRepo.ListTagIDs(ctx, video.ID)
	if err != nil {
		s.log.Error("open resource: load detail list tag ids failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tags, err := s.metaRepo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		s.log.Error("open resource: load detail get tags failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	episodes, err := s.playRepo.ListEpisodesByVideo(ctx, int64(video.ID), true)
	if err != nil {
		s.log.Error("open resource: load detail list episodes failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sources, err := s.playRepo.ListSources(ctx)
	if err != nil {
		s.log.Error("open resource: load detail list sources failed", zap.Uint64("video_id", video.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sourceMap := map[uint64]model.PlaySources{}
	for _, src := range sources {
		if src.Status != constant.StatusEnabled {
			continue
		}
		sourceMap[src.ID] = src
	}
	groups := map[uint64]*opendto.VideoSource{}
	order := make([]uint64, 0)
	for _, ep := range episodes {
		src, ok := sourceMap[ep.SourceID]
		if !ok {
			continue
		}
		g, ok := groups[ep.SourceID]
		if !ok {
			g = &opendto.VideoSource{ID: src.ID, Name: src.Name, Episodes: []opendto.VideoSourceEpisode{}}
			groups[ep.SourceID] = g
			order = append(order, ep.SourceID)
		}
		g.Episodes = append(g.Episodes, opendto.VideoSourceEpisode{
			Episode: ep.EpisodeNumber, Title: ep.Title, URL: ep.PlayURL,
		})
	}
	sourceGroups := make([]opendto.VideoSource, 0, len(order))
	for _, sid := range order {
		sourceGroups = append(sourceGroups, *groups[sid])
	}
	return &detailBundle{
		Video: video, Directors: directors, Actors: actorItems, Tags: tags, Sources: sourceGroups,
	}, nil
}

func mapDefaultListItem(v *model.Videos) opendto.VideoListItem {
	return opendto.VideoListItem{
		ID:         v.ID,
		Title:      v.Title,
		CategoryID: v.CategoryID,
		CreatedAt:  utils.FormatTimeStr(v.CreatedAt),
	}
}

func mapDefaultDetail(d *detailBundle) opendto.VideoDetailItem {
	v := d.Video
	desc := ""
	if v.Description != nil {
		desc = *v.Description
	}
	dirs := make([]string, 0, len(d.Directors))
	for _, d0 := range d.Directors {
		dirs = append(dirs, d0.Name)
	}
	acts := make([]string, 0, len(d.Actors))
	for _, a := range d.Actors {
		acts = append(acts, a.Name)
	}
	return opendto.VideoDetailItem{
		ID:          v.ID,
		Title:       v.Title,
		Subtitle:    v.Subtitle,
		Cover:       v.CoverImage,
		CategoryID:  v.CategoryID,
		Year:        v.Year,
		Rating:      v.Rating,
		ReleaseDate: v.ReleaseDate,
		Region:      v.Region,
		Language:    v.Language,
		Description: desc,
		Directors:   dirs,
		Actors:      acts,
		Sources:     d.Sources,
		CreatedAt:   utils.FormatTimeStr(v.CreatedAt),
	}
}

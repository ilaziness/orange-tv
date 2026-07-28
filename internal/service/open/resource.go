package open

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strconv"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// ResourceService serves third-party resource-station APIs.
type ResourceService interface {
	// Authorize checks third-party collect switch and optional API key.
	Authorize(ctx context.Context, providedKey string) (*adminsvc.ResourceConfig, error)
	ListVideos(ctx context.Context, page, pageSize int, format string) (any, error)
	GetVideo(ctx context.Context, id int64, format string) (any, error)
	ListCategories(ctx context.Context) ([]shareddto.CategoryResponse, error)
}

type resourceService struct {
	settings  adminsvc.SettingsService
	videoRepo repository.VideoRepository
	metaRepo  repository.MetadataRepository
	playRepo  repository.PlayRepository
	catRepo   repository.CategoryRepository
	cache     *cache.Manager
	log       *zap.Logger
}

// NewResourceService creates a ResourceService.
func NewResourceService(
	settings adminsvc.SettingsService,
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
		settings: settings, videoRepo: videoRepo, metaRepo: metaRepo,
		playRepo: playRepo, catRepo: catRepo, cache: c, log: log,
	}
}

func (s *resourceService) Authorize(ctx context.Context, providedKey string) (*adminsvc.ResourceConfig, error) {
	cfg, err := s.settings.ResourceConfig(ctx)
	if err != nil {
		s.log.Error("open resource: get resource config failed", zap.Error(err))
		return nil, err
	}
	if !cfg.EnableThirdPartyCollect {
		return nil, errcode.ResourceAPIDisabled
	}
	expected := strings.TrimSpace(cfg.APIKey)
	if expected != "" {
		got := strings.TrimSpace(providedKey)
		// ConstantTimeCompare panics when lengths differ; mismatch is always invalid.
		if len(got) != len(expected) || subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
			return nil, errcode.ResourceAPIKeyInvalid
		}
	}
	return cfg, nil
}

func (s *resourceService) ListVideos(ctx context.Context, page, pageSize int, format string) (any, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	format, nerr := normalizeFormat(format)
	if nerr != nil {
		return nil, nerr
	}
	cacheKey := cache.OpenVideoListKey(format, page, pageSize)
	if v, err := s.cache.GetOpenVideoList(ctx, cacheKey); err == nil && v != nil {
		return v, nil
	}

	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		OnlyOnline: true,
		Sort:       "id_desc",
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	})
	if err != nil {
		s.log.Error("open resource: list videos failed", zap.Int("page", page), zap.Int("page_size", pageSize), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	totalPages := 0
	if pageSize > 0 && total > 0 {
		totalPages = total / pageSize
		if total%pageSize > 0 {
			totalPages++
		}
	}

	var out any
	if format == constant.APIOutputAppleCMS {
		list := make([]map[string]any, 0, len(items))
		for _, it := range items {
			list = append(list, mapAppleListItem(&it))
		}
		out = map[string]any{
			"code":      1,
			"msg":       "数据列表",
			"page":      page,
			"pagecount": totalPages,
			"limit":     strconv.Itoa(pageSize),
			"total":     total,
			"list":      list,
		}
	} else {
		list := make([]map[string]any, 0, len(items))
		for _, it := range items {
			list = append(list, mapDefaultListItem(&it))
		}
		out = map[string]any{
			"code":    200,
			"message": "success",
			"data": map[string]any{
				"list": list,
				"pagination": map[string]any{
					"page":     page,
					"pageSize": pageSize,
					"total":    total,
				},
			},
		}
	}
	_ = s.cache.SetOpenVideoList(ctx, cacheKey, out)
	return out, nil
}

func (s *resourceService) GetVideo(ctx context.Context, id int64, format string) (any, error) {
	if id <= 0 {
		return nil, errcode.ParamError
	}
	format, nerr := normalizeFormat(format)
	if nerr != nil {
		return nil, nerr
	}
	cacheKey := cache.OpenVideoDetailKey(format, id)
	if v, err := s.cache.GetOpenVideoDetail(ctx, cacheKey); err == nil && v != nil {
		return v, nil
	}

	detail, err := s.loadDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	var out any
	if format == constant.APIOutputAppleCMS {
		out = map[string]any{
			"code": 1,
			"msg":  "数据详情",
			"list": []map[string]any{mapAppleDetail(detail)},
		}
	} else {
		out = map[string]any{
			"code":    200,
			"message": "success",
			"data":    mapDefaultDetail(detail),
		}
	}
	_ = s.cache.SetOpenVideoDetail(ctx, cacheKey, out)
	return out, nil
}

func (s *resourceService) ListCategories(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	if tree, err := s.cache.GetOpenCategories(ctx); err == nil && tree != nil {
		return tree, nil
	}
	items, err := s.catRepo.List(ctx, true)
	if err != nil {
		s.log.Error("open resource: list categories failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tree := utils.BuildCategoryTree(items)
	_ = s.cache.SetOpenCategories(ctx, tree)
	return tree, nil
}

type detailBundle struct {
	Video     *model.Videos
	Directors []model.Directors
	Actors    []shareddto.ActorItem
	Tags      []model.Tags
	Sources   []shareddto.VideoSourceGroup
}

func (s *resourceService) loadDetail(ctx context.Context, id int64) (*detailBundle, error) {
	video, err := s.videoRepo.GetByID(ctx, uint64(id))
	if err != nil {
		s.log.Error("open resource: load detail get video failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil || video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}
	directorIDs, err := s.videoRepo.ListDirectorIDs(ctx, uint64(id))
	if err != nil {
		s.log.Error("open resource: load detail list director ids failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	directors, err := s.metaRepo.GetDirectorsByIDs(ctx, directorIDs)
	if err != nil {
		s.log.Error("open resource: load detail get directors failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorRels, err := s.videoRepo.ListActorRels(ctx, uint64(id))
	if err != nil {
		s.log.Error("open resource: load detail list actor rels failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorIDs := make([]uint64, 0, len(actorRels))
	for _, rel := range actorRels {
		actorIDs = append(actorIDs, rel.ActorID)
	}
	actors, err := s.metaRepo.GetActorsByIDs(ctx, actorIDs)
	if err != nil {
		s.log.Error("open resource: load detail get actors failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorName := map[uint64]string{}
	for _, a := range actors {
		actorName[a.ID] = a.Name
	}
	actorItems := make([]shareddto.ActorItem, 0, len(actorRels))
	for _, rel := range actorRels {
		actorItems = append(actorItems, shareddto.ActorItem{ID: rel.ActorID, Name: actorName[rel.ActorID], Role: rel.Role})
	}
	tagIDs, err := s.videoRepo.ListTagIDs(ctx, uint64(id))
	if err != nil {
		s.log.Error("open resource: load detail list tag ids failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tags, err := s.metaRepo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		s.log.Error("open resource: load detail get tags failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	episodes, err := s.playRepo.ListEpisodesByVideo(ctx, int64(id), true)
	if err != nil {
		s.log.Error("open resource: load detail list episodes failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sources, err := s.playRepo.ListSources(ctx)
	if err != nil {
		s.log.Error("open resource: load detail list sources failed", zap.Int64("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sourceMap := map[uint64]model.PlaySources{}
	for _, src := range sources {
		if src.Status != constant.StatusEnabled {
			continue
		}
		sourceMap[src.ID] = src
	}
	groups := map[uint64]*shareddto.VideoSourceGroup{}
	order := make([]uint64, 0)
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
			Episode: ep.EpisodeNumber, Title: ep.Title, URL: ep.PlayURL, Quality: ep.Quality, Format: ep.Format,
		})
	}
	sourceGroups := make([]shareddto.VideoSourceGroup, 0, len(order))
	for _, sid := range order {
		sourceGroups = append(sourceGroups, *groups[sid])
	}
	return &detailBundle{
		Video: video, Directors: directors, Actors: actorItems, Tags: tags, Sources: sourceGroups,
	}, nil
}

// normalizeFormat: empty/default → system format; only apple_cms is alternate.
// Non-empty unknown values are rejected.
func normalizeFormat(format string) (string, error) {
	format = strings.TrimSpace(strings.ToLower(format))
	switch format {
	case "", constant.APIOutputDefault:
		return constant.APIOutputDefault, nil
	case constant.APIOutputAppleCMS:
		return constant.APIOutputAppleCMS, nil
	default:
		return "", errcode.WithMessage(errcode.ParamError, "format 仅支持 default 或 apple_cms")
	}
}

func mapDefaultListItem(v *model.Videos) map[string]any {
	return map[string]any{
		"id":          strconv.FormatUint(v.ID, 10),
		"title":       v.Title,
		"cover":       v.CoverImage,
		"category_id": v.CategoryID,
		"year":        v.Year,
		"rating":      v.Rating,
		"region":      v.Region,
	}
}

func mapDefaultDetail(d *detailBundle) map[string]any {
	v := d.Video
	desc := ""
	if v.Description != nil {
		desc = *v.Description
	}
	sources := make([]map[string]any, 0, len(d.Sources))
	for _, src := range d.Sources {
		eps := make([]map[string]any, 0, len(src.Episodes))
		for _, ep := range src.Episodes {
			eps = append(eps, map[string]any{"episode": ep.Episode, "url": ep.URL, "title": ep.Title})
		}
		sources = append(sources, map[string]any{"name": src.Name, "episodes": eps})
	}
	dirs := make([]string, 0, len(d.Directors))
	for _, d0 := range d.Directors {
		dirs = append(dirs, d0.Name)
	}
	acts := make([]string, 0, len(d.Actors))
	for _, a := range d.Actors {
		acts = append(acts, a.Name)
	}
	return map[string]any{
		"id":          strconv.FormatUint(v.ID, 10),
		"title":       v.Title,
		"subtitle":    v.Subtitle,
		"cover":       v.CoverImage,
		"category_id": v.CategoryID,
		"year":        v.Year,
		"rating":      v.Rating,
		"region":      v.Region,
		"language":    v.Language,
		"description": desc,
		"directors":   dirs,
		"actors":      acts,
		"sources":     sources,
	}
}

func mapAppleListItem(v *model.Videos) map[string]any {
	return map[string]any{
		"vod_id":       strconv.FormatUint(v.ID, 10),
		"type_id":      strconv.FormatUint(v.CategoryID, 10),
		"vod_name":     v.Title,
		"vod_sub":      v.Subtitle,
		"vod_pic":      v.CoverImage,
		"vod_year":     v.Year,
		"vod_area":     v.Region,
		"vod_lang":     v.Language,
		"vod_score":    v.Rating,
		"douban_score": v.Rating,
	}
}

func mapAppleDetail(d *detailBundle) map[string]any {
	v := d.Video
	desc := ""
	if v.Description != nil {
		desc = *v.Description
	}
	// join play_from / play_url as apple cms style
	fromParts := make([]string, 0, len(d.Sources))
	urlParts := make([]string, 0, len(d.Sources))
	for _, src := range d.Sources {
		fromParts = append(fromParts, src.Name)
		epParts := make([]string, 0, len(src.Episodes))
		for _, ep := range src.Episodes {
			title := ep.Title
			if title == "" {
				title = fmt.Sprintf("第%d集", ep.Episode)
			}
			epParts = append(epParts, title+"$"+ep.URL)
		}
		urlParts = append(urlParts, strings.Join(epParts, "#"))
	}
	dirs := make([]string, 0, len(d.Directors))
	for _, d0 := range d.Directors {
		dirs = append(dirs, d0.Name)
	}
	acts := make([]string, 0, len(d.Actors))
	for _, a := range d.Actors {
		acts = append(acts, a.Name)
	}
	return map[string]any{
		"vod_id":        strconv.FormatUint(v.ID, 10),
		"type_id":       strconv.FormatUint(v.CategoryID, 10),
		"vod_name":      v.Title,
		"vod_sub":       v.Subtitle,
		"vod_pic":       v.CoverImage,
		"vod_actor":     strings.Join(acts, ","),
		"vod_director":  strings.Join(dirs, ","),
		"vod_content":   desc,
		"vod_year":      v.Year,
		"vod_area":      v.Region,
		"vod_lang":      v.Language,
		"douban_score":  v.Rating,
		"vod_play_from": strings.Join(fromParts, "$$$"),
		"vod_play_url":  strings.Join(urlParts, "$$$"),
	}
}

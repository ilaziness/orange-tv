package client

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// VideoService provides client video queries.
type VideoService interface {
	List(ctx context.Context, req *clientdto.VideoListRequest) ([]clientdto.VideoListItem, int, error)
	Search(ctx context.Context, req *clientdto.SearchRequest) ([]clientdto.VideoListItem, int, error)
	Get(ctx context.Context, id uint32) (*clientdto.VideoDetailResponse, error)
	Related(ctx context.Context, id uint32, limit int) ([]clientdto.VideoListItem, error)
	GetEpisode(ctx context.Context, videoID, episodeID uint32) (*clientdto.PlayEpisodeResponse, error)
}

type videoService struct {
	videoRepo repository.VideoRepository
	metaRepo  repository.MetadataRepository
	playRepo  repository.PlayRepository
	cache     *cache.Manager
	log       *zap.Logger
}

// NewVideoService creates a client VideoService.
func NewVideoService(
	videoRepo repository.VideoRepository,
	metaRepo repository.MetadataRepository,
	playRepo repository.PlayRepository,
	c *cache.Manager,
	log *zap.Logger,
) VideoService {
	if log == nil {
		log = zap.NewNop()
	}
	return &videoService{videoRepo: videoRepo, metaRepo: metaRepo, playRepo: playRepo, cache: c, log: log}
}

func (s *videoService) List(ctx context.Context, req *clientdto.VideoListRequest) ([]clientdto.VideoListItem, int, error) {
	// 默认按年份倒序；首页等场景可通过 sort 参数指定其他排序
	sort := strings.TrimSpace(req.Sort)
	if sort == "" {
		sort = "year_desc"
	}
	// Cache hot home/list queries without keyword search complexity.
	cacheable := strings.TrimSpace(req.Region) == "" && req.YearStart == 0 && req.YearEnd == 0
	cacheKey := ""
	if cacheable {
		cacheKey = cache.VideoListKey(req.CategoryID, req.ParentCategoryID, sort, req.GetPage(), req.GetLimit())
		if e, err := s.cache.GetVideoListClient(ctx, cacheKey); err == nil && e != nil {
			return e.Items, e.Total, nil
		}
	}

	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		CategoryID:       req.CategoryID,
		ParentCategoryID: req.ParentCategoryID,
		YearStart:        req.YearStart,
		YearEnd:          req.YearEnd,
		Region:           strings.TrimSpace(req.Region),
		Sort:             sort,
		OnlyOnline:       true,
		Offset:           req.GetOffset(),
		Limit:            req.GetLimit(),
	})
	if err != nil {
		s.log.Error("client video: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := mapVideoList(items, s.loadVideoTags(ctx, videoIDsFromItems(items)))
	if cacheKey != "" {
		_ = s.cache.SetVideoListClient(ctx, cacheKey, &cache.VideoListCacheEntry{Items: out, Total: total})
	}
	return out, total, nil
}

func (s *videoService) Search(ctx context.Context, req *clientdto.SearchRequest) ([]clientdto.VideoListItem, int, error) {
	sort := strings.TrimSpace(req.Sort)
	if sort == "" {
		sort = "year_desc"
	}
	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		Keyword:          strings.TrimSpace(req.Keyword),
		CategoryID:       req.CategoryID,
		ParentCategoryID: req.ParentCategoryID,
		YearStart:        req.YearStart,
		YearEnd:          req.YearEnd,
		Region:           strings.TrimSpace(req.Region),
		Sort:             sort,
		OnlyOnline:       true,
		Offset:           req.GetOffset(),
		Limit:            req.GetLimit(),
	})
	if err != nil {
		s.log.Error("client video: search failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapVideoList(items, s.loadVideoTags(ctx, videoIDsFromItems(items))), total, nil
}

func (s *videoService) Related(ctx context.Context, id uint32, limit int) ([]clientdto.VideoListItem, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("client video: related get video failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil || video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}
	// practical related: same category first (skip if category unset), then same region fill
	out := make([]model.Videos, 0, limit)
	seen := map[uint32]bool{id: true}
	if video.CategoryID > 0 {
		candidates, _, err := s.videoRepo.List(ctx, repository.VideoListFilter{
			CategoryID: video.CategoryID,
			Sort:       "rating_desc",
			OnlyOnline: true,
			Offset:     0,
			Limit:      limit + 5,
		})
		if err != nil {
			s.log.Error("client video: related list by category failed", zap.Uint32("video_id", id), zap.Uint32("category_id", video.CategoryID), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		for _, it := range candidates {
			if seen[it.ID] {
				continue
			}
			seen[it.ID] = true
			out = append(out, it)
			if len(out) >= limit {
				break
			}
		}
	}
	if len(out) < limit && strings.TrimSpace(video.Region) != "" {
		more, _, err := s.videoRepo.List(ctx, repository.VideoListFilter{
			Region:     strings.TrimSpace(video.Region),
			Sort:       "view_count_desc",
			OnlyOnline: true,
			Offset:     0,
			Limit:      limit + 5,
		})
		if err != nil {
			s.log.Error("client video: related list by region failed", zap.Uint32("video_id", id), zap.String("region", video.Region), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		for _, it := range more {
			if seen[it.ID] {
				continue
			}
			seen[it.ID] = true
			out = append(out, it)
			if len(out) >= limit {
				break
			}
		}
	}
	return mapVideoList(out, s.loadVideoTags(ctx, videoIDsFromItems(out))), nil
}

func (s *videoService) Get(ctx context.Context, id uint32) (*clientdto.VideoDetailResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("client video: get by id failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil || video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}

	directorIDs, err := s.videoRepo.ListDirectorIDs(ctx, id)
	if err != nil {
		s.log.Error("client video: list director ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	directors, err := s.metaRepo.GetDirectorsByIDs(ctx, directorIDs)
	if err != nil {
		s.log.Error("client video: get directors failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorRels, err := s.videoRepo.ListActorRels(ctx, id)
	if err != nil {
		s.log.Error("client video: list actor rels failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorIDs := make([]uint32, 0, len(actorRels))
	for _, rel := range actorRels {
		actorIDs = append(actorIDs, rel.ActorID)
	}
	actors, err := s.metaRepo.GetActorsByIDs(ctx, actorIDs)
	if err != nil {
		s.log.Error("client video: get actors failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorName := map[uint32]string{}
	for _, a := range actors {
		actorName[a.ID] = a.Name
	}
	tagIDs, err := s.videoRepo.ListTagIDs(ctx, id)
	if err != nil {
		s.log.Error("client video: list tag ids failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tags, err := s.metaRepo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		s.log.Error("client video: get tags failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	episodes, err := s.playRepo.ListEpisodesByVideo(ctx, id, true)
	if err != nil {
		s.log.Error("client video: list episodes failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sources, err := s.playRepo.ListSources(ctx)
	if err != nil {
		s.log.Error("client video: list sources failed", zap.Uint32("video_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sourceMap := map[uint32]model.PlaySources{}
	for _, src := range sources {
		if src.Status != constant.StatusEnabled {
			continue
		}
		sourceMap[src.ID] = src
	}

	groups := map[uint32]*clientdto.VideoDetailSourceGroup{}
	order := make([]uint32, 0)
	for _, ep := range episodes {
		src, ok := sourceMap[ep.SourceID]
		if !ok {
			continue
		}
		g, ok := groups[ep.SourceID]
		if !ok {
			g = &clientdto.VideoDetailSourceGroup{ID: src.ID, Name: src.Name, Episodes: []clientdto.VideoDetailEpisode{}}
			groups[ep.SourceID] = g
			order = append(order, ep.SourceID)
		}
		g.Episodes = append(g.Episodes, clientdto.VideoDetailEpisode{
			ID:      ep.ID,
			Episode: ep.EpisodeNumber,
			Title:   ep.Title,
		})
	}
	sourceGroups := make([]clientdto.VideoDetailSourceGroup, 0, len(order))
	for _, sid := range order {
		sourceGroups = append(sourceGroups, *groups[sid])
	}

	dirItems := make([]shareddto.NamedItem, 0, len(directors))
	for _, d := range directors {
		dirItems = append(dirItems, shareddto.NamedItem{ID: d.ID, Name: d.Name})
	}
	actorItems := make([]shareddto.NamedItem, 0, len(actorRels))
	for _, rel := range actorRels {
		actorItems = append(actorItems, shareddto.NamedItem{ID: rel.ActorID, Name: actorName[rel.ActorID]})
	}
	tagItems := make([]shareddto.NamedItem, 0, len(tags))
	for _, tg := range tags {
		tagItems = append(tagItems, shareddto.NamedItem{ID: tg.ID, Name: tg.Name})
	}
	desc := video.Description
	return &clientdto.VideoDetailResponse{
		ID: video.ID, Title: video.Title, Subtitle: video.Subtitle, Description: desc,
		CategoryID: video.CategoryID, SerialStatus: video.SerialStatus, Cover: video.CoverImage, Poster: video.PosterImage,
		Year: video.Year, Region: video.Region, Language: video.Language, Duration: video.Duration,
		ReleaseDate: video.ReleaseDate, Rating: video.Rating, RatingCount: video.RatingCount, ViewCount: video.ViewCount,
		Directors: dirItems, Actors: actorItems, Tags: tagItems, Sources: sourceGroups,
	}, nil
}

func (s *videoService) GetEpisode(ctx context.Context, videoID, episodeID uint32) (*clientdto.PlayEpisodeResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("client video: get episode - get video failed", zap.Uint32("video_id", videoID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil || video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}
	ep, err := s.playRepo.GetPlayableEpisodeByID(ctx, videoID, episodeID)
	if err != nil {
		s.log.Error("client video: get playable episode failed", zap.Uint32("video_id", videoID), zap.Uint32("episode_id", episodeID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if ep == nil {
		return nil, errcode.PlayEpisodeNotFound
	}
	return &clientdto.PlayEpisodeResponse{
		URL:     ep.PlayURL,
		Quality: ep.Quality,
		Format:  ep.Format,
	}, nil
}

func mapVideoList(items []model.Videos, tagMap map[uint32][]shareddto.NamedItem) []clientdto.VideoListItem {
	out := make([]clientdto.VideoListItem, 0, len(items))
	for _, v := range items {
		out = append(out, clientdto.VideoListItem{
			ID: v.ID, Title: v.Title, Subtitle: v.Subtitle, Cover: v.CoverImage, Poster: v.PosterImage,
			Year: v.Year, Region: v.Region, Language: v.Language, Rating: v.Rating, CategoryID: v.CategoryID,
			SerialStatus: v.SerialStatus, Duration: v.Duration, ViewCount: v.ViewCount,
			Tags: tagMap[v.ID],
		})
	}
	return out
}

// loadVideoTags batch-loads up to 2 tags per video. Returns an empty map on error
// so the list degrades gracefully (no tags shown).
func (s *videoService) loadVideoTags(ctx context.Context, videoIDs []uint32) map[uint32][]shareddto.NamedItem {
	tagMap := make(map[uint32][]shareddto.NamedItem)
	if len(videoIDs) == 0 {
		return tagMap
	}
	rows, err := s.videoRepo.ListTagsByVideoIDs(ctx, videoIDs)
	if err != nil {
		s.log.Error("client video: load tags failed", zap.Error(err))
		return tagMap
	}
	for _, row := range rows {
		if len(tagMap[row.VideoID]) >= 2 {
			continue
		}
		tagMap[row.VideoID] = append(tagMap[row.VideoID], shareddto.NamedItem{ID: row.TagID, Name: row.Name})
	}
	return tagMap
}

// videoIDsFromItems extracts video IDs from a model.Videos slice.
func videoIDsFromItems(items []model.Videos) []uint32 {
	ids := make([]uint32, 0, len(items))
	for _, v := range items {
		ids = append(ids, v.ID)
	}
	return ids
}

package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// VideoService provides client video queries.
type VideoService interface {
	List(ctx context.Context, req *clientdto.VideoListRequest) ([]shareddto.VideoListItem, int, error)
	Search(ctx context.Context, req *clientdto.SearchRequest) ([]shareddto.VideoListItem, int, error)
	Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error)
	Related(ctx context.Context, id int64, limit int) ([]shareddto.VideoListItem, error)
}

type videoService struct {
	videoRepo repository.VideoRepository
	metaRepo  repository.MetadataRepository
	playRepo  repository.PlayRepository
	cache     cache.Cache
}

type videoListCacheEntry struct {
	Items []shareddto.VideoListItem
	Total int
}

// NewVideoService creates a client VideoService.
func NewVideoService(
	videoRepo repository.VideoRepository,
	metaRepo repository.MetadataRepository,
	playRepo repository.PlayRepository,
	c cache.Cache,
) VideoService {
	if c == nil {
		c = cache.NewNopCache()
	}
	return &videoService{videoRepo: videoRepo, metaRepo: metaRepo, playRepo: playRepo, cache: c}
}

func (s *videoService) List(ctx context.Context, req *clientdto.VideoListRequest) ([]shareddto.VideoListItem, int, error) {
	// Cache hot home/list queries without keyword search complexity.
	cacheable := strings.TrimSpace(req.Region) == "" && strings.TrimSpace(req.Language) == "" && req.Year == 0
	cacheKey := ""
	if cacheable {
		cacheKey = fmt.Sprintf("video:list:%d:%s:%d:%d", req.CategoryID, req.Sort, req.GetPage(), req.GetLimit())
		if v, err := s.cache.Get(ctx, cacheKey); err == nil {
			if e, ok := v.(*videoListCacheEntry); ok && e != nil {
				return e.Items, e.Total, nil
			}
		}
	}

	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		CategoryID: req.CategoryID,
		Year:       req.Year,
		Region:     strings.TrimSpace(req.Region),
		Language:   strings.TrimSpace(req.Language),
		Sort:       req.Sort,
		OnlyOnline: true,
		Offset:     req.GetOffset(),
		Limit:      req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := mapVideoList(items)
	if cacheKey != "" {
		_ = s.cache.Set(ctx, cacheKey, &videoListCacheEntry{Items: out, Total: total}, 2*time.Minute)
	}
	return out, total, nil
}

func (s *videoService) Search(ctx context.Context, req *clientdto.SearchRequest) ([]shareddto.VideoListItem, int, error) {
	items, total, err := s.videoRepo.List(ctx, repository.VideoListFilter{
		Keyword:    strings.TrimSpace(req.Keyword),
		CategoryID: req.CategoryID,
		Year:       req.Year,
		Region:     strings.TrimSpace(req.Region),
		Language:   strings.TrimSpace(req.Language),
		Sort:       req.Sort,
		OnlyOnline: true,
		Offset:     req.GetOffset(),
		Limit:      req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapVideoList(items), total, nil
}

func (s *videoService) Related(ctx context.Context, id int64, limit int) ([]shareddto.VideoListItem, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	video, err := s.videoRepo.GetByID(ctx, uint64(id))
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil || video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}
	// practical related: same category first (skip if category unset), then same region fill
	out := make([]model.Videos, 0, limit)
	seen := map[uint64]bool{uint64(id): true}
	if video.CategoryID > 0 {
		candidates, _, err := s.videoRepo.List(ctx, repository.VideoListFilter{
			CategoryID: video.CategoryID,
			Sort:       "rating_desc",
			OnlyOnline: true,
			Offset:     0,
			Limit:      limit + 5,
		})
		if err != nil {
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
	return mapVideoList(out), nil
}

func (s *videoService) Get(ctx context.Context, id int64) (*shareddto.VideoDetailResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, uint64(id))
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if video == nil || video.PublishStatus != constant.PublishStatusOnline {
		return nil, errcode.VideoNotFound
	}

	directorIDs, err := s.videoRepo.ListDirectorIDs(ctx, uint64(id))
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	directors, err := s.metaRepo.GetDirectorsByIDs(ctx, directorIDs)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorRels, err := s.videoRepo.ListActorRels(ctx, uint64(id))
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorIDs := make([]uint64, 0, len(actorRels))
	for _, rel := range actorRels {
		actorIDs = append(actorIDs, rel.ActorID)
	}
	actors, err := s.metaRepo.GetActorsByIDs(ctx, actorIDs)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	actorName := map[uint64]string{}
	for _, a := range actors {
		actorName[a.ID] = a.Name
	}
	tagIDs, err := s.videoRepo.ListTagIDs(ctx, uint64(id))
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	tags, err := s.metaRepo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	episodes, err := s.playRepo.ListEpisodesByVideo(ctx, int64(id), true)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	sources, err := s.playRepo.ListSources(ctx)
	if err != nil {
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
		actorItems = append(actorItems, shareddto.ActorItem{ID: rel.ActorID, Name: actorName[rel.ActorID], Role: rel.Role})
	}
	tagItems := make([]shareddto.NamedItem, 0, len(tags))
	for _, tg := range tags {
		tagItems = append(tagItems, shareddto.NamedItem{ID: tg.ID, Name: tg.Name})
	}
	desc := ""
	if video.Description != nil {
		desc = *video.Description
	}
	release := ""
	if video.ReleaseDate != nil {
		release = video.ReleaseDate.Format("2006-01-02")
	}
	return &shareddto.VideoDetailResponse{
		ID: video.ID, Title: video.Title, Subtitle: video.Subtitle, Description: desc,
		CategoryID: video.CategoryID, SerialStatus: video.SerialStatus, Cover: video.CoverImage, Poster: video.PosterImage,
		Year: video.Year, Region: video.Region, Language: video.Language, Duration: video.Duration,
		ReleaseDate: release, Rating: video.Rating, ViewCount: video.ViewCount,
		Directors: dirItems, Actors: actorItems, Tags: tagItems, Sources: sourceGroups,
	}, nil
}

func mapVideoList(items []model.Videos) []shareddto.VideoListItem {
	out := make([]shareddto.VideoListItem, 0, len(items))
	for _, v := range items {
		out = append(out, shareddto.VideoListItem{
			ID: v.ID, Title: v.Title, Subtitle: v.Subtitle, Cover: v.CoverImage, Poster: v.PosterImage,
			Year: v.Year, Region: v.Region, Language: v.Language, Rating: v.Rating, CategoryID: v.CategoryID,
			SerialStatus: v.SerialStatus, Duration: v.Duration, ViewCount: v.ViewCount,
		})
	}
	return out
}

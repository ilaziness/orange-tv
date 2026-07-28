package client

import (
	"context"
	"net/url"
	"strings"
	"time"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/pkg/cache"
	"go.uber.org/zap"
)

const (
	liveListCacheKey = "live:list:client"
	liveCacheTTL     = 5 * time.Minute
)

// LiveService provides public live channel queries.
type LiveService interface {
	List(ctx context.Context, req *clientdto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error)
	GetStreamURL(ctx context.Context, id int64) (string, error)
	// AllowedStreamDomains 返回所有在线频道的流地址域名集合，用于代理 ts 分片时的 SSRF 校验。
	AllowedStreamDomains(ctx context.Context) (map[string]struct{}, error)
}

type liveService struct {
	repo  repository.LiveRepository
	cache cache.Cache
	log   *zap.Logger
}

// NewLiveService creates a client LiveService.
func NewLiveService(repo repository.LiveRepository, c cache.Cache, log *zap.Logger) LiveService {
	if c == nil {
		c = cache.NewNopCache()
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &liveService{repo: repo, cache: c, log: log}
}

// List 返回所有在线直播频道，前端按 category 自行分组展示。
// stream_url 不返回给前端，前端通过 /live/play/:id 代理播放。
func (s *liveService) List(ctx context.Context, req *clientdto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error) {
	if v, err := s.cache.Get(ctx, liveListCacheKey); err == nil {
		if items, ok := v.([]shareddto.LiveChannelItem); ok {
			return filterByCategory(items, strings.TrimSpace(req.Category)), len(items), nil
		}
	}
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		s.log.Error("client live: list all failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]shareddto.LiveChannelItem, 0, len(items))
	for i := range items {
		m := &items[i]
		if m.Status != 1 {
			continue
		}
		out = append(out, toPublicLiveItem(m))
	}
	_ = s.cache.Set(ctx, liveListCacheKey, out, liveCacheTTL)
	return filterByCategory(out, strings.TrimSpace(req.Category)), len(out), nil
}

// GetStreamURL 返回指定频道的真实播放地址，供代理 handler 使用。
func (s *liveService) GetStreamURL(ctx context.Context, id int64) (string, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("client live: get stream url failed", zap.Int64("id", id), zap.Error(err))
		return "", errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil || item.Status != 1 {
		return "", errcode.LiveChannelNotFound
	}
	return item.StreamURL, nil
}

// AllowedStreamDomains 返回所有在线频道 stream_url 的 host 集合。
func (s *liveService) AllowedStreamDomains(ctx context.Context) (map[string]struct{}, error) {
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	domains := make(map[string]struct{}, len(items))
	for i := range items {
		if items[i].Status != 1 {
			continue
		}
		if u, err := url.Parse(items[i].StreamURL); err == nil && u.Host != "" {
			domains[u.Host] = struct{}{}
		}
	}
	return domains, nil
}

func filterByCategory(items []shareddto.LiveChannelItem, category string) []shareddto.LiveChannelItem {
	if category == "" {
		return items
	}
	filtered := make([]shareddto.LiveChannelItem, 0, len(items))
	for _, item := range items {
		if item.Category == category {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// toPublicLiveItem 将模型转为对外 DTO，stream_url 不返回给前端。
func toPublicLiveItem(m *model.LiveChannels) shareddto.LiveChannelItem {
	desc := ""
	if m.Description != nil {
		desc = *m.Description
	}
	return shareddto.LiveChannelItem{
		ID:          m.ID,
		Name:        m.Name,
		Category:    m.Category,
		StreamURL:   "",
		Logo:        m.Logo,
		Description: desc,
		SortOrder:   m.SortOrder,
	}
}

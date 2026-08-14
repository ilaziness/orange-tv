package client

import (
	"context"
	"path"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/clienttype"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// LiveService provides public live channel queries.
type LiveService interface {
	List(ctx context.Context, req *clientdto.LiveListRequest) ([]clientdto.LiveChannelItem, int, error)
	GetStreamURL(ctx context.Context, id uint32) (string, error)
}

type liveService struct {
	repo  repository.LiveRepository
	cache *cache.Manager
	log   *zap.Logger
}

// NewLiveService creates a client LiveService.
func NewLiveService(repo repository.LiveRepository, c *cache.Manager, log *zap.Logger) LiveService {
	if log == nil {
		log = zap.NewNop()
	}
	return &liveService{repo: repo, cache: c, log: log}
}

// List 返回所有在线直播频道，前端按 category 自行分组展示。
// web 端不返回 stream_url（走 /live/play/:id 代理播放）；
// app/tv/desktop 端直接返回 stream_url，由客户端直连或自行封装播放。
func (s *liveService) List(ctx context.Context, req *clientdto.LiveListRequest) ([]clientdto.LiveChannelItem, int, error) {
	withStream := clienttype.IsStreamEnabled(ctx)
	if items, err := s.cache.GetLiveListClient(ctx, withStream); err == nil && items != nil {
		return filterByCategory(items, strings.TrimSpace(req.Category)), len(items), nil
	}
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		s.log.Error("client live: list all failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.LiveChannelItem, 0, len(items))
	for i := range items {
		m := &items[i]
		if m.Status != 1 {
			continue
		}
		out = append(out, toPublicLiveItem(m, withStream))
	}
	_ = s.cache.SetLiveListClient(ctx, out, withStream)
	return filterByCategory(out, strings.TrimSpace(req.Category)), len(out), nil
}

// GetStreamURL 返回指定频道的真实播放地址，供代理 handler 使用。
func (s *liveService) GetStreamURL(ctx context.Context, id uint32) (string, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("client live: get stream url failed", zap.Uint32("id", id), zap.Error(err))
		return "", errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil || item.Status != 1 {
		return "", errcode.LiveChannelNotFound
	}
	return item.StreamURL, nil
}

func filterByCategory(items []clientdto.LiveChannelItem, category string) []clientdto.LiveChannelItem {
	if category == "" {
		return items
	}
	filtered := make([]clientdto.LiveChannelItem, 0, len(items))
	for _, item := range items {
		if item.Category == category {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// toPublicLiveItem 将模型转为对外 DTO。
// includeStream=false（web）时 StreamURL 留空，json omitempty 隐藏该字段。
func toPublicLiveItem(m *model.LiveChannels, includeStream bool) clientdto.LiveChannelItem {
	item := clientdto.LiveChannelItem{
		ID:          m.ID,
		Name:        m.Name,
		Category:    m.Category,
		Logo:        m.Logo,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		Format:      inferStreamFormat(m.StreamURL),
	}
	if includeStream {
		item.StreamURL = strings.TrimSpace(m.StreamURL)
	}
	return item
}

// inferStreamFormat 根据 stream_url 后缀推断播放格式。
// 无后缀或无法识别时默认返回 "hls"（IPTV 源以 HLS 为主）。
func inferStreamFormat(streamURL string) string {
	ext := strings.ToLower(path.Ext(strings.TrimSpace(streamURL)))
	switch ext {
	case ".flv":
		return "flv"
	case ".mp4":
		return "mp4"
	default:
		return "hls"
	}
}

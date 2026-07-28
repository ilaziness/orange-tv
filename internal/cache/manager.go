// Package cache provides business-specific cache management.
// Generic cache implementation lives in pkg/cache.
package cache

import (
	"context"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	"github.com/ilaziness/orange-tv/internal/model"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
)

// VideoListCacheEntry 客户端视频列表缓存条目。
type VideoListCacheEntry struct {
	Items []shareddto.VideoListItem
	Total int
}

// Manager 业务缓存管理器，封装底层 Cache 并提供业务级操作。
type Manager struct {
	cache pkgcache.Cache
}

// NewManager 创建业务缓存管理器，c 为 nil 时使用 NopCache。
func NewManager(c pkgcache.Cache) *Manager {
	if c == nil {
		c = pkgcache.NewNopCache()
	}
	return &Manager{cache: c}
}

// Close 关闭缓存连接。
func (m *Manager) Close() error { return m.cache.Close() }

// Clear 清空所有缓存。
func (m *Manager) Clear(ctx context.Context) error { return m.cache.Clear(ctx) }

// --- Category ---

// GetCategoryTreeClient 获取客户端分类树缓存。
func (m *Manager) GetCategoryTreeClient(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	v, err := m.cache.Get(ctx, KeyCategoryTreeClient)
	if err != nil {
		return nil, err
	}
	tree, ok := v.([]shareddto.CategoryResponse)
	if !ok {
		return nil, nil
	}
	return tree, nil
}

// SetCategoryTreeClient 设置客户端分类树缓存。
func (m *Manager) SetCategoryTreeClient(ctx context.Context, tree []shareddto.CategoryResponse) error {
	return m.cache.Set(ctx, KeyCategoryTreeClient, tree, TTLCategoryTree)
}

// InvalidateCategory 失效分类树缓存（client + open）。
func (m *Manager) InvalidateCategory(ctx context.Context) {
	_ = m.cache.Delete(ctx, KeyCategoryTreeClient)
	_ = m.cache.Delete(ctx, KeyOpenCategories)
}

// --- Live ---

// GetLiveListClient 获取客户端直播列表缓存。
func (m *Manager) GetLiveListClient(ctx context.Context) ([]shareddto.LiveChannelItem, error) {
	v, err := m.cache.Get(ctx, KeyLiveListClient)
	if err != nil {
		return nil, err
	}
	items, ok := v.([]shareddto.LiveChannelItem)
	if !ok {
		return nil, nil
	}
	return items, nil
}

// SetLiveListClient 设置客户端直播列表缓存。
func (m *Manager) SetLiveListClient(ctx context.Context, items []shareddto.LiveChannelItem) error {
	return m.cache.Set(ctx, KeyLiveListClient, items, TTLLiveList)
}

// InvalidateLive 失效直播列表缓存。
func (m *Manager) InvalidateLive(ctx context.Context) {
	_ = m.cache.Delete(ctx, KeyLiveListClient)
}

// --- Video (client) ---

// GetVideoListClient 获取客户端视频列表缓存。
func (m *Manager) GetVideoListClient(ctx context.Context, key string) (*VideoListCacheEntry, error) {
	v, err := m.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	entry, ok := v.(*VideoListCacheEntry)
	if !ok || entry == nil {
		return nil, nil
	}
	return entry, nil
}

// SetVideoListClient 设置客户端视频列表缓存。
func (m *Manager) SetVideoListClient(ctx context.Context, key string, entry *VideoListCacheEntry) error {
	return m.cache.Set(ctx, key, entry, TTLVideoList)
}

// InvalidateVideo 失效视频相关缓存（client 列表 + open 列表/详情/分类）。
func (m *Manager) InvalidateVideo(ctx context.Context, videoID int64) {
	for _, sort := range []string{"", "id_desc", "rating_desc", "view_count_desc", "created_at_desc"} {
		for _, page := range []int{1, 2} {
			for _, limit := range []int{12, 20, 24} {
				_ = m.cache.Delete(ctx, VideoListKey(0, sort, page, limit))
				_ = m.cache.Delete(ctx, OpenVideoListKey("default", page, limit))
				_ = m.cache.Delete(ctx, OpenVideoListKey("apple_cms", page, limit))
			}
		}
	}
	_ = m.cache.Delete(ctx, KeyOpenCategories)
	if videoID > 0 {
		_ = m.cache.Delete(ctx, OpenVideoDetailKey("default", videoID))
		_ = m.cache.Delete(ctx, OpenVideoDetailKey("apple_cms", videoID))
	}
}

// --- Settings ---

// GetSettingsAll 获取全部设置缓存。
func (m *Manager) GetSettingsAll(ctx context.Context) (map[string]model.SystemSettings, error) {
	v, err := m.cache.Get(ctx, KeySettingsAll)
	if err != nil {
		return nil, err
	}
	settings, ok := v.(map[string]model.SystemSettings)
	if !ok || settings == nil {
		return nil, nil
	}
	return settings, nil
}

// SetSettingsAll 设置全部设置缓存。
func (m *Manager) SetSettingsAll(ctx context.Context, settings map[string]model.SystemSettings) error {
	return m.cache.Set(ctx, KeySettingsAll, settings, TTLSettings)
}

// GetSettingsPublic 获取公开站点设置缓存。
func (m *Manager) GetSettingsPublic(ctx context.Context) (*admindto.PublicSiteResponse, error) {
	v, err := m.cache.Get(ctx, KeySettingsPublic)
	if err != nil {
		return nil, err
	}
	pub, ok := v.(*admindto.PublicSiteResponse)
	if !ok || pub == nil {
		return nil, nil
	}
	return pub, nil
}

// SetSettingsPublic 设置公开站点设置缓存。
func (m *Manager) SetSettingsPublic(ctx context.Context, pub *admindto.PublicSiteResponse) error {
	return m.cache.Set(ctx, KeySettingsPublic, pub, TTLSettings)
}

// InvalidateSettings 失效设置缓存。
func (m *Manager) InvalidateSettings(ctx context.Context) {
	_ = m.cache.Delete(ctx, KeySettingsAll)
	_ = m.cache.Delete(ctx, KeySettingsPublic)
}

// --- Open API ---

// GetOpenVideoList 获取 Open API 视频列表缓存。
func (m *Manager) GetOpenVideoList(ctx context.Context, key string) (any, error) {
	v, err := m.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// SetOpenVideoList 设置 Open API 视频列表缓存。
func (m *Manager) SetOpenVideoList(ctx context.Context, key string, value any) error {
	return m.cache.Set(ctx, key, value, TTLOpenVideoList)
}

// GetOpenVideoDetail 获取 Open API 视频详情缓存。
func (m *Manager) GetOpenVideoDetail(ctx context.Context, key string) (any, error) {
	v, err := m.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// SetOpenVideoDetail 设置 Open API 视频详情缓存。
func (m *Manager) SetOpenVideoDetail(ctx context.Context, key string, value any) error {
	return m.cache.Set(ctx, key, value, TTLOpenVideoDetail)
}

// GetOpenCategories 获取 Open API 分类树缓存。
func (m *Manager) GetOpenCategories(ctx context.Context) ([]shareddto.CategoryResponse, error) {
	v, err := m.cache.Get(ctx, KeyOpenCategories)
	if err != nil {
		return nil, err
	}
	tree, ok := v.([]shareddto.CategoryResponse)
	if !ok {
		return nil, nil
	}
	return tree, nil
}

// SetOpenCategories 设置 Open API 分类树缓存。
func (m *Manager) SetOpenCategories(ctx context.Context, tree []shareddto.CategoryResponse) error {
	return m.cache.Set(ctx, KeyOpenCategories, tree, TTLOpenCategories)
}

// Package cache provides business-specific cache management.
// Generic cache implementation lives in pkg/cache.
package cache

import (
	"context"
	"encoding/json"

	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	opendto "github.com/ilaziness/orange-tv/internal/dto/open"
	"github.com/ilaziness/orange-tv/internal/model"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
)

// VideoListCacheEntry 客户端视频列表缓存条目。
type VideoListCacheEntry struct {
	Items []clientdto.VideoListItem
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
func (m *Manager) GetCategoryTreeClient(ctx context.Context) ([]clientdto.CategoryResponse, error) {
	v, err := m.cache.Get(ctx, KeyCategoryTreeClient)
	if err != nil {
		return nil, err
	}
	tree, ok := v.([]clientdto.CategoryResponse)
	if !ok {
		return nil, nil
	}
	return tree, nil
}

// SetCategoryTreeClient 设置客户端分类树缓存。
func (m *Manager) SetCategoryTreeClient(ctx context.Context, tree []clientdto.CategoryResponse) error {
	return m.cache.Set(ctx, KeyCategoryTreeClient, tree, TTLCategoryTree)
}

// InvalidateCategory 失效分类树缓存（client + open）。
func (m *Manager) InvalidateCategory(ctx context.Context) {
	_ = m.cache.Delete(ctx, KeyCategoryTreeClient)
	_ = m.cache.Delete(ctx, KeyOpenCategories)
}

// --- LiveTV ---

// GetLiveTVListClient 获取客户端直播列表缓存。
// withStream=true 获取含流地址的缓存（app/tv/desktop），false 获取不含流地址的缓存（web）。
func (m *Manager) GetLiveTVListClient(ctx context.Context, withStream bool) ([]clientdto.LiveTVChannelItem, error) {
	v, err := m.cache.Get(ctx, LiveTVListKey(withStream))
	if err != nil {
		return nil, err
	}
	items, ok := v.([]clientdto.LiveTVChannelItem)
	if !ok {
		return nil, nil
	}
	return items, nil
}

// SetLiveTVListClient 设置客户端直播列表缓存。
// withStream=true 缓存含流地址条目（app/tv/desktop），false 缓存不含流地址条目（web）。
func (m *Manager) SetLiveTVListClient(ctx context.Context, items []clientdto.LiveTVChannelItem, withStream bool) error {
	return m.cache.Set(ctx, LiveTVListKey(withStream), items, TTLLiveTVList)
}

// InvalidateLiveTV 失效直播列表缓存（含流与不含流两个 key）。
func (m *Manager) InvalidateLiveTV(ctx context.Context) {
	_ = m.cache.Delete(ctx, LiveTVListKey(true))
	_ = m.cache.Delete(ctx, LiveTVListKey(false))
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

// InvalidateVideo 失效视频相关缓存（client 列表 + open 列表/分类）。
// SEO sitemap 依赖 5 分钟 TTL 自然过期，避免批量改片时反复清空并打爆重建。
func (m *Manager) InvalidateVideo(ctx context.Context) {
	for _, sort := range []string{"", "id_desc", "rating_desc", "view_count_desc", "created_at_desc"} {
		for _, page := range []int{1, 2} {
			for _, limit := range []int{12, 20, 24} {
				_ = m.cache.Delete(ctx, VideoListKey(0, 0, sort, page, limit))
				_ = m.cache.Delete(ctx, OpenVideoListKey(page, limit, "", ""))
			}
		}
	}
	_ = m.cache.Delete(ctx, KeyOpenCategories)
}

// --- Settings ---

// GetSettingsByGroup retrieves cached settings for a single group.
func (m *Manager) GetSettingsByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	v, err := m.cache.Get(ctx, SettingsGroupKey(group))
	if err != nil {
		return nil, err
	}
	settings, ok := v.(map[string]model.SystemSettings)
	if !ok || settings == nil {
		return nil, nil
	}
	return settings, nil
}

// SetSettingsByGroup caches settings for a single group.
func (m *Manager) SetSettingsByGroup(ctx context.Context, group string, settings map[string]model.SystemSettings) error {
	return m.cache.Set(ctx, SettingsGroupKey(group), settings, TTLSettings)
}

// InvalidateSettings invalidates all settings caches and SEO documents that depend on them.
func (m *Manager) InvalidateSettings(ctx context.Context) {
	for _, g := range constant.AllSettingGroups() {
		_ = m.cache.Delete(ctx, SettingsGroupKey(g))
	}
	m.InvalidateSEO(ctx)
}

// SEODocumentEntry is a JSON-friendly cache payload for SEO text/xml documents.
type SEODocumentEntry struct {
	Body        []byte `json:"body"`
	ContentType string `json:"content_type"`
	Status      int    `json:"status"`
}

// GetSEODocument retrieves a cached SEO document by key.
func (m *Manager) GetSEODocument(ctx context.Context, key string) (*SEODocumentEntry, error) {
	v, err := m.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case *SEODocumentEntry:
		return t, nil
	case SEODocumentEntry:
		return &t, nil
	case string:
		var entry SEODocumentEntry
		if json.Unmarshal([]byte(t), &entry) != nil {
			return nil, nil
		}
		return &entry, nil
	case []byte:
		var entry SEODocumentEntry
		if json.Unmarshal(t, &entry) != nil {
			return nil, nil
		}
		return &entry, nil
	default:
		return nil, nil
	}
}

// SetSEODocument caches an SEO document.
func (m *Manager) SetSEODocument(ctx context.Context, key string, entry *SEODocumentEntry) error {
	return m.cache.Set(ctx, key, entry, TTLSEODocument)
}

// GetSEOPageCursor retrieves a cached sitemap page afterID cursor.
func (m *Manager) GetSEOPageCursor(ctx context.Context, page int) (uint32, bool, error) {
	v, err := m.cache.Get(ctx, SEOPageCursorKey(page))
	if err != nil {
		return 0, false, err
	}
	switch t := v.(type) {
	case uint32:
		return t, true, nil
	case int:
		return uint32(t), true, nil
	case float64: // JSON numbers
		return uint32(t), true, nil
	case string:
		var id uint32
		if json.Unmarshal([]byte(t), &id) != nil {
			return 0, false, nil
		}
		return id, true, nil
	case []byte:
		var id uint32
		if json.Unmarshal(t, &id) != nil {
			return 0, false, nil
		}
		return id, true, nil
	default:
		return 0, false, nil
	}
}

// SetSEOPageCursor caches a sitemap page afterID cursor.
func (m *Manager) SetSEOPageCursor(ctx context.Context, page int, afterID uint32) error {
	return m.cache.Set(ctx, SEOPageCursorKey(page), afterID, TTLSEODocument)
}

// InvalidateSEO clears SEO document and cursor caches.
func (m *Manager) InvalidateSEO(ctx context.Context) {
	for _, name := range []string{"robots", "sitemap", "sitemap-pages", "llms"} {
		_ = m.cache.Delete(ctx, SEODocumentKey(name))
	}
	for page := 1; page <= MaxSEOSitemapVideoPages; page++ {
		_ = m.cache.Delete(ctx, SEOSitemapVideosKey(page))
		_ = m.cache.Delete(ctx, SEOPageCursorKey(page))
	}
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

// GetOpenCategories 获取 Open API 分类列表缓存。
func (m *Manager) GetOpenCategories(ctx context.Context) ([]opendto.CategoryItem, error) {
	v, err := m.cache.Get(ctx, KeyOpenCategories)
	if err != nil {
		return nil, err
	}
	items, ok := v.([]opendto.CategoryItem)
	if !ok {
		return nil, nil
	}
	return items, nil
}

// SetOpenCategories 设置 Open API 分类列表缓存。
func (m *Manager) SetOpenCategories(ctx context.Context, items []opendto.CategoryItem) error {
	return m.cache.Set(ctx, KeyOpenCategories, items, TTLOpenCategories)
}

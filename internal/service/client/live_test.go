package client

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
	"go.uber.org/zap"
)

type stubLiveRepo struct {
	items []model.LiveChannels
}

func (r *stubLiveRepo) List(ctx context.Context, f repository.LiveListFilter) ([]model.LiveChannels, int, error) {
	return nil, 0, nil
}

func (r *stubLiveRepo) ListAll(ctx context.Context) ([]model.LiveChannels, error) {
	return r.items, nil
}

func (r *stubLiveRepo) GetByID(ctx context.Context, id uint32) (*model.LiveChannels, error) {
	return nil, nil
}

func (r *stubLiveRepo) Create(ctx context.Context, m *model.LiveChannels) error {
	return nil
}

func (r *stubLiveRepo) BatchCreate(ctx context.Context, items []model.LiveChannels) error {
	return nil
}

func (r *stubLiveRepo) Update(ctx context.Context, m *model.LiveChannels) error {
	return nil
}

func (r *stubLiveRepo) Delete(ctx context.Context, id uint32) error {
	return nil
}

func (r *stubLiveRepo) DeleteByIDs(ctx context.Context, ids []uint32) error {
	return nil
}

func newTestLiveService(t *testing.T) *liveService {
	t.Helper()
	mgr := cache.NewManager(pkgcache.NewNopCache())
	return newLiveServiceWithCache(t, mgr)
}

func newLiveServiceWithCache(t *testing.T, mgr *cache.Manager) *liveService {
	t.Helper()
	svc := NewLiveService(&stubLiveRepo{
		items: []model.LiveChannels{
			{ID: 1, Name: "CCTV1", Category: "央视", StreamURL: "http://example.com/cctv1.m3u8", Status: 1, SortOrder: 1},
			{ID: 2, Name: "HBO", Category: "国际", StreamURL: "http://example.com/hbo.flv", Status: 1, SortOrder: 2},
			{ID: 3, Name: "Disabled", Category: "央视", StreamURL: "http://example.com/off.m3u8", Status: 0, SortOrder: 3},
		},
	}, mgr, zap.NewNop())
	return svc.(*liveService)
}

// mapCache 简单内存缓存实现，用于验证直播缓存按端分 key 的隔离行为。
type mapCache struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMapCache() *mapCache {
	return &mapCache{data: make(map[string][]byte)}
}

func (c *mapCache) Get(ctx context.Context, key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	raw, ok := c.data[key]
	if !ok {
		return nil, pkgcache.ErrCacheMiss
	}
	var items []clientdto.LiveChannelItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *mapCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.data[key] = data
	return nil
}

func (c *mapCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (c *mapCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.data[key]
	return ok, nil
}

func (c *mapCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string][]byte)
	return nil
}

func (c *mapCache) Close() error { return nil }

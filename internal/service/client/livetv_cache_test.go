package client

import (
	"context"
	"testing"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/clienttype"
	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLiveTVService_List_CacheIsolatedByStream 验证直播缓存按是否含流分 key，web 与 app/tv/desktop 互不污染。
func TestLiveTVService_List_CacheIsolatedByStream(t *testing.T) {
	mc := newMapCache()
	mgr := cache.NewManager(mc)
	svc := newLiveTVServiceWithCache(t, mgr)

	webCtx := clienttype.WithContext(context.Background(), constant.ClientTypeWeb)
	appCtx := clienttype.WithContext(context.Background(), constant.ClientTypeApp)

	// 1. web 首次请求：写 web 缓存（无流）
	items, _, err := svc.List(webCtx, &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	for _, it := range items {
		assert.Empty(t, it.StreamURL)
	}

	// 2. app 首次请求：写 stream 缓存（有流）
	items, _, err = svc.List(appCtx, &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	for _, it := range items {
		assert.NotEmpty(t, it.StreamURL)
	}

	// 3. web 再次请求：命中 web 缓存，仍无流（隔离）
	items, _, err = svc.List(webCtx, &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	for _, it := range items {
		assert.Empty(t, it.StreamURL, "web cache must never return stream_url")
	}

	// 4. app 再次请求：命中 stream 缓存，仍有流（隔离）
	items, _, err = svc.List(appCtx, &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	for _, it := range items {
		assert.NotEmpty(t, it.StreamURL, "app cache should keep stream_url")
	}

	// 5. 两个缓存 key 均已写入
	assert.NotEmpty(t, mc.data[cache.LiveTVListKey(false)], "web cache key should exist")
	assert.NotEmpty(t, mc.data[cache.LiveTVListKey(true)], "stream cache key should exist")

	// 6. web 缓存内容不含 stream_url，stream 缓存内容含 stream_url（防共享引用污染）
	assert.NotContains(t, string(mc.data[cache.LiveTVListKey(false)]), "stream_url", "web cache payload must not contain stream_url")
	assert.Contains(t, string(mc.data[cache.LiveTVListKey(true)]), "stream_url", "stream cache payload must contain stream_url")
}

// TestLiveCache_Invalidate_deletesBothKeys 验证 InvaliditeLive 同时删除 web 与 stream 两个缓存 key。
func TestLiveCache_Invalidate_deletesBothKeys(t *testing.T) {
	mc := newMapCache()
	mgr := cache.NewManager(mc)
	svc := newLiveTVServiceWithCache(t, mgr)

	appCtx := clienttype.WithContext(context.Background(), constant.ClientTypeApp)
	_, _, err := svc.List(appCtx, &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, mc.data[cache.LiveTVListKey(true)])

	webCtx := clienttype.WithContext(context.Background(), constant.ClientTypeWeb)
	_, _, err = svc.List(webCtx, &clientdto.LiveTVListRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, mc.data[cache.LiveTVListKey(false)])

	mgr.InvalidateLiveTV(context.Background())
	assert.Empty(t, mc.data[cache.LiveTVListKey(true)])
	assert.Empty(t, mc.data[cache.LiveTVListKey(false)])
}

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func setupTestMultiCache(t *testing.T) (*MultiCache, *miniredis.Miniredis) {
	t.Helper()

	// 创建 L1 内存缓存
	l1Config := MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1024 * 1024,
		BufferItems: 64,
	}
	l1, err := NewMemoryCache(l1Config)
	if err != nil {
		t.Fatalf("failed to create memory cache: %v", err)
	}

	// 创建 L2 Redis 缓存
	mr, redisConfig := setupTestRedis(t)
	cacheConfig := RedisCacheConfig{
		KeyPrefix:  "multi:",
		DefaultTTL: 3600,
	}
	l2, err := NewRedisCache(cacheConfig, redisConfig)
	if err != nil {
		t.Fatalf("failed to create redis cache: %v", err)
	}

	multiCache := NewMultiCache(l1, l2)

	t.Cleanup(func() {
		multiCache.Close()
	})

	return multiCache, mr
}

func TestMultiCache_GetFromL1(t *testing.T) {
	cache, _ := setupTestMultiCache(t)
	ctx := context.Background()

	// 先在 L1 设置值
	err := cache.l1.Set(ctx, "l1_key", "l1_value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set value in L1: %v", err)
	}

	// 从多级缓存获取，应该从 L1 命中
	value, err := cache.Get(ctx, "l1_key")
	if err != nil {
		t.Fatalf("failed to get value: %v", err)
	}

	if value != "l1_value" {
		t.Errorf("expected l1_value, got %v", value)
	}
}

func TestMultiCache_SetBothLevels(t *testing.T) {
	cache, _ := setupTestMultiCache(t)
	ctx := context.Background()

	// 设置值，应该同时写入 L1 和 L2
	err := cache.Set(ctx, "multi_key", "multi_value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	// 验证 L1 中有值
	value, err := cache.l1.Get(ctx, "multi_key")
	if err != nil {
		t.Errorf("expected value in L1, got error: %v", err)
	}
	if value != "multi_value" {
		t.Errorf("expected multi_value in L1, got %v", value)
	}

	// 验证 L2 中有值（JSON 格式）
	value, err = cache.l2.Get(ctx, "multi_key")
	if err != nil {
		t.Errorf("expected value in L2, got error: %v", err)
	}
	expected := `"multi_value"`
	if value != expected {
		t.Errorf("expected %s in L2, got %v", expected, value)
	}
}

func TestMultiCache_Delete(t *testing.T) {
	cache, _ := setupTestMultiCache(t)
	ctx := context.Background()

	// 设置值
	err := cache.Set(ctx, "delete_key", "delete_value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	// 删除值
	err = cache.Delete(ctx, "delete_key")
	if err != nil {
		t.Fatalf("failed to delete value: %v", err)
	}

	// 验证 L1 中已删除
	_, err = cache.l1.Get(ctx, "delete_key")
	if err != ErrCacheMiss {
		t.Error("expected key to be deleted from L1")
	}

	// 验证 L2 中已删除
	_, err = cache.l2.Get(ctx, "delete_key")
	if err != ErrCacheMiss {
		t.Error("expected key to be deleted from L2")
	}
}

func TestMultiCache_Exists(t *testing.T) {
	cache, _ := setupTestMultiCache(t)
	ctx := context.Background()

	// 设置值
	err := cache.Set(ctx, "exists_key", "exists_value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	// 检查存在性
	exists, err := cache.Exists(ctx, "exists_key")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}

	if !exists {
		t.Error("expected key to exist")
	}
}

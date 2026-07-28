package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, RedisOptions) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	t.Cleanup(func() {
		mr.Close()
	})

	port := 0
	if p, err := strconv.Atoi(mr.Port()); err == nil {
		port = p
	}

	opts := RedisOptions{
		Host:     mr.Host(),
		Port:     port,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	return mr, opts
}

func TestRedisCache_SetAndGet(t *testing.T) {
	_, opts := setupTestRedis(t)

	cacheConfig := RedisCacheConfig{
		KeyPrefix:  "test:",
		DefaultTTL: 3600,
	}

	cache, err := NewRedisCache(cacheConfig, opts)
	if err != nil {
		t.Fatalf("failed to create redis cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("Set and Get string value", func(t *testing.T) {
		err := cache.Set(ctx, "test_key", "test_value", 5*time.Minute)
		if err != nil {
			t.Fatalf("failed to set value: %v", err)
		}

		value, err := cache.Get(ctx, "test_key")
		if err != nil {
			t.Fatalf("failed to get value: %v", err)
		}

		// Redis 返回的是 JSON 字符串
		expected := `"test_value"`
		if value != expected {
			t.Errorf("expected %s, got %v", expected, value)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, err := cache.Get(ctx, "non_existent")
		if err != ErrCacheMiss {
			t.Errorf("expected ErrCacheMiss, got %v", err)
		}
	})
}

func TestRedisCache_Delete(t *testing.T) {
	_, opts := setupTestRedis(t)

	cacheConfig := RedisCacheConfig{
		KeyPrefix:  "test:",
		DefaultTTL: 3600,
	}

	cache, err := NewRedisCache(cacheConfig, opts)
	if err != nil {
		t.Fatalf("failed to create redis cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Set a value
	err = cache.Set(ctx, "delete_key", "delete_value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	// Delete the value
	err = cache.Delete(ctx, "delete_key")
	if err != nil {
		t.Fatalf("failed to delete value: %v", err)
	}

	// Verify it's deleted
	_, err = cache.Get(ctx, "delete_key")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after delete, got %v", err)
	}
}

func TestRedisCache_Exists(t *testing.T) {
	_, opts := setupTestRedis(t)

	cacheConfig := RedisCacheConfig{
		KeyPrefix:  "test:",
		DefaultTTL: 3600,
	}

	cache, err := NewRedisCache(cacheConfig, opts)
	if err != nil {
		t.Fatalf("failed to create redis cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("Key exists", func(t *testing.T) {
		err := cache.Set(ctx, "exists_key", "value", 5*time.Minute)
		if err != nil {
			t.Fatalf("failed to set value: %v", err)
		}

		exists, err := cache.Exists(ctx, "exists_key")
		if err != nil {
			t.Fatalf("failed to check existence: %v", err)
		}

		if !exists {
			t.Error("expected key to exist")
		}
	})

	t.Run("Key does not exist", func(t *testing.T) {
		exists, err := cache.Exists(ctx, "not_exists_key")
		if err != nil {
			t.Fatalf("failed to check existence: %v", err)
		}

		if exists {
			t.Error("expected key to not exist")
		}
	})
}

func TestRedisCache_KeyPrefix(t *testing.T) {
	mr, opts := setupTestRedis(t)

	cacheConfig := RedisCacheConfig{
		KeyPrefix:  "myprefix:",
		DefaultTTL: 3600,
	}

	cache, err := NewRedisCache(cacheConfig, opts)
	if err != nil {
		t.Fatalf("failed to create redis cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	err = cache.Set(ctx, "key1", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	// 验证键在 Redis 中带有前缀
	keys := mr.Keys()
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	expectedKey := "myprefix:key1"
	if keys[0] != expectedKey {
		t.Errorf("expected key %s, got %s", expectedKey, keys[0])
	}
}

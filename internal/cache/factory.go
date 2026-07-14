package cache

import (
	"fmt"

	"github.com/ilaziness/orange-tv/internal/config"
)

// CacheFactory 缓存工厂
type CacheFactory struct {
	config config.CacheConfig
	redis  config.RedisConfig
}

// NewCacheFactory 创建缓存工厂
func NewCacheFactory(cfg *config.Config) *CacheFactory {
	return &CacheFactory{
		config: cfg.Cache,
		redis:  cfg.Redis,
	}
}

// Create 创建缓存实例
func (f *CacheFactory) Create() (Cache, error) {
	if !f.config.Enabled {
		return NewNopCache(), nil
	}

	switch f.config.Driver {
	case "memory":
		memConfig := MemoryCacheConfig{
			NumCounters: f.config.Memory.NumCounters,
			MaxCost:     f.config.Memory.MaxCost,
			BufferItems: f.config.Memory.BufferItems,
		}
		return NewMemoryCache(memConfig)
	case "redis":
		redisCacheConfig := RedisCacheConfig{
			KeyPrefix:  f.config.Redis.KeyPrefix,
			DefaultTTL: f.config.Redis.DefaultTTL,
		}
		return NewRedisCache(redisCacheConfig, f.redis)
	case "multi":
		memConfig := MemoryCacheConfig{
			NumCounters: f.config.Memory.NumCounters,
			MaxCost:     f.config.Memory.MaxCost,
			BufferItems: f.config.Memory.BufferItems,
		}
		l1, err := NewMemoryCache(memConfig)
		if err != nil {
			return nil, err
		}

		redisCacheConfig := RedisCacheConfig{
			KeyPrefix:  f.config.Redis.KeyPrefix,
			DefaultTTL: f.config.Redis.DefaultTTL,
		}
		l2, err := NewRedisCache(redisCacheConfig, f.redis)
		if err != nil {
			return nil, err
		}

		return NewMultiCache(l1, l2), nil
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", f.config.Driver)
	}
}

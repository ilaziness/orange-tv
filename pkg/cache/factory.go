package cache

import (
	"fmt"
)

// CacheFactoryOptions 缓存工厂参数（与业务配置解耦）
type CacheFactoryOptions struct {
	Enabled    bool
	Driver     string
	Memory     MemoryCacheConfig
	RedisCache RedisCacheConfig // prefix, default TTL
	RedisConn  RedisOptions     // connection params
}

// CacheFactory 缓存工厂
type CacheFactory struct {
	opts CacheFactoryOptions
}

// NewCacheFactory 创建缓存工厂
func NewCacheFactory(opts CacheFactoryOptions) *CacheFactory {
	return &CacheFactory{opts: opts}
}

// Create 创建缓存实例
func (f *CacheFactory) Create() (Cache, error) {
	if !f.opts.Enabled {
		return NewNopCache(), nil
	}

	switch f.opts.Driver {
	case "memory":
		return NewMemoryCache(f.opts.Memory)
	case "redis":
		return NewRedisCache(f.opts.RedisCache, f.opts.RedisConn)
	case "multi":
		l1, err := NewMemoryCache(f.opts.Memory)
		if err != nil {
			return nil, err
		}

		l2, err := NewRedisCache(f.opts.RedisCache, f.opts.RedisConn)
		if err != nil {
			return nil, err
		}

		return NewMultiCache(l1, l2), nil
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", f.opts.Driver)
	}
}

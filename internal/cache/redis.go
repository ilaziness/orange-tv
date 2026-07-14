package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	KeyPrefix  string `mapstructure:"key_prefix"`
	DefaultTTL int    `mapstructure:"default_ttl"` // 默认过期时间（秒）
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewRedisCache 创建新的Redis缓存实例
func NewRedisCache(cacheConfig RedisCacheConfig, redisConfig config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:        redisConfig.Password,
		DB:              redisConfig.DB,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		ConnMaxIdleTime: time.Duration(redisConfig.IdleTimeout) * time.Second,
	})

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	prefix := "app:"
	if cacheConfig.KeyPrefix != "" {
		prefix = cacheConfig.KeyPrefix
	}

	defaultTTL := 1 * time.Hour
	if cacheConfig.DefaultTTL > 0 {
		defaultTTL = time.Duration(cacheConfig.DefaultTTL) * time.Second
	}

	return &RedisCache{
		client: client,
		prefix: prefix,
		ttl:    defaultTTL,
	}, nil
}

// buildKey 构建带前缀的键
func (c *RedisCache) buildKey(key string) string {
	return c.prefix + key
}

// Get 获取缓存值
func (c *RedisCache) Get(ctx context.Context, key string) (any, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	data, err := c.client.Get(ctx, c.buildKey(key)).Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, &CacheError{
			Code:    "REDIS_GET_ERROR",
			Message: "failed to get value from Redis",
			Err:     err,
		}
	}

	// 返回原始 JSON 数据，由调用方决定如何反序列化
	return data, nil
}

// Set 设置缓存值
func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	if ttl <= 0 {
		ttl = c.ttl
	}

	// 使用 JSON 序列化
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{
			Code:    "REDIS_MARSHAL_ERROR",
			Message: "failed to marshal value",
			Err:     err,
		}
	}

	err = c.client.Set(ctx, c.buildKey(key), data, ttl).Err()
	if err != nil {
		return &CacheError{
			Code:    "REDIS_SET_ERROR",
			Message: "failed to set value to Redis",
			Err:     err,
		}
	}

	return nil
}

// Delete 删除缓存值
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	err := c.client.Del(ctx, c.buildKey(key)).Err()
	if err != nil {
		return &CacheError{
			Code:    "REDIS_DELETE_ERROR",
			Message: "failed to delete value from Redis",
			Err:     err,
		}
	}

	return nil
}

// Exists 检查键是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	count, err := c.client.Exists(ctx, c.buildKey(key)).Result()
	if err != nil {
		return false, &CacheError{
			Code:    "REDIS_EXISTS_ERROR",
			Message: "failed to check existence in Redis",
			Err:     err,
		}
	}

	return count > 0, nil
}

// Clear 清空缓存（删除所有带前缀的键）
func (c *RedisCache) Clear(ctx context.Context) error {
	// 注意：生产环境慎用，这里仅用于测试
	pattern := c.prefix + "*"

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return &CacheError{
				Code:    "REDIS_CLEAR_ERROR",
				Message: "failed to scan keys in Redis",
				Err:     err,
			}
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return &CacheError{
					Code:    "REDIS_CLEAR_ERROR",
					Message: "failed to delete keys in Redis",
					Err:     err,
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// Close 关闭Redis连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

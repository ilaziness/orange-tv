package cache

import (
	"context"
	"time"
)

// Cache 定义缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (any, error)

	// Set 设置缓存值
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete 删除缓存值
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Clear 清空缓存
	Clear(ctx context.Context) error

	// Close 关闭缓存连接
	Close() error
}

// KeyGenerator 缓存键生成器
type KeyGenerator interface {
	// Generate 生成缓存键
	Generate(prefix string, parts ...string) string
}

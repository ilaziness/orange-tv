package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
)

// MemoryCacheConfig 内存缓存配置
type MemoryCacheConfig struct {
	NumCounters int64 `mapstructure:"num_counters"` // 预计键数量的 10 倍
	MaxCost     int64 `mapstructure:"max_cost"`     // 最大内存成本（字节）
	BufferItems int64 `mapstructure:"buffer_items"` // 默认 64
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	cache  *ristretto.Cache
	config MemoryCacheConfig
	mu     sync.RWMutex // 保护 cache 实例的并发访问
}

// NewMemoryCache 创建新的内存缓存实例
func NewMemoryCache(config MemoryCacheConfig) (*MemoryCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	return &MemoryCache{
		cache:  cache,
		config: config,
	}, nil
}

// Get 获取缓存值
func (c *MemoryCache) Get(ctx context.Context, key string) (any, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	value, found := c.cache.Get(key)
	if !found {
		return nil, ErrCacheMiss
	}

	return value, nil
}

// Set 设置缓存值
// 注意: Ristretto 是异步缓存，Wait() 会阻塞直到写入完成。
// 在高性能场景下，可以考虑移除 Wait() 调用以提升吞吐量，
// 但需要接受 Set 后立即 Get 可能返回缓存未命中的情况。
func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	// 计算成本：根据类型使用更准确的估算
	var cost int64
	switch v := value.(type) {
	case string:
		cost = int64(len(v))
	case []byte:
		cost = int64(len(v))
	default:
		// 对于复杂类型，使用字符串长度的粗略估算
		cost = int64(len(fmt.Sprintf("%v", value)))
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// SetWithTTL 是异步操作，会立即返回
	c.cache.SetWithTTL(key, value, cost, ttl)

	// Wait 确保值已被处理（保证强一致性，但会影响性能）
	// 对于高吞吐场景，可以移除此调用，接受短暂的最终一致性
	c.cache.Wait()

	return nil
}

// Delete 删除缓存值
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	c.cache.Del(key)
	return nil
}

// Exists 检查键是否存在
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	_, found := c.cache.Get(key)
	return found, nil
}

// Clear 清空缓存
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ristretto 没有直接的 Clear 方法，需要关闭并重新创建
	c.cache.Close()

	// 重新创建缓存
	newCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: c.config.NumCounters,
		MaxCost:     c.config.MaxCost,
		BufferItems: c.config.BufferItems,
	})
	if err != nil {
		return fmt.Errorf("failed to recreate cache: %w", err)
	}

	c.cache = newCache
	return nil
}

// Close 关闭缓存连接
func (c *MemoryCache) Close() error {
	c.cache.Close()
	return nil
}

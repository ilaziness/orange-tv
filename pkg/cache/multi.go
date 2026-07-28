package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MultiCache 多级缓存实现（L1: 内存, L2: Redis）
type MultiCache struct {
	l1 Cache          // 内存缓存
	l2 Cache          // Redis 缓存
	wg sync.WaitGroup // 跟踪异步回写 goroutine
}

// NewMultiCache 创建多级缓存
func NewMultiCache(l1, l2 Cache) *MultiCache {
	return &MultiCache{
		l1: l1,
		l2: l2,
	}
}

// Get 获取缓存值（先查L1，再查L2）
func (c *MultiCache) Get(ctx context.Context, key string) (any, error) {
	// 1. 尝试从 L1 获取
	if value, err := c.l1.Get(ctx, key); err == nil {
		return value, nil
	}

	// 2. L1 未命中，尝试从 L2 获取
	value, err := c.l2.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// 3. 回写到 L1（异步）
	c.wg.Add(1)
	go func() { //nolint:gosec // intentional fire-and-forget cache warm-up without request context
		defer c.wg.Done()
		// 使用带超时的 context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.l1.Set(ctx, key, value, 5*time.Minute)
	}()

	return value, nil
}

// Set 设置缓存值（同时写入 L1 和 L2）
func (c *MultiCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err1 := c.l1.Set(ctx, key, value, ttl)
	err2 := c.l2.Set(ctx, key, value, ttl)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to set cache: l1=%v, l2=%v", err1, err2)
	}
	return nil
}

// Delete 删除缓存值
func (c *MultiCache) Delete(ctx context.Context, key string) error {
	err1 := c.l1.Delete(ctx, key)
	err2 := c.l2.Delete(ctx, key)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to delete cache: l1=%v, l2=%v", err1, err2)
	}
	return nil
}

// Exists 检查键是否存在
func (c *MultiCache) Exists(ctx context.Context, key string) (bool, error) {
	// 先检查 L1
	exists, err := c.l1.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// 再检查 L2
	return c.l2.Exists(ctx, key)
}

// Clear 清空缓存
func (c *MultiCache) Clear(ctx context.Context) error {
	err1 := c.l1.Clear(ctx)
	err2 := c.l2.Clear(ctx)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to clear cache: l1=%v, l2=%v", err1, err2)
	}
	return nil
}

// Close 关闭缓存
func (c *MultiCache) Close() error {
	// 先等待所有异步操作完成
	c.wg.Wait()

	err1 := c.l1.Close()
	err2 := c.l2.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

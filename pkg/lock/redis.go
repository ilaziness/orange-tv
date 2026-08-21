package lock

import (
	"context"
	"errors"
	"fmt"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

// redisLocker 基于 bsm/redislock 实现 Locker。
// 利用 token CAS 保证只有持有者能释放，避免误释放他人锁。
type redisLocker struct {
	client *redislock.Client
}

// NewRedisLocker 基于 redis client 创建 Locker。传入 nil 会 panic，
// 调用方应通过 factory 降级为内存锁，避免直接传入 nil。
func NewRedisLocker(c redis.UniversalClient) Locker {
	return &redisLocker{client: redislock.New(c)}
}

// Lock 使用 NoRetry 策略抢锁：失败立即返回，不阻塞等待。
func (l *redisLocker) Lock(ctx context.Context, key string, opts ...Option) (Lock, error) {
	o := applyOptions(opts)
	lock, err := l.client.Obtain(ctx, key, o.TTL, &redislock.Options{
		RetryStrategy: redislock.NoRetry(),
	})
	if errors.Is(err, redislock.ErrNotObtained) {
		return nil, ErrLockNotHeld
	}
	if err != nil {
		return nil, fmt.Errorf("lock: obtain redis lock: %w", err)
	}
	return &redisLock{lock: lock}, nil
}

// redisLock 包装 redislock.Lock，实现 Lock 接口。
type redisLock struct {
	lock *redislock.Lock
}

// Release 释放锁。redislock 内部用 token CAS，锁已丢失时返回 ErrLockNotHeld，
// 这里转换为 nil（幂等释放），便于调用方在 defer 中无条件 Release。
func (l *redisLock) Release(ctx context.Context) error {
	if l.lock == nil {
		return nil
	}
	err := l.lock.Release(ctx)
	if err != nil && !errors.Is(err, redislock.ErrLockNotHeld) {
		return fmt.Errorf("lock: release redis lock: %w", err)
	}
	return nil
}

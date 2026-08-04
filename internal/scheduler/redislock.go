// 本文件基于 github.com/bsm/redislock 实现 LockProvider，
// 用于多机部署时保证只有一个实例运行 cron 调度。
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

// 锁相关常量。
const (
	// schedulerLockKey 调度器单实例锁的 Redis key。
	schedulerLockKey = "orange-tv:scheduler:lock"
	// schedulerLockTTL 锁的过期时间，需大于心跳续期间隔以留出余量。
	schedulerLockTTL = 30 * time.Second
)

// redisLockProvider 基于 bsm/redislock 实现 LockProvider。
// 利用 token-based Lua 脚本保证释放与续期只能由持有者执行，
// 避免 stale lock 被误释放/误续期。
type redisLockProvider struct {
	locker *redislock.Client
	// lock 持有当前获取到的锁实例，仅在 Acquire 成功后非 nil。
	// Acquire 在 Start 主 goroutine 中调用，Renew 在心跳 goroutine 中调用，
	// Release 在 releaseLock（已等待心跳退出）后调用，三者串行不存在并发，无需加锁保护。
	lock *redislock.Lock
}

// NewRedisLockProvider 基于 redis client 创建调度器分布式锁实现。
// client 为 nil 时应返回 nil（由调用方处理），这里不做防御以保持显式。
func NewRedisLockProvider(client redis.UniversalClient) LockProvider {
	if client == nil {
		return nil
	}
	return &redisLockProvider{locker: redislock.New(client)}
}

// AcquireSchedulerLock 尝试获取调度器锁。
// 使用 NoRetry 策略：多机场景下抢锁失败应立即返回，不阻塞等待。
func (p *redisLockProvider) AcquireSchedulerLock(ctx context.Context) (bool, error) {
	lock, err := p.locker.Obtain(ctx, schedulerLockKey, schedulerLockTTL, &redislock.Options{
		RetryStrategy: redislock.NoRetry(),
	})
	if errors.Is(err, redislock.ErrNotObtained) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("obtain scheduler lock: %w", err)
	}
	p.lock = lock
	return true, nil
}

// RenewSchedulerLock 续期调度器锁，重置 TTL。
// redislock 的 Refresh 内部用 token CAS，只有持有者能续期。
// 锁已丢失（过期或被抢占）时返回 ErrLockNotHeld，调用方应据此停止受保护的工作。
func (p *redisLockProvider) RenewSchedulerLock(ctx context.Context) error {
	if p.lock == nil {
		return ErrLockNotHeld
	}
	if err := p.lock.Refresh(ctx, schedulerLockTTL, nil); err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			return ErrLockNotHeld
		}
		return fmt.Errorf("renew scheduler lock: %w", err)
	}
	return nil
}

// ReleaseSchedulerLock 释放调度器锁。
// redislock 的 Release 内部用 token CAS，只有持有者能释放，避免误释放他人锁。
func (p *redisLockProvider) ReleaseSchedulerLock(ctx context.Context) error {
	if p.lock == nil {
		return nil
	}
	if err := p.lock.Release(ctx); err != nil && !errors.Is(err, redislock.ErrLockNotHeld) {
		return fmt.Errorf("release scheduler lock: %w", err)
	}
	p.lock = nil
	return nil
}

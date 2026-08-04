package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// setupTestRedisClient 启动 miniredis 并返回连接它的 redis client。
func setupTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() { _ = client.Close() })

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping miniredis: %v", err)
	}
	return client
}

func TestRedisLockProvider_AcquireAndRelease(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	client := setupTestRedisClient(t)
	p := newRedisLockProvider(client)

	// 首次获取成功
	ok, err := p.AcquireSchedulerLock(ctx)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	if !ok {
		t.Error("expected first Acquire to succeed")
	}

	// 同一 provider 再次获取应失败（锁已被自己持有）
	ok2, err := p.AcquireSchedulerLock(ctx)
	if err != nil {
		t.Fatalf("second Acquire unexpected error: %v", err)
	}
	if ok2 {
		t.Error("expected second Acquire to fail, lock already held")
	}

	// 释放
	if err := p.ReleaseSchedulerLock(ctx); err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// 释放后可再次获取
	ok3, err := p.AcquireSchedulerLock(ctx)
	if err != nil {
		t.Fatalf("Acquire after Release failed: %v", err)
	}
	if !ok3 {
		t.Error("expected Acquire to succeed after Release")
	}
	_ = p.ReleaseSchedulerLock(ctx)
}

func TestRedisLockProvider_MutualExclusion(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	client := setupTestRedisClient(t)

	// 两个独立 provider 共享同一 Redis，模拟两个实例
	p1 := newRedisLockProvider(client)
	p2 := newRedisLockProvider(client)

	// p1 获取锁
	ok1, err := p1.AcquireSchedulerLock(ctx)
	if err != nil || !ok1 {
		t.Fatalf("p1 Acquire failed: ok=%v err=%v", ok1, err)
	}

	// p2 应获取失败
	ok2, err := p2.AcquireSchedulerLock(ctx)
	if err != nil {
		t.Fatalf("p2 Acquire unexpected error: %v", err)
	}
	if ok2 {
		t.Error("expected p2 Acquire to fail while p1 holds lock")
	}

	// p1 释放后 p2 可获取
	if err := p1.ReleaseSchedulerLock(ctx); err != nil {
		t.Fatalf("p1 Release failed: %v", err)
	}
	ok3, err := p2.AcquireSchedulerLock(ctx)
	if err != nil || !ok3 {
		t.Fatalf("p2 Acquire after p1 Release failed: ok=%v err=%v", ok3, err)
	}
	_ = p2.ReleaseSchedulerLock(ctx)
}

func TestRedisLockProvider_Renew(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	client := setupTestRedisClient(t)
	p := newRedisLockProvider(client)

	if _, err := p.AcquireSchedulerLock(ctx); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	// 续期应成功
	if err := p.RenewSchedulerLock(ctx); err != nil {
		t.Errorf("Renew failed: %v", err)
	}

	_ = p.ReleaseSchedulerLock(ctx)
}

func TestRedisLockProvider_RenewWithoutAcquire(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	client := setupTestRedisClient(t)
	p := newRedisLockProvider(client)

	// 未获取锁时续期应返回 ErrLockNotHeld
	err := p.RenewSchedulerLock(ctx)
	if err == nil {
		t.Fatal("expected Renew to error when lock not held")
	}
	if !errors.Is(err, ErrLockNotHeld) {
		t.Errorf("expected ErrLockNotHeld, got %v", err)
	}
}

// TestNewScheduler_NilClientUsesNoopLock 验证 NewScheduler 在 redisClient 为 nil 时
// 使用 NoopLockProvider（空锁），不 panic、不启动心跳。
func TestNewScheduler_NilClientUsesNoopLock(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	sched := NewScheduler(nil)
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start with nil redisClient failed: %v", err)
	}
	if !sched.cronStarted.Load() {
		t.Error("expected cronStarted to be true")
	}
	if !sched.lockOwned {
		t.Error("expected lockOwned to be true (NoopLockProvider always acquires)")
	}
	if sched.heartbeatCancel != nil {
		t.Error("expected no heartbeat goroutine with NoopLockProvider")
	}
	if err := sched.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestRedisLockProvider_ReleaseWithoutAcquire(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	client := setupTestRedisClient(t)
	p := newRedisLockProvider(client)

	// 未获取锁时 Release 应安全返回（无 panic）
	if err := p.ReleaseSchedulerLock(ctx); err != nil {
		t.Errorf("Release without Acquire should not error: %v", err)
	}
}

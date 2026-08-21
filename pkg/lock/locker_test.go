package lock

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis 启动 miniredis 并返回可用的 redis.UniversalClient。
func setupTestRedis(t *testing.T) (redis.UniversalClient, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	port := 0
	if p, err := strconv.Atoi(mr.Port()); err == nil {
		port = p
	}
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{mr.Host() + ":" + strconv.Itoa(port)},
	})
	t.Cleanup(func() { _ = client.Close() })
	return client, mr
}

// errIsLockNotHeld 判断错误是否为 ErrLockNotHeld。
func errIsLockNotHeld(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrLockNotHeld) {
		t.Fatalf("expected ErrLockNotHeld, got %v", err)
	}
}

func TestMemoryLocker_LockAndRelease(t *testing.T) {
	l := NewMemoryLocker()
	ctx := context.Background()

	lock, err := l.Lock(ctx, "k1")
	if err != nil {
		t.Fatalf("first lock failed: %v", err)
	}
	if lock == nil {
		t.Fatal("first lock returned nil")
	}

	// 同 key 第二次抢锁应失败。
	_, err = l.Lock(ctx, "k1")
	errIsLockNotHeld(t, err)

	// 释放后应可重新获取。
	if err := lock.Release(ctx); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	lock2, err := l.Lock(ctx, "k1")
	if err != nil {
		t.Fatalf("re-acquire after release failed: %v", err)
	}
	_ = lock2.Release(ctx)
}

func TestMemoryLocker_ReleaseIsIdempotent(t *testing.T) {
	l := NewMemoryLocker()
	ctx := context.Background()

	lock, _ := l.Lock(ctx, "k")
	// 多次释放应幂等返回 nil。
	for i := 0; i < 3; i++ {
		if err := lock.Release(ctx); err != nil {
			t.Fatalf("release %d failed: %v", i, err)
		}
	}
}

func TestMemoryLocker_TTLExpiration(t *testing.T) {
	l := NewMemoryLocker()
	ctx := context.Background()

	// 设置极短 TTL，过期后应可被新调用获取。
	_, err := l.Lock(ctx, "k", WithTTL(50*time.Millisecond))
	if err != nil {
		t.Fatalf("first lock failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	lock, err := l.Lock(ctx, "k", WithTTL(time.Minute))
	if err != nil {
		t.Fatalf("re-acquire after TTL expiration failed: %v", err)
	}
	_ = lock.Release(ctx)
}

func TestMemoryLocker_ConcurrentOnlyOneWinner(t *testing.T) {
	l := NewMemoryLocker()
	ctx := context.Background()

	const n = 100
	var success int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			lock, err := l.Lock(ctx, "shared")
			if err != nil {
				return
			}
			atomic.AddInt32(&success, 1)
			// 故意不立即释放，让其他 goroutine 全部失败。
			time.Sleep(50 * time.Millisecond)
			_ = lock.Release(ctx)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&success); got != 1 {
		t.Fatalf("expected exactly 1 winner, got %d", got)
	}
}

func TestMemoryLocker_DifferentKeysIndependent(t *testing.T) {
	l := NewMemoryLocker()
	ctx := context.Background()

	_, err := l.Lock(ctx, "a")
	if err != nil {
		t.Fatalf("lock a failed: %v", err)
	}
	// 不同 key 不互斥。
	_, err = l.Lock(ctx, "b")
	if err != nil {
		t.Fatalf("lock b failed: %v", err)
	}
}

func TestRedisLocker_LockAndRelease(t *testing.T) {
	client, _ := setupTestRedis(t)
	l := NewRedisLocker(client)
	ctx := context.Background()

	lock, err := l.Lock(ctx, "k1")
	if err != nil {
		t.Fatalf("first lock failed: %v", err)
	}
	// 同 key 抢锁应失败。
	_, err = l.Lock(ctx, "k1")
	errIsLockNotHeld(t, err)

	if err := lock.Release(ctx); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	// 释放后可重新获取。
	lock2, err := l.Lock(ctx, "k1")
	if err != nil {
		t.Fatalf("re-acquire after release failed: %v", err)
	}
	_ = lock2.Release(ctx)
}

func TestRedisLocker_ConcurrentOnlyOneWinner(t *testing.T) {
	client, _ := setupTestRedis(t)
	l := NewRedisLocker(client)
	ctx := context.Background()

	const n = 50
	var success int32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			lock, err := l.Lock(ctx, "shared")
			if err != nil {
				return
			}
			atomic.AddInt32(&success, 1)
			time.Sleep(50 * time.Millisecond)
			_ = lock.Release(ctx)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&success); got != 1 {
		t.Fatalf("expected exactly 1 winner, got %d", got)
	}
}

func TestFactory_CreateMemoryWhenRedisNil(t *testing.T) {
	f := NewLockerFactory(LockerFactoryOptions{RedisClient: nil})
	l := f.Create()
	if _, ok := l.(*memoryLocker); !ok {
		t.Fatalf("expected *memoryLocker, got %T", l)
	}
}

// TestFactory_CreateMemoryWhenRedisTypedNil 覆盖生产场景：
// 调用方把 (*redis.Client)(nil) 赋给 redis.UniversalClient 接口字段时，
// 接口值非 nil（type=*redis.Client, value=nil），简单 == nil 判断无法识别。
// factory 必须正确降级为 memoryLocker，否则会创建出持有 nil client 的
// redisLocker，运行时 nil 解引用 panic。
func TestFactory_CreateMemoryWhenRedisTypedNil(t *testing.T) {
	var nilClient *redis.Client
	f := NewLockerFactory(LockerFactoryOptions{RedisClient: nilClient})
	l := f.Create()
	if _, ok := l.(*memoryLocker); !ok {
		t.Fatalf("expected *memoryLocker for typed-nil client, got %T", l)
	}
}

func TestFactory_CreateRedisWhenClientProvided(t *testing.T) {
	client, _ := setupTestRedis(t)
	f := NewLockerFactory(LockerFactoryOptions{RedisClient: client})
	l := f.Create()
	if _, ok := l.(*redisLocker); !ok {
		t.Fatalf("expected *redisLocker, got %T", l)
	}
}

func TestWithTTLDefault(t *testing.T) {
	o := applyOptions(nil)
	if o.TTL != defaultTTL {
		t.Fatalf("expected default TTL %v, got %v", defaultTTL, o.TTL)
	}

	o2 := applyOptions([]Option{WithTTL(0)})
	if o2.TTL != defaultTTL {
		t.Fatalf("expected 0 TTL to fall back to default, got %v", o2.TTL)
	}

	o3 := applyOptions([]Option{WithTTL(10 * time.Second)})
	if o3.TTL != 10*time.Second {
		t.Fatalf("expected 10s TTL, got %v", o3.TTL)
	}
}

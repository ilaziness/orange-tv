package captcha

import (
	"context"
	"errors"
	"testing"
	"time"

	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
)

// 外部 cache store 适配器：基于内存 cache 的端到端验证。
func TestCacheStoreAdapter(t *testing.T) {
	mem, err := pkgcache.NewMemoryCache(pkgcache.MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatalf("create memory cache: %v", err)
	}
	defer mem.Close()

	store := NewCacheStore(mem, time.Minute)
	c := New(WithStore(store))
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	// ristretto 为异步写入，等待 Set 生效后再读取。
	waitForSettle()
	// 通过适配器 Get 取答案（clear=false）
	answer := store.Get(img.ID, false)
	if answer == "" {
		t.Fatal("cache store should hold answer")
	}
	if err := c.Verify(context.Background(), img.ID, answer); err != nil {
		t.Fatalf("Verify via cache store failed: %v", err)
	}
	// 一次性：再次校验 NotFound
	if err := c.Verify(context.Background(), img.ID, answer); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after one-time verify, got %v", err)
	}
}

// WithCacheStore 便捷 Option：等价于 WithStore(NewCacheStore(c, ttl))，
// 且 ttl <= 0 时使用默认有效期，避免调用方感知 base64Captcha.Store 类型。
func TestWithCacheStore(t *testing.T) {
	mem, err := pkgcache.NewMemoryCache(pkgcache.MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatalf("create memory cache: %v", err)
	}
	defer mem.Close()

	// ttl=0 走默认有效期
	c := New(WithCacheStore(mem, 0))
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	waitForSettle()
	pc := c.(*pngCaptcha)
	answer := pc.captcha.Store.Get(img.ID, false)
	if answer == "" {
		t.Fatal("WithCacheStore should hold answer")
	}
	if err := c.Verify(context.Background(), img.ID, answer); err != nil {
		t.Fatalf("Verify via WithCacheStore failed: %v", err)
	}
}

// 外部 cache store：答案不匹配返回 ErrNotMatched 并作废。
func TestCacheStoreAdapterNotMatched(t *testing.T) {
	mem, err := pkgcache.NewMemoryCache(pkgcache.MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		t.Fatalf("create memory cache: %v", err)
	}
	defer mem.Close()

	c := New(WithStore(NewCacheStore(mem, time.Minute)))
	img, err := c.Generate(context.Background(), "login")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	waitForSettle()
	if err := c.Verify(context.Background(), img.ID, "wrong"); !errors.Is(err, ErrNotMatched) {
		t.Fatalf("expected ErrNotMatched, got %v", err)
	}
	waitForSettle()
	if err := c.Verify(context.Background(), img.ID, "wrong"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after one-time, got %v", err)
	}
}

// waitForSettle 等待 ristretto 异步写入落盘，避免测试时序 flaky。
func waitForSettle() {
	time.Sleep(20 * time.Millisecond)
}

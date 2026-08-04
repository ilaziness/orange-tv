package event_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/pkg/event"
)

// resetDefault 在测试前后清理默认总线，避免污染其他测试。
func resetDefault() {
	event.SetDefault(nil)
}

func TestDefault_NotSet(t *testing.T) {
	resetDefault()
	defer resetDefault()

	if event.Default() != nil {
		t.Fatal("expected nil default bus after reset")
	}

	// Publish 安全降级，返回 nil 不 panic
	if err := event.Publish(context.Background(), &event.Event{Name: "x"}); err != nil {
		t.Fatalf("Publish should be no-op when default not set, got %v", err)
	}

	// Unsubscribe 安全降级
	if err := event.Unsubscribe("x", func(ctx context.Context, e *event.Event) error { return nil }); err != nil {
		t.Fatalf("Unsubscribe should be no-op when default not set, got %v", err)
	}

	// Subscribe 返回 error
	if err := event.Subscribe("x", func(ctx context.Context, e *event.Event) error { return nil }); err == nil {
		t.Fatal("Subscribe should return error when default not set")
	}
	if err := event.SubscribeWildcard("x.*", func(ctx context.Context, e *event.Event) error { return nil }); err == nil {
		t.Fatal("SubscribeWildcard should return error when default not set")
	}
}

func TestDefault_PublishAndSubscribe(t *testing.T) {
	resetDefault()
	defer resetDefault()

	bus := event.NewEventBus()
	defer bus.Close()
	event.SetDefault(bus)

	var got atomic.Int32
	_ = event.Subscribe("test.default", func(ctx context.Context, e *event.Event) error {
		if s, ok := e.Payload.(string); ok && s == "hello" {
			got.Add(1)
		}
		return nil
	})

	if err := event.Publish(context.Background(), &event.Event{
		Name:    "test.default",
		Payload: "hello",
	}); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if got.Load() != 1 {
		t.Fatalf("expected handler called once, got %d", got.Load())
	}
}

func TestDefault_OverwriteAndReset(t *testing.T) {
	resetDefault()
	defer resetDefault()

	bus1 := event.NewEventBus()
	bus2 := event.NewEventBus()
	defer bus1.Close()
	defer bus2.Close()

	event.SetDefault(bus1)
	if event.Default() != bus1 {
		t.Fatal("expected bus1 as default")
	}

	// 覆盖
	event.SetDefault(bus2)
	if event.Default() != bus2 {
		t.Fatal("expected bus2 as default after overwrite")
	}

	// 重置
	event.SetDefault(nil)
	if event.Default() != nil {
		t.Fatal("expected nil after reset")
	}
}

func TestDefault_ConcurrentAccess(t *testing.T) {
	resetDefault()
	defer resetDefault()

	bus := event.NewEventBus()
	defer bus.Close()

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n * 2)

	// 并发 SetDefault / Default 读取
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			event.SetDefault(bus)
		}()
	}
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_ = event.Default()
		}()
	}

	wg.Wait()
	// 最后稳定为 bus
	event.SetDefault(bus)
	if event.Default() != bus {
		t.Fatal("expected bus as default after concurrent access")
	}

	// 并发 Publish
	var got atomic.Int32
	_ = event.Subscribe("concurrent", func(ctx context.Context, e *event.Event) error {
		got.Add(1)
		return nil
	})

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_ = event.Publish(context.Background(), &event.Event{Name: "concurrent"})
		}()
	}
	wg.Wait()

	// 同步处理，应全部完成
	time.Sleep(50 * time.Millisecond)
	if got.Load() != int32(n) {
		t.Fatalf("expected %d events, got %d", n, got.Load())
	}
}

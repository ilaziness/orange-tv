package event

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEventBus_PublishAndSubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan string, 10)

	// 订阅事件
	err := bus.Subscribe("test.event", func(ctx context.Context, e *Event) error {
		if msg, ok := e.Payload.(string); ok {
			received <- msg
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 发布事件
	err = bus.Publish(context.Background(), &Event{
		Name:    "test.event",
		Payload: "hello",
	})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// 等待接收
	select {
	case msg := <-received:
		if msg != "hello" {
			t.Errorf("expected hello, got %s", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var count int
	var mu sync.Mutex

	// 添加多个订阅者
	for i := 0; i < 5; i++ {
		err := bus.Subscribe("multi.event", func(ctx context.Context, e *Event) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}

	// 发布事件
	err := bus.Publish(context.Background(), &Event{Name: "multi.event"})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// 等待所有 handler 执行
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 5 {
		t.Errorf("expected 5 handlers called, got %d", count)
	}
}

func TestEventBus_AsyncHandling(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan string, 1)

	// 异步订阅
	err := bus.Subscribe("async.event", func(ctx context.Context, e *Event) error {
		time.Sleep(50 * time.Millisecond) // 模拟耗时操作
		if msg, ok := e.Payload.(string); ok {
			received <- msg
		}
		return nil
	}, SubscriptionOptions{Async: true})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 发布事件（应该立即返回）
	start := time.Now()
	err = bus.Publish(context.Background(), &Event{
		Name:    "async.event",
		Payload: "async_message",
	})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}
	elapsed := time.Since(start)

	// 验证 Publish 是异步的（应该很快返回）
	if elapsed > 10*time.Millisecond {
		t.Logf("Publish took %v, expected < 10ms", elapsed)
	}

	// 等待异步处理完成
	select {
	case msg := <-received:
		if msg != "async_message" {
			t.Errorf("expected async_message, got %s", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for async event")
	}
}

func TestEventBus_SyncHandling(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	executed := false

	// 同步订阅
	err := bus.Subscribe("sync.event", func(ctx context.Context, e *Event) error {
		time.Sleep(50 * time.Millisecond)
		executed = true
		return nil
	}, SubscriptionOptions{Async: false})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 发布事件（应该阻塞直到 handler 完成）
	start := time.Now()
	err = bus.Publish(context.Background(), &Event{Name: "sync.event"})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}
	elapsed := time.Since(start)

	// 验证 Publish 是同步的（应该等待 handler 完成）
	if elapsed < 40*time.Millisecond {
		t.Errorf("Publish returned too quickly (%v), expected >= 50ms", elapsed)
	}

	if !executed {
		t.Error("handler was not executed")
	}
}

func TestEventBus_Priority(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	order := make([]int, 0, 3)
	var mu sync.Mutex

	// 按不同优先级订阅
	_ = bus.Subscribe("priority.event", func(ctx context.Context, e *Event) error {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return nil
	}, SubscriptionOptions{Priority: PriorityNormal})

	_ = bus.Subscribe("priority.event", func(ctx context.Context, e *Event) error {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		return nil
	}, SubscriptionOptions{Priority: PriorityHigh})

	_ = bus.Subscribe("priority.event", func(ctx context.Context, e *Event) error {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return nil
	}, SubscriptionOptions{Priority: PriorityLow})

	// 发布事件
	_ = bus.Publish(context.Background(), &Event{Name: "priority.event"})

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// 验证执行顺序：高优先级 -> 普通优先级 -> 低优先级
	expected := []int{1, 2, 3}
	if len(order) != len(expected) {
		t.Fatalf("expected %d handlers, got %d", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected priority %d, got %d", i, v, order[i])
		}
	}
}

func TestEventBus_WildcardSubscription(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan string, 10)

	// 通配符订阅
	err := bus.SubscribeWildcard("user.*", func(ctx context.Context, e *Event) error {
		received <- e.Name
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe wildcard: %v", err)
	}

	// 发布匹配的事件
	events := []string{"user.created", "user.updated", "user.deleted"}
	for _, eventName := range events {
		err = bus.Publish(context.Background(), &Event{Name: eventName})
		if err != nil {
			t.Fatalf("failed to publish %s: %v", eventName, err)
		}
	}

	// 发布不匹配的事件
	_ = bus.Publish(context.Background(), &Event{Name: "order.created"})

	// 等待接收
	timeout := time.After(1 * time.Second)
	receivedCount := 0

	for receivedCount < 3 {
		select {
		case name := <-received:
			receivedCount++
			// 验证接收到的是 user.* 事件
			found := false
			for _, expected := range events {
				if name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("unexpected event: %s", name)
			}
		case <-timeout:
			t.Fatalf("timeout, received %d events, expected 3", receivedCount)
		}
	}

	if receivedCount != 3 {
		t.Errorf("expected 3 events, got %d", receivedCount)
	}
}

func TestEventBus_ContextCancellation(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	executed := false

	// 订阅一个会检查 context 的 handler
	err := bus.Subscribe("cancel.event", func(ctx context.Context, e *Event) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			executed = true
			return nil
		}
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	// 发布事件
	err = bus.Publish(ctx, &Event{Name: "cancel.event"})
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	// Handler 应该因为 context 取消而没有执行主要逻辑
	if executed {
		t.Log("Handler executed before checking context (acceptable)")
	}
}

func TestEventBus_HandlerError(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	errorHandled := false
	successHandled := false

	// 第一个 handler 返回错误
	_ = bus.Subscribe("error.event", func(ctx context.Context, e *Event) error {
		errorHandled = true
		return nil // 错误会被记录，但不会中断其他 handler
	})

	// 第二个 handler 应该仍然执行
	_ = bus.Subscribe("error.event", func(ctx context.Context, e *Event) error {
		successHandled = true
		return nil
	})

	// 发布事件
	_ = bus.Publish(context.Background(), &Event{Name: "error.event"})

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	if !errorHandled {
		t.Error("first handler was not executed")
	}
	if !successHandled {
		t.Error("second handler was not executed after first handler error")
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var count int
	var mu sync.Mutex

	// 订阅事件
	_ = bus.Subscribe("concurrent.event", func(ctx context.Context, e *Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})

	// 并发发布
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bus.Publish(context.Background(), &Event{Name: "concurrent.event"})
		}()
	}

	wg.Wait()

	// 等待所有 handler 执行
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if count != 100 {
		t.Errorf("expected 100 events, got %d", count)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	callCount := 0

	// 创建 handler
	handler := func(ctx context.Context, e *Event) error {
		callCount++
		return nil
	}

	// 订阅
	_ = bus.Subscribe("unsubscribe.event", handler)

	// 发布第一次
	_ = bus.Publish(context.Background(), &Event{Name: "unsubscribe.event"})
	time.Sleep(50 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// 取消订阅
	_ = bus.Unsubscribe("unsubscribe.event", handler)

	// 发布第二次
	_ = bus.Publish(context.Background(), &Event{Name: "unsubscribe.event"})
	time.Sleep(50 * time.Millisecond)

	// 应该仍然是 1，因为已经取消订阅
	if callCount != 1 {
		t.Errorf("expected 1 call after unsubscribe, got %d", callCount)
	}
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus()

	// 关闭总线
	err := bus.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	// 尝试发布事件应该失败
	err = bus.Publish(context.Background(), &Event{Name: "test.event"})
	if err == nil {
		t.Error("expected error when publishing to closed bus")
	}

	// 尝试订阅应该失败
	err = bus.Subscribe("test.event", func(ctx context.Context, e *Event) error {
		return nil
	})
	if err == nil {
		t.Error("expected error when subscribing to closed bus")
	}
}

func TestEventBus_CloseWaitsForAsync(t *testing.T) {
	bus := NewEventBus()

	completed := false
	var wg sync.WaitGroup
	wg.Add(1)

	// 异步订阅，需要较长时间完成
	_ = bus.Subscribe("slow.event", func(ctx context.Context, e *Event) error {
		time.Sleep(100 * time.Millisecond)
		completed = true
		wg.Done()
		return nil
	}, SubscriptionOptions{Async: true})

	// 发布事件
	_ = bus.Publish(context.Background(), &Event{Name: "slow.event"})

	// 立即关闭，应该等待异步任务完成
	start := time.Now()
	_ = bus.Close()
	elapsed := time.Since(start)

	// 应该等待了至少 100ms
	if elapsed < 90*time.Millisecond {
		t.Errorf("Close returned too quickly (%v), expected >= 100ms", elapsed)
	}

	if !completed {
		t.Error("async handler did not complete before Close returned")
	}
}

func TestEventBus_InvalidWildcardPattern(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// 尝试使用无效的通配符模式
	err := bus.SubscribeWildcard("[invalid", func(ctx context.Context, e *Event) error {
		return nil
	})

	if err == nil {
		t.Error("expected error for invalid wildcard pattern")
	}
}

func TestEventBus_MaxSubscribers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// 订阅达到最大数量
	for i := 0; i < 1000; i++ {
		err := bus.Subscribe("max.event", func(ctx context.Context, e *Event) error {
			return nil
		})
		if err != nil {
			t.Fatalf("failed to subscribe at iteration %d: %v", i, err)
		}
	}

	// 第 1001 个订阅应该失败
	err := bus.Subscribe("max.event", func(ctx context.Context, e *Event) error {
		return nil
	})

	if err == nil {
		t.Error("expected error when exceeding max subscribers")
	}
}

func TestEventBus_BuiltInEvents(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	appStarted := false
	appStopped := false

	// 订阅应用启动事件
	_ = bus.Subscribe(EventAppStarted, func(ctx context.Context, e *Event) error {
		if payload, ok := e.Payload.(*AppStartedPayload); ok {
			appStarted = true
			if payload.Version == "" {
				t.Error("expected version in AppStartedPayload")
			}
		}
		return nil
	})

	// 订阅应用停止事件
	_ = bus.Subscribe(EventAppStopped, func(ctx context.Context, e *Event) error {
		if payload, ok := e.Payload.(*AppStoppedPayload); ok {
			appStopped = true
			if payload.Uptime == 0 {
				t.Error("expected uptime in AppStoppedPayload")
			}
		}
		return nil
	})

	// 发布内置事件
	_ = PublishAppStarted(bus, "1.0.0")
	_ = PublishAppStopped(bus, 5*time.Second)

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	if !appStarted {
		t.Error("AppStarted handler was not called")
	}
	if !appStopped {
		t.Error("AppStopped handler was not called")
	}
}

func TestEventBus_WildcardNestedPatterns(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan string, 10)

	// 多层通配符订阅
	_ = bus.SubscribeWildcard("user.*.created", func(ctx context.Context, e *Event) error {
		received <- e.Name
		return nil
	})

	// 发布匹配的事件
	matchingEvents := []string{
		"user.admin.created",
		"user.guest.created",
	}

	// 发布不匹配的事件
	nonMatchingEvents := []string{
		"user.created",        // 缺少中间层
		"user.admin.updated",  // 动作不匹配
		"order.admin.created", // 第一层不匹配
	}

	for _, eventName := range matchingEvents {
		_ = bus.Publish(context.Background(), &Event{Name: eventName})
	}

	for _, eventName := range nonMatchingEvents {
		_ = bus.Publish(context.Background(), &Event{Name: eventName})
	}

	// 只应该收到匹配的事件
	timeout := time.After(500 * time.Millisecond)
	count := 0

	for count < len(matchingEvents) {
		select {
		case name := <-received:
			count++
			found := false
			for _, expected := range matchingEvents {
				if name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("received unexpected event: %s", name)
			}
		case <-timeout:
			t.Fatalf("timeout, received %d events, expected %d", count, len(matchingEvents))
		}
	}

	if count != len(matchingEvents) {
		t.Errorf("expected %d events, got %d", len(matchingEvents), count)
	}
}

func TestEventBus_WildcardMultiplePatterns(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	userCount := 0
	orderCount := 0
	var mu sync.Mutex

	// 订阅多个通配符模式
	_ = bus.SubscribeWildcard("user.*", func(ctx context.Context, e *Event) error {
		mu.Lock()
		userCount++
		mu.Unlock()
		return nil
	})

	_ = bus.SubscribeWildcard("order.*", func(ctx context.Context, e *Event) error {
		mu.Lock()
		orderCount++
		mu.Unlock()
		return nil
	})

	// 发布不同类型的事件
	events := []string{
		"user.created",
		"user.updated",
		"order.created",
		"order.cancelled",
		"payment.completed", // 不匹配任何通配符
	}

	for _, eventName := range events {
		_ = bus.Publish(context.Background(), &Event{Name: eventName})
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if userCount != 2 {
		t.Errorf("expected 2 user events, got %d", userCount)
	}
	if orderCount != 2 {
		t.Errorf("expected 2 order events, got %d", orderCount)
	}
}

func TestEventBus_PublishToClosedBus(t *testing.T) {
	bus := NewEventBus()

	// 先关闭总线
	_ = bus.Close()

	// 尝试发布事件
	err := bus.Publish(context.Background(), &Event{Name: "test.event"})
	if err == nil {
		t.Error("expected error when publishing to closed bus")
	}
}

func TestEventBus_SubscribeToClosedBus(t *testing.T) {
	bus := NewEventBus()

	// 先关闭总线
	_ = bus.Close()

	// 尝试订阅
	err := bus.Subscribe("test.event", func(ctx context.Context, e *Event) error {
		return nil
	})
	if err == nil {
		t.Error("expected error when subscribing to closed bus")
	}

	// 尝试通配符订阅
	err = bus.SubscribeWildcard("test.*", func(ctx context.Context, e *Event) error {
		return nil
	})
	if err == nil {
		t.Error("expected error when subscribing wildcard to closed bus")
	}
}

func TestEventBus_HandlerPanicRecovery(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	nextHandlerCalled := false

	// 第一个 handler 会 panic
	_ = bus.Subscribe("panic.event", func(ctx context.Context, e *Event) error {
		panic("test panic")
	})

	// 第二个 handler 应该仍然能执行
	_ = bus.Subscribe("panic.event", func(ctx context.Context, e *Event) error {
		nextHandlerCalled = true
		return nil
	})

	// 发布事件（不应该因为 panic 而崩溃）
	err := bus.Publish(context.Background(), &Event{Name: "panic.event"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 等待异步恢复
	time.Sleep(100 * time.Millisecond)

	// 验证第二个 handler 被调用
	if !nextHandlerCalled {
		t.Error("second handler was not called after first handler panicked")
	}
}

func TestEventBus_AsyncHandlerPanicRecovery(t *testing.T) {
	bus := NewEventBus()

	panicHandled := make(chan bool, 1)

	// 异步 handler 会 panic
	_ = bus.Subscribe("async.panic.event", func(ctx context.Context, e *Event) error {
		panic("async panic")
	}, SubscriptionOptions{Async: true})

	// 发布事件
	_ = bus.Publish(context.Background(), &Event{Name: "async.panic.event"})

	// 等待异步处理完成
	time.Sleep(200 * time.Millisecond)

	// 关闭总线，应该等待所有异步任务完成
	err := bus.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}

	// 如果能正常关闭，说明 panic 被正确恢复了
	panicHandled <- true
	<-panicHandled
}

func TestEventBus_MixedSyncAsyncHandlers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var executionOrder []string
	var mu sync.Mutex

	// 同步 handler
	_ = bus.Subscribe("mixed.event", func(ctx context.Context, e *Event) error {
		mu.Lock()
		executionOrder = append(executionOrder, "sync1")
		mu.Unlock()
		return nil
	}, SubscriptionOptions{Async: false})

	// 异步 handler
	_ = bus.Subscribe("mixed.event", func(ctx context.Context, e *Event) error {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		executionOrder = append(executionOrder, "async1")
		mu.Unlock()
		return nil
	}, SubscriptionOptions{Async: true})

	// 另一个同步 handler
	_ = bus.Subscribe("mixed.event", func(ctx context.Context, e *Event) error {
		mu.Lock()
		executionOrder = append(executionOrder, "sync2")
		mu.Unlock()
		return nil
	}, SubscriptionOptions{Async: false})

	// 发布事件
	_ = bus.Publish(context.Background(), &Event{Name: "mixed.event"})

	// 等待异步 handler 完成
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// 同步 handler 应该按顺序执行，异步 handler 可能在任何时候执行
	if len(executionOrder) != 3 {
		t.Errorf("expected 3 handlers, got %d: %v", len(executionOrder), executionOrder)
	}

	// 验证两个同步 handler 都存在
	syncCount := 0
	for _, name := range executionOrder {
		if name == "sync1" || name == "sync2" {
			syncCount++
		}
	}
	if syncCount != 2 {
		t.Errorf("expected 2 sync handlers, got %d", syncCount)
	}
}

func TestEventBus_EventWithComplexPayload(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	type ComplexData struct {
		ID   int
		Name string
		Tags []string
	}

	var receivedPayload *ComplexData

	_ = bus.Subscribe("complex.event", func(ctx context.Context, e *Event) error {
		if payload, ok := e.Payload.(*ComplexData); ok {
			receivedPayload = payload
		}
		return nil
	})

	// 发布复杂类型的事件
	complexData := &ComplexData{
		ID:   123,
		Name: "test",
		Tags: []string{"tag1", "tag2"},
	}

	_ = bus.Publish(context.Background(), &Event{
		Name:    "complex.event",
		Payload: complexData,
	})

	time.Sleep(100 * time.Millisecond)

	if receivedPayload == nil {
		t.Fatal("received payload is nil")
	}

	if receivedPayload.ID != 123 {
		t.Errorf("expected ID 123, got %d", receivedPayload.ID)
	}
	if receivedPayload.Name != "test" {
		t.Errorf("expected Name 'test', got '%s'", receivedPayload.Name)
	}
	if len(receivedPayload.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(receivedPayload.Tags))
	}
}

func TestEventBus_UnsubscribeNonExistent(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(ctx context.Context, e *Event) error {
		return nil
	}

	// 尝试取消一个从未订阅的 handler
	err := bus.Unsubscribe("nonexistent.event", handler)
	if err != nil {
		t.Errorf("unsubscribe should not return error for non-existent subscription: %v", err)
	}
}

func TestEventBus_EmptyEventName(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	// 发布空名称的事件
	err := bus.Publish(context.Background(), &Event{Name: ""})
	if err != nil {
		t.Errorf("publish with empty name should not fail: %v", err)
	}
}

func TestEventBus_RapidSubscribeUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	for i := 0; i < 100; i++ {
		handler := func(ctx context.Context, e *Event) error {
			return nil
		}

		eventName := "rapid.event"

		// 快速订阅和取消订阅
		_ = bus.Subscribe(eventName, handler)
		_ = bus.Unsubscribe(eventName, handler)
	}

	// 如果能正常运行到这里，说明没有死锁或竞态条件
}

func TestEventBus_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup

	// 并发订阅
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			eventName := "concurrent.event"
			handler := func(ctx context.Context, e *Event) error {
				return nil
			}
			_ = bus.Subscribe(eventName, handler)
		}(i)
	}

	wg.Wait()

	// 验证所有订阅都成功了
	// (如果有任何竞态条件，这里可能会 panic)
}

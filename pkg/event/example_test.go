package event_test

import (
	"context"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/pkg/event"
)

// ExampleEventBus 事件总线基本使用示例
// 展示如何订阅和发布事件
func ExampleEventBus() {
	bus := event.NewEventBus()
	defer bus.Close()

	// 订阅事件
	err := bus.Subscribe("user.created", func(ctx context.Context, e *event.Event) error {
		if payload, ok := e.Payload.(string); ok {
			fmt.Printf("User created: %s\n", payload)
		}
		return nil
	})
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// 发布事件
	err = bus.Publish(context.Background(), &event.Event{
		Name:    "user.created",
		Payload: "john",
	})
	if err != nil {
		fmt.Println("publish error:", err)
		return
	}

	// Output: User created: john
}

// ExampleEventBus_async 异步事件处理示例
// 适用于不需要立即完成的后台任务（如发送通知、记录日志等）
func ExampleEventBus_async() {
	bus := event.NewEventBus()
	defer bus.Close()

	received := make(chan string, 1)

	// 异步订阅：handler 会在 goroutine 中执行，不会阻塞 Publish
	err := bus.Subscribe("notification.sent", func(ctx context.Context, e *event.Event) error {
		if msg, ok := e.Payload.(string); ok {
			received <- msg
		}
		return nil
	}, event.SubscriptionOptions{
		Async: true,
	})
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// 发布事件（会立即返回，不等待 handler 完成）
	err = bus.Publish(context.Background(), &event.Event{
		Name:    "notification.sent",
		Payload: "Welcome!",
	})
	if err != nil {
		fmt.Println("publish error:", err)
		return
	}

	// 等待异步处理完成
	msg := <-received
	fmt.Println(msg)
	// Output: Welcome!
}

// ExampleEventBus_wildcard 通配符订阅示例
// 使用通配符匹配多个相关事件，适合事件分组处理
func ExampleEventBus_wildcard() {
	bus := event.NewEventBus()
	defer bus.Close()

	eventCount := 0

	// 通配符订阅：匹配所有 user.* 事件
	err := bus.SubscribeWildcard("user.*", func(ctx context.Context, e *event.Event) error {
		eventCount++
		return nil
	})
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// 发布多个匹配的事件
	events := []string{"user.created", "user.updated", "user.deleted"}
	for _, eventName := range events {
		_ = bus.Publish(context.Background(), &event.Event{Name: eventName})
	}

	// 这个事件不会被上面的订阅者接收
	_ = bus.Publish(context.Background(), &event.Event{Name: "order.created"})

	fmt.Println(eventCount)
	// Output: 3
}

// ExampleEventBus_priority 事件优先级示例
// 高优先级订阅者会先于低优先级订阅者执行
func ExampleEventBus_priority() {
	bus := event.NewEventBus()
	defer bus.Close()

	var executionOrder []string

	// 低优先级订阅者
	_ = bus.Subscribe("important.event", func(ctx context.Context, e *event.Event) error {
		executionOrder = append(executionOrder, "low")
		return nil
	}, event.SubscriptionOptions{
		Priority: event.PriorityLow,
	})

	// 高优先级订阅者
	_ = bus.Subscribe("important.event", func(ctx context.Context, e *event.Event) error {
		executionOrder = append(executionOrder, "high")
		return nil
	}, event.SubscriptionOptions{
		Priority: event.PriorityHigh,
	})

	// 普通优先级订阅者
	_ = bus.Subscribe("important.event", func(ctx context.Context, e *event.Event) error {
		executionOrder = append(executionOrder, "normal")
		return nil
	}, event.SubscriptionOptions{
		Priority: event.PriorityNormal,
	})

	// 发布事件
	_ = bus.Publish(context.Background(), &event.Event{Name: "important.event"})

	// 等待所有 handler 执行完成
	time.Sleep(100 * time.Millisecond)

	// 验证执行顺序：高 -> 普通 -> 低
	fmt.Println(executionOrder)
	// Output: [high normal low]
}

// ExampleEventBus_withContext 带 Context 的事件处理示例
// 支持超时控制和取消传播
func ExampleEventBus_withContext() {
	bus := event.NewEventBus()
	defer bus.Close()

	// 订阅一个会检查 context 的 handler
	err := bus.Subscribe("timeout.event", func(ctx context.Context, e *event.Event) error {
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled:", ctx.Err())
			return ctx.Err()
		case <-time.After(2 * time.Second):
			fmt.Println("Handler completed")
			return nil
		}
	})
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 发布事件（handler 会因为超时被取消）
	_ = bus.Publish(ctx, &event.Event{Name: "timeout.event"})

	// 等待 handler 执行
	time.Sleep(200 * time.Millisecond)

	// Output: Context canceled: context deadline exceeded
}

// ExampleEventBus_multipleHandlers 多个订阅者示例
// 一个事件可以有多个订阅者，它们都会收到事件
func ExampleEventBus_multipleHandlers() {
	bus := event.NewEventBus()
	defer bus.Close()

	// 第一个订阅者：记录日志
	_ = bus.Subscribe("order.placed", func(ctx context.Context, e *event.Event) error {
		fmt.Println("[Logger] Order placed event received")
		return nil
	})

	// 第二个订阅者：发送通知
	_ = bus.Subscribe("order.placed", func(ctx context.Context, e *event.Event) error {
		fmt.Println("[Notifier] Sending notification for order")
		return nil
	})

	// 第三个订阅者：更新统计
	_ = bus.Subscribe("order.placed", func(ctx context.Context, e *event.Event) error {
		fmt.Println("[Analytics] Updating order statistics")
		return nil
	})

	// 发布事件，所有订阅者都会收到
	_ = bus.Publish(context.Background(), &event.Event{
		Name: "order.placed",
	})

	// Output:
	// [Logger] Order placed event received
	// [Notifier] Sending notification for order
	// [Analytics] Updating order statistics
}

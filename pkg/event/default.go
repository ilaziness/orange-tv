package event

import (
	"context"
	"fmt"
	"sync"
)

// 受控单例：应用启动时由 app 层调用 SetDefault 设置一次，运行期只读。
// 测试可通过 SetDefault(bus) + defer SetDefault(nil) 替换/重置。
//
// 注意：此为用户明确要求的受控单例设计，非任意可变全局状态。
var (
	defaultMu  sync.RWMutex
	defaultBus EventBus
)

// SetDefault 设置全局默认事件总线。应在应用启动时调用一次。
// 重复调用会覆盖之前的默认总线（用于测试场景）。传 nil 可重置。
func SetDefault(bus EventBus) {
	defaultMu.Lock()
	defaultBus = bus
	defaultMu.Unlock()
}

// Default 返回当前默认事件总线，未设置时返回 nil。
func Default() EventBus {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultBus
}

// Publish 通过默认总线发布事件。默认总线未设置时返回 nil（安全降级，不 panic）。
func Publish(ctx context.Context, e *Event) error {
	bus := Default()
	if bus == nil {
		return nil
	}
	return bus.Publish(ctx, e)
}

// Subscribe 通过默认总线订阅事件。默认总线未设置时返回 error。
func Subscribe(eventName string, handler EventHandler, opts ...SubscriptionOptions) error {
	bus := Default()
	if bus == nil {
		return fmt.Errorf("default event bus is not set")
	}
	return bus.Subscribe(eventName, handler, opts...)
}

// Unsubscribe 通过默认总线取消订阅。默认总线未设置时返回 nil。
func Unsubscribe(eventName string, handler EventHandler) error {
	bus := Default()
	if bus == nil {
		return nil
	}
	return bus.Unsubscribe(eventName, handler)
}

// SubscribeWildcard 通过默认总线通配符订阅。默认总线未设置时返回 error。
func SubscribeWildcard(pattern string, handler EventHandler, opts ...SubscriptionOptions) error {
	bus := Default()
	if bus == nil {
		return fmt.Errorf("default event bus is not set")
	}
	return bus.SubscribeWildcard(pattern, handler, opts...)
}

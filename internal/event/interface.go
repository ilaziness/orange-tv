package event

import (
	"context"
)

// EventBus 事件总线接口
type EventBus interface {
	// Publish 发布事件
	Publish(ctx context.Context, event *Event) error

	// Subscribe 订阅事件
	Subscribe(eventName string, handler EventHandler, opts ...SubscriptionOptions) error

	// Unsubscribe 取消订阅
	Unsubscribe(eventName string, handler EventHandler) error

	// SubscribeWildcard 通配符订阅（支持 "user.*" 匹配 "user.created"）
	SubscribeWildcard(pattern string, handler EventHandler, opts ...SubscriptionOptions) error

	// Close 关闭事件总线
	Close() error
}

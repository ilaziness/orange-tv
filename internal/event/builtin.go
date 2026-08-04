package event

import (
	"context"
	"time"

	pkgevent "github.com/ilaziness/orange-tv/pkg/event"
)

// 重导出核心类型，下游包通过 internal/event 即可使用，无需直接 import pkg/event
type EventBus = pkgevent.EventBus
type Event = pkgevent.Event
type EventHandler = pkgevent.EventHandler
type SubscriptionOptions = pkgevent.SubscriptionOptions

// Subscribe 通过默认总线订阅事件。下游包只需 import internal/event 即可订阅，
// 无需直接依赖 pkg/event 或传递 EventBus 实例。
func Subscribe(eventName string, handler EventHandler, opts ...SubscriptionOptions) error {
	return pkgevent.Subscribe(eventName, handler, opts...)
}

// 内置事件名称常量
const (
	// EventAppStarted 应用启动事件
	EventAppStarted = "app.started"
	// EventAppStopped 应用停止事件
	EventAppStopped = "app.stopped"
	// EventServiceReady 服务就绪事件
	EventServiceReady = "service.ready"
	// EventConnOpened 连接建立事件
	EventConnOpened = "connection.opened"
	// EventConnClosed 连接断开事件
	EventConnClosed = "connection.closed"
)

// AppStartedPayload 应用启动事件数据
type AppStartedPayload struct {
	StartTime time.Time
	Version   string
}

// AppStoppedPayload 应用停止事件数据
type AppStoppedPayload struct {
	StopTime time.Time
	Uptime   time.Duration
}

// PublishAppStarted 发布应用启动事件（通过默认事件总线）
func PublishAppStarted(version string) error {
	return pkgevent.Publish(context.Background(), &pkgevent.Event{
		Name:      EventAppStarted,
		Payload:   &AppStartedPayload{StartTime: time.Now(), Version: version},
		Timestamp: time.Now(),
	})
}

// PublishAppStopped 发布应用停止事件（通过默认事件总线）
func PublishAppStopped(uptime time.Duration) error {
	return pkgevent.Publish(context.Background(), &pkgevent.Event{
		Name:      EventAppStopped,
		Payload:   &AppStoppedPayload{StopTime: time.Now(), Uptime: uptime},
		Timestamp: time.Now(),
	})
}

package event

import (
	"context"
	"time"
)

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

// PublishAppStarted 发布应用启动事件
func PublishAppStarted(bus EventBus, version string) error {
	event := &Event{
		Name: EventAppStarted,
		Payload: &AppStartedPayload{
			StartTime: time.Now(),
			Version:   version,
		},
		Timestamp: time.Now(),
		Context:   context.Background(),
	}
	return bus.Publish(context.Background(), event)
}

// PublishAppStopped 发布应用停止事件
func PublishAppStopped(bus EventBus, uptime time.Duration) error {
	event := &Event{
		Name: EventAppStopped,
		Payload: &AppStoppedPayload{
			StopTime: time.Now(),
			Uptime:   uptime,
		},
		Timestamp: time.Now(),
		Context:   context.Background(),
	}
	return bus.Publish(context.Background(), event)
}

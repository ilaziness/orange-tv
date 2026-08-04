package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/event"
	pkgevent "github.com/ilaziness/orange-tv/pkg/event"
)

func TestBuiltInEvents(t *testing.T) {
	bus := pkgevent.NewEventBus()
	defer bus.Close()
	pkgevent.SetDefault(bus)
	defer pkgevent.SetDefault(nil)

	appStarted := false
	appStopped := false

	// 订阅应用启动事件
	_ = pkgevent.Subscribe(event.EventAppStarted, func(ctx context.Context, e *event.Event) error {
		if payload, ok := e.Payload.(*event.AppStartedPayload); ok {
			appStarted = true
			if payload.Version == "" {
				t.Error("expected version in AppStartedPayload")
			}
		}
		return nil
	})

	// 订阅应用停止事件
	_ = pkgevent.Subscribe(event.EventAppStopped, func(ctx context.Context, e *event.Event) error {
		if payload, ok := e.Payload.(*event.AppStoppedPayload); ok {
			appStopped = true
			if payload.Uptime == 0 {
				t.Error("expected uptime in AppStoppedPayload")
			}
		}
		return nil
	})

	// 发布内置事件（通过默认总线）
	_ = event.PublishAppStarted("1.0.0")
	_ = event.PublishAppStopped(5 * time.Second)

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	if !appStarted {
		t.Error("AppStarted handler was not called")
	}
	if !appStopped {
		t.Error("AppStopped handler was not called")
	}
}

package event

import (
	"context"
	"time"

	pkgevent "github.com/ilaziness/orange-tv/pkg/event"
)

// 采集相关事件名称常量
const (
	// EventCollectScheduleChanged 采集调度变更事件
	// 由 service 在 EnableSchedule/DisableSchedule/DeleteSource 成功后发布，
	// scheduler 订阅后执行 Reload 重新加载 cron 任务。
	EventCollectScheduleChanged = "collect.schedule.changed"
)

// CollectScheduleChangedPayload 采集调度变更事件数据
type CollectScheduleChangedPayload struct {
	SourceID uint32 // 触发变更的采集源 ID，仅用于日志追踪
}

// PublishCollectScheduleChanged 通过默认总线发布采集调度变更事件。
// 默认总线未设置时返回 nil（安全降级，便于测试）。
func PublishCollectScheduleChanged(sourceID uint32) error {
	return pkgevent.Publish(context.Background(), &pkgevent.Event{
		Name:      EventCollectScheduleChanged,
		Payload:   &CollectScheduleChangedPayload{SourceID: sourceID},
		Timestamp: time.Now(),
	})
}

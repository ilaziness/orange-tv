package event

import (
	"context"
	"time"
)

// Event 事件基础结构
type Event struct {
	Name      string          // 事件名称
	Payload   any             // 事件数据
	Timestamp time.Time       // 事件时间戳
	Context   context.Context // 上下文（用于链路追踪等）
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// HandlerPriority 处理器优先级
type HandlerPriority int

const (
	// PriorityHigh 高优先级
	PriorityHigh HandlerPriority = 1
	// PriorityNormal 普通优先级
	PriorityNormal HandlerPriority = 2
	// PriorityLow 低优先级
	PriorityLow HandlerPriority = 3
)

// SubscriptionOptions 订阅选项
type SubscriptionOptions struct {
	Async    bool            // 是否异步处理
	Priority HandlerPriority // 优先级
}

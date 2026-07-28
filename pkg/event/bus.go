package event

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"go.uber.org/zap"
)

const (
	// maxSubscribersPerEvent 每个事件的最大订阅者数量
	maxSubscribersPerEvent = 1000
)

type subscriber struct {
	handler  EventHandler
	async    bool
	priority HandlerPriority
	id       uint64
}

// EventBusImpl 事件总线实现
type EventBusImpl struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscriber
	patterns    map[string][]*subscriber // 通配符订阅
	nextID      uint64
	wg          sync.WaitGroup
	closed      bool
	logger      *zap.Logger
}

// NewEventBus 创建新的事件总线
func NewEventBus() EventBus {
	return NewEventBusWithLogger(zap.NewNop())
}

// NewEventBusWithLogger 创建新的事件总线，使用指定的 logger
func NewEventBusWithLogger(logger *zap.Logger) EventBus {
	return &EventBusImpl{
		subscribers: make(map[string][]*subscriber),
		patterns:    make(map[string][]*subscriber),
		nextID:      1,
		logger:      logger,
	}
}

// Publish 发布事件
func (bus *EventBusImpl) Publish(ctx context.Context, event *Event) error {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if bus.closed {
		return fmt.Errorf("event bus is closed")
	}

	// 1. 精确匹配订阅者
	if subs, ok := bus.subscribers[event.Name]; ok {
		for _, sub := range subs {
			bus.dispatchEvent(ctx, event, sub)
		}
	}

	// 2. 通配符匹配订阅者
	for pattern, subs := range bus.patterns {
		if bus.matchPattern(pattern, event.Name) {
			for _, sub := range subs {
				bus.dispatchEvent(ctx, event, sub)
			}
		}
	}

	return nil
}

// dispatchEvent 分发事件到订阅者
func (bus *EventBusImpl) dispatchEvent(ctx context.Context, event *Event, sub *subscriber) {
	if sub.async {
		bus.wg.Add(1)
		go func() {
			defer bus.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// 记录 panic 信息，防止 goroutine 崩溃
					bus.logger.Warn("event handler panic (async)", zap.Any("recover", r), zap.String("event", event.Name))
				}
			}()
			if err := sub.handler(ctx, event); err != nil {
				// 记录错误但不中断其他 handler
				bus.logger.Warn("event handler error (async)", zap.Error(err), zap.String("event", event.Name))
			}
		}()
	} else {
		defer func() {
			if r := recover(); r != nil {
				bus.logger.Warn("event handler panic (sync)", zap.Any("recover", r), zap.String("event", event.Name))
			}
		}()
		if err := sub.handler(ctx, event); err != nil {
			bus.logger.Warn("event handler error (sync)", zap.Error(err), zap.String("event", event.Name))
		}
	}
}

// Subscribe 订阅事件
func (bus *EventBusImpl) Subscribe(eventName string, handler EventHandler, opts ...SubscriptionOptions) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		return fmt.Errorf("event bus is closed")
	}

	opt := SubscriptionOptions{
		Async:    false,
		Priority: PriorityNormal,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 检查订阅者数量限制
	if len(bus.subscribers[eventName]) >= maxSubscribersPerEvent {
		return fmt.Errorf("maximum subscribers (%d) reached for event %s", maxSubscribersPerEvent, eventName)
	}

	sub := &subscriber{
		handler:  handler,
		async:    opt.Async,
		priority: opt.Priority,
		id:       bus.nextID,
	}
	bus.nextID++

	bus.subscribers[eventName] = append(bus.subscribers[eventName], sub)

	// 按优先级排序
	bus.sortSubscribers(eventName)

	return nil
}

// Unsubscribe 取消订阅
func (bus *EventBusImpl) Unsubscribe(eventName string, handler EventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	subs, ok := bus.subscribers[eventName]
	if !ok {
		return nil
	}

	// 找到并移除订阅者
	newSubs := make([]*subscriber, 0, len(subs))
	for _, sub := range subs {
		// 通过函数指针比较来识别订阅者
		if fmt.Sprintf("%p", sub.handler) != fmt.Sprintf("%p", handler) {
			newSubs = append(newSubs, sub)
		}
	}

	bus.subscribers[eventName] = newSubs
	return nil
}

// SubscribeWildcard 通配符订阅
func (bus *EventBusImpl) SubscribeWildcard(pattern string, handler EventHandler, opts ...SubscriptionOptions) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		return fmt.Errorf("event bus is closed")
	}

	// 验证模式有效性
	if _, err := filepath.Match(pattern, "test"); err != nil {
		return fmt.Errorf("invalid wildcard pattern: %w", err)
	}

	opt := SubscriptionOptions{
		Async:    false,
		Priority: PriorityNormal,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}

	sub := &subscriber{
		handler:  handler,
		async:    opt.Async,
		priority: opt.Priority,
		id:       bus.nextID,
	}
	bus.nextID++

	bus.patterns[pattern] = append(bus.patterns[pattern], sub)
	// 按优先级排序
	bus.sortWildcardSubscribers(pattern)
	return nil
}

// sortWildcardSubscribers 按优先级排序通配符订阅者
func (bus *EventBusImpl) sortWildcardSubscribers(pattern string) {
	subs := bus.patterns[pattern]
	sort.Slice(subs, func(i, j int) bool {
		return subs[i].priority < subs[j].priority
	})
}

// matchPattern 检查事件名是否匹配模式
func (bus *EventBusImpl) matchPattern(pattern, eventName string) bool {
	// 使用 filepath.Match 支持通配符
	matched, err := filepath.Match(pattern, eventName)
	if err != nil {
		// 记录错误但不中断处理
		bus.logger.Warn("invalid pattern", zap.String("pattern", pattern), zap.Error(err))
		return false
	}
	return matched
}

// sortSubscribers 按优先级排序订阅者
func (bus *EventBusImpl) sortSubscribers(eventName string) {
	subs := bus.subscribers[eventName]
	sort.Slice(subs, func(i, j int) bool {
		return subs[i].priority < subs[j].priority
	})
}

// Close 关闭事件总线
func (bus *EventBusImpl) Close() error {
	bus.mu.Lock()
	bus.closed = true
	bus.mu.Unlock()

	bus.wg.Wait() // 等待所有异步事件处理完成
	return nil
}

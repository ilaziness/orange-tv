package scheduler

import (
	"context"
	"strings"
	"sync"

	"github.com/ilaziness/orange-tv/internal/event"
	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// JobRunner 由 collect service 实现，调度器只关心"触发后执行采集"。
type JobRunner interface {
	RunScheduledJob(source *model.CollectSources, dataRange string) error
}

// CollectScheduler 采集调度注册器：订阅调度变更事件 → 从 DB 加载启用的采集源 → 注册/重载 cron 任务。
// 不持有 EventBus，订阅通过 event 包默认总线便捷函数完成。
// cron 实例的生命周期（Start/Stop）由 Manager 统一管理，本结构只负责注册任务。
type CollectScheduler struct {
	mgr    *Manager
	repo   repository.CollectRepository
	runner JobRunner

	mu      sync.Mutex
	cronIDs map[uint32]cron.EntryID
}

// NewCollectScheduler 创建采集调度器。
func NewCollectScheduler(mgr *Manager, repo repository.CollectRepository, runner JobRunner) *CollectScheduler {
	return &CollectScheduler{
		mgr:     mgr,
		repo:    repo,
		runner:  runner,
		cronIDs: map[uint32]cron.EntryID{},
	}
}

// Init 订阅调度变更事件并执行首次 Reload。
// cron 实例的启动由 Manager.Start 统一完成，应在 Init 之前调用。
func (s *CollectScheduler) Init(ctx context.Context) error {
	if err := event.Subscribe(event.EventCollectScheduleChanged, s.handleScheduleChanged); err != nil {
		logger.Log.Warn("subscribe collect schedule changed event failed", zap.Error(err))
	}
	return s.Reload(ctx)
}

// handleScheduleChanged 是调度变更事件处理器，触发全量 Reload。
func (s *CollectScheduler) handleScheduleChanged(ctx context.Context, ev *event.Event) error {
	if payload, ok := ev.Payload.(*event.CollectScheduleChangedPayload); ok {
		logger.Log.Info("collect schedule changed, reloading cron jobs",
			zap.Uint32("source_id", payload.SourceID))
	}
	return s.Reload(ctx)
}

// Reload 从 DB 重新加载启用的采集源并注册 cron 任务。
// 先查询 DB，成功后再替换旧 entry，确保 DB 故障时现有 cron 任务不受影响。
// 整个操作持锁，防止并发 Reload 导致重复注册 cron entry。
func (s *CollectScheduler) Reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 先查询 DB，失败时保留旧 entry，避免 DB 故障导致所有定时任务丢失
	sources, err := s.repo.ListEnabledCronSources(ctx)
	if err != nil {
		logger.Log.Error("collect: list enabled cron sources failed", zap.Error(err))
		return err
	}

	// 查询成功，移除旧 entry
	for id, entryID := range s.cronIDs {
		s.mgr.Remove(entryID)
		delete(s.cronIDs, id)
	}

	parser := DefaultParser()
	for i := range sources {
		src := sources[i]
		expr := strings.TrimSpace(src.CronExpr)
		if expr == "" {
			continue
		}
		if _, err := parser.Parse(expr); err != nil {
			logger.Log.Warn("skip invalid cron", zap.Uint32("source_id", src.ID), zap.Error(err))
			continue
		}
		sourceID := src.ID
		dataRange := src.DataRange
		entryID, err := s.mgr.AddFunc(expr, func() {
			_ = s.runner.RunScheduledJob(&src, dataRange)
		})
		if err != nil {
			logger.Log.Warn("add cron failed", zap.Uint32("source_id", sourceID), zap.Error(err))
			continue
		}
		s.cronIDs[sourceID] = entryID
	}
	return nil
}

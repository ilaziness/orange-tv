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

// CollectRunFunc 是采集任务的执行回调，由 collect service 实现。
// 调度器只关心"触发后执行采集"，不依赖具体 service 类型。
type CollectRunFunc func(source *model.CollectSources, dataRange string) error

// CollectJob 采集调度任务：订阅调度变更事件 → 从 DB 加载启用的采集源 → 注册/重载 cron 任务。
// 实现 Job 接口，由 Scheduler 在 Start 时调用 Init 完成首次注册。
type CollectJob struct {
	repo  repository.CollectRepository
	runFn CollectRunFunc
	sched *Scheduler // 在 Init 中由 Scheduler 注入，供后续 Reload 使用

	mu      sync.Mutex
	cronIDs map[uint32]cron.EntryID
}

// NewCollectJob 创建采集调度任务。
// 仅需业务依赖（repo + runFn），Scheduler 引用在 Init 时注入。
func NewCollectJob(repo repository.CollectRepository, runFn CollectRunFunc) *CollectJob {
	return &CollectJob{
		repo:    repo,
		runFn:   runFn,
		cronIDs: map[uint32]cron.EntryID{},
	}
}

// Init 订阅调度变更事件并执行首次 Reload。
// sched 由 Scheduler 在 Start 时传入，job 保存引用供后续事件触发的 Reload 使用。
func (j *CollectJob) Init(ctx context.Context, sched *Scheduler) error {
	j.sched = sched
	if err := event.Subscribe(event.EventCollectScheduleChanged, j.handleScheduleChanged); err != nil {
		logger.Log.Warn("subscribe collect schedule changed event failed", zap.Error(err))
	}
	return j.Reload(ctx)
}

// handleScheduleChanged 是调度变更事件处理器，触发全量 Reload。
func (j *CollectJob) handleScheduleChanged(ctx context.Context, ev *event.Event) error {
	if payload, ok := ev.Payload.(*event.CollectScheduleChangedPayload); ok {
		logger.Log.Info("collect schedule changed, reloading cron jobs",
			zap.Uint32("source_id", payload.SourceID))
	}
	return j.Reload(ctx)
}

// Reload 从 DB 重新加载启用的采集源并注册 cron 任务。
// 先查询 DB，成功后再替换旧 entry，确保 DB 故障时现有 cron 任务不受影响。
// 整个操作持锁，防止并发 Reload 导致重复注册 cron entry。
func (j *CollectJob) Reload(ctx context.Context) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// 先查询 DB，失败时保留旧 entry，避免 DB 故障导致所有定时任务丢失
	sources, err := j.repo.ListEnabledCronSources(ctx)
	if err != nil {
		logger.Log.Error("collect: list enabled cron sources failed", zap.Error(err))
		return err
	}

	// 查询成功，移除旧 entry
	for id, entryID := range j.cronIDs {
		j.sched.Remove(entryID)
		delete(j.cronIDs, id)
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
		entryID, err := j.sched.AddFunc(expr, func() {
			_ = j.runFn(&src, dataRange)
		})
		if err != nil {
			logger.Log.Warn("add cron failed", zap.Uint32("source_id", sourceID), zap.Error(err))
			continue
		}
		j.cronIDs[sourceID] = entryID
	}
	return nil
}

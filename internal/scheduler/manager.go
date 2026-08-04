// Package scheduler 提供定时任务（cron）调度管理，与业务 service 解耦。
// 调度生命周期由 Manager 统一管理，业务执行通过 JobRunner 接口回调。
package scheduler

import (
	"context"
	"time"

	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/robfig/cron/v3"
)

// defaultParser 项目标准 cron parser（5 字段 + descriptor）。
// service 校验与调度注册共用同一 parser，保证一致性。
var defaultParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// DefaultParser 返回项目标准 cron parser。
func DefaultParser() cron.Parser {
	return defaultParser
}

// Job 定义一个定时调度任务。每个任务实现 Init，在 cron 启动后由 Manager 统一调用。
// Init 内部通常订阅调度变更事件并执行首次 Reload。
type Job interface {
	Init(ctx context.Context) error
}

// Manager 封装 robfig/cron/v3，提供通用 cron 调度生命周期管理。
// 持有所有注册的 Job，统一启动（Start）和停止（Stop）。
// 日志使用全局 logger.Log，不注入。
type Manager struct {
	cron *cron.Cron
	jobs []Job
}

// NewManager 创建通用 Cron 管理器。
func NewManager() *Manager {
	return &Manager{
		cron: cron.New(
			cron.WithParser(defaultParser),
			cron.WithChain(cron.Recover(cron.DefaultLogger)),
		),
	}
}

// Register 注册一个调度任务。应在 Start 之前调用。
func (m *Manager) Register(job Job) {
	m.jobs = append(m.jobs, job)
}

// Start 启动 cron 调度，并依次调用所有已注册 Job 的 Init。
// 某个 Job Init 失败会中止后续 Job 的初始化，停止 cron 并返回错误。
func (m *Manager) Start(ctx context.Context) error {
	m.cron.Start()
	for _, job := range m.jobs {
		if err := job.Init(ctx); err != nil {
			// cron 已启动，失败时必须停止，否则 goroutine 泄漏
			// （lifecycle 不会对 OnStart 失败的 hook 调用 OnStop）。
			m.stopCron(ctx)
			return err
		}
	}
	return nil
}

// Stop 优雅停止 cron 调度，等待运行中的任务完成或超时。
func (m *Manager) Stop(ctx context.Context) error {
	if m.cron == nil {
		return nil
	}
	m.stopCron(ctx)
	return nil
}

// stopCron 停止 cron 并等待运行中的任务完成，最多等待 5 秒。
func (m *Manager) stopCron(ctx context.Context) {
	stopCtx := m.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		logger.Log.Warn("cron manager stop timeout after 5s")
	}
}

// AddFunc 注册一个 cron 任务，返回 entry ID。
func (m *Manager) AddFunc(expr string, fn func()) (cron.EntryID, error) {
	return m.cron.AddFunc(expr, fn)
}

// Remove 移除一个 cron 任务。
func (m *Manager) Remove(id cron.EntryID) {
	m.cron.Remove(id)
}

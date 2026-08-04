// Package scheduler 提供定时任务（cron）调度管理。
//
// Scheduler 封装 robfig/cron/v3，统一管理所有 Job 的生命周期（Start/Stop），
// 并可选通过 LockProvider 实现多实例分布式锁保护。
//
// 新增任务只需实现 Job 接口并调用 Register 注册：
//
//	sched := scheduler.NewScheduler(redisClient)
//	sched.Register(scheduler.NewCollectJob(repo, runner))
//	sched.Start(ctx)
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// defaultParser 项目标准 cron parser（5 字段 + descriptor）。
// service 校验与调度注册共用同一 parser，保证一致性。
var defaultParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// DefaultParser 返回项目标准 cron parser。
func DefaultParser() cron.Parser {
	return defaultParser
}

// Job 定义一个定时调度任务。
// Init 在 cron 启动后由 Scheduler 统一调用，任务在 Init 内完成首次注册
// （通过传入的 Scheduler 调用 AddFunc）并订阅所需事件。
type Job interface {
	// Init 在 Scheduler.Start 期间、cron 已启动后调用。
	// sched 用于注册/移除 cron entry，job 可保存引用供后续 Reload 使用。
	Init(ctx context.Context, sched *Scheduler) error
}

// ErrLockNotHeld 表示锁已丢失（过期或被其他实例抢占）。
// RenewSchedulerLock 返回此错误时，Scheduler 会停止 cron 以避免 split-brain。
var ErrLockNotHeld = errors.New("scheduler: lock not held")

// LockProvider 提供调度器单实例分布式锁能力。
// 多机部署时，通过分布式锁保证只有一个实例运行 cron 调度，
// 避免重复执行采集等定时任务。
// 单实例部署时使用 NoopLockProvider（空锁），不做互斥保护。
// 调用方通过 NewScheduler 传入 redis client，由其自动选择锁策略，无需直接构造。
type LockProvider interface {
	// AcquireSchedulerLock 尝试获取调度器锁，成功返回 true。
	AcquireSchedulerLock(ctx context.Context) (bool, error)
	// RenewSchedulerLock 续期调度器锁，重置 TTL。
	RenewSchedulerLock(ctx context.Context) error
	// ReleaseSchedulerLock 释放调度器锁。
	ReleaseSchedulerLock(ctx context.Context) error
}

// defaultLockRenewInterval 默认锁续期间隔，需小于锁 TTL 以留出续期余量。
const defaultLockRenewInterval = 10 * time.Second

// NoopLockProvider 是空锁实现，用于未启用 Redis 的单实例部署场景。
// Acquire 始终返回 true，Renew/Release 为空操作，不提供任何互斥保护。
// 实现 noHeartbeatLock 接口，Scheduler 会跳过心跳 goroutine（空锁无需续期）。
type NoopLockProvider struct{}

// AcquireSchedulerLock 始终返回成功（true, nil）。
func (NoopLockProvider) AcquireSchedulerLock(ctx context.Context) (bool, error) {
	return true, nil
}

// RenewSchedulerLock 空操作，始终返回 nil。
func (NoopLockProvider) RenewSchedulerLock(ctx context.Context) error {
	return nil
}

// ReleaseSchedulerLock 空操作，始终返回 nil。
func (NoopLockProvider) ReleaseSchedulerLock(ctx context.Context) error {
	return nil
}

// noHeartbeat 标记此锁不需要心跳续期，Scheduler 据此跳过 startHeartbeat。
func (NoopLockProvider) noHeartbeat() {}

// Scheduler 封装 robfig/cron/v3，提供通用 cron 调度生命周期管理。
// 持有所有注册的 Job，统一启动（Start）和停止（Stop）。
// 日志使用全局 logger.Log，不注入。
type Scheduler struct {
	cron *cron.Cron
	jobs []Job
	lock LockProvider

	// cronStarted 标记 cron 是否已实际启动，用于 Stop 判断是否需要停止。
	// 使用 atomic.Bool 因为心跳 goroutine 在锁丢失时也会写入此字段。
	cronStarted atomic.Bool

	// lockOwned 标记当前实例是否持有分布式锁，用于 Stop 判断是否需要释放锁。
	// 仅在主 goroutine（Start/Stop/releaseLock）中读写，无需原子操作。
	// 未配置 LockProvider 时始终为 false。
	lockOwned bool

	// lockRenewInterval 锁续期间隔，需小于锁 TTL 以留出续期余量。
	// 测试时可覆盖为短间隔以加速。
	lockRenewInterval time.Duration

	// heartbeat 相关字段：用 context 控制心跳 goroutine 退出，
	// heartbeatDone 在 goroutine 退出时关闭，供 releaseLock 等待。
	heartbeatCancel context.CancelFunc
	heartbeatDone   chan struct{}
}

// newSchedulerWithLock 是内部构造函数，用指定的锁创建调度器。
// 外部应使用 NewScheduler（redislock.go）传入 redis client，由其决定锁策略。
func newSchedulerWithLock(lock LockProvider) *Scheduler {
	return &Scheduler{
		cron: cron.New(
			cron.WithParser(defaultParser),
			cron.WithChain(cron.Recover(cron.DefaultLogger)),
		),
		lock:              lock,
		lockRenewInterval: defaultLockRenewInterval,
	}
}

// Register 注册一个调度任务。应在 Start 之前调用。
func (s *Scheduler) Register(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start 启动 cron 调度，并依次调用所有已注册 Job 的 Init。
//
// 若配置了锁（LockProvider 非 nil），会先尝试获取锁：
//   - 获取失败（Redis 锁被其他实例持有）：本实例跳过启动，直接返回 nil。
//     HTTP 服务等其他模块不受影响，仍正常运行。
//   - 获取成功：启动 cron 与 Job Init。Redis 锁会开启后台 goroutine 定期续期，
//     NoopLockProvider 跳过心跳（空锁无需续期）。
//
// 某个 Job Init 失败会中止后续 Job 的初始化，停止 cron、释放锁并返回错误。
// s.lock 为 nil 时（仅测试场景，通过 newSchedulerWithLock(nil)）跳过锁逻辑直接启动 cron。
func (s *Scheduler) Start(ctx context.Context) error {
	// 多实例保护：尝试获取锁（NoopLockProvider 始终成功，Redis 锁竞争失败则跳过启动）
	if s.lock != nil {
		ok, err := s.lock.AcquireSchedulerLock(ctx)
		if err != nil {
			return fmt.Errorf("scheduler: acquire lock: %w", err)
		}
		if !ok {
			logger.Log.Info("scheduler: another instance already holds the lock, skipping cron startup")
			return nil
		}
		s.lockOwned = true
		// NoopLockProvider 等实现 noHeartbeatLock 接口的锁不需要续期，跳过心跳 goroutine。
		if _, skip := s.lock.(interface{ noHeartbeat() }); !skip {
			s.startHeartbeat()
		}
	}

	s.cron.Start()
	s.cronStarted.Store(true)
	for _, job := range s.jobs {
		if err := job.Init(ctx, s); err != nil {
			// cron 已启动，失败时必须停止，否则 goroutine 泄漏
			// （lifecycle 不会对 OnStart 失败的 hook 调用 OnStop）。
			s.stopCron(ctx)
			s.cronStarted.Store(false)
			s.releaseLock(ctx)
			return err
		}
	}
	return nil
}

// startHeartbeat 启动后台 goroutine 定期续期调度器锁，防止锁过期被其他实例抢占。
// 续期返回 ErrLockNotHeld（锁已丢失）时停止 cron 并退出，避免 split-brain。
// 临时错误仅记录日志并继续重试，锁 TTL 内恢复则不影响。
// panic 由 utils.Go 统一 recover 并记录日志。
func (s *Scheduler) startHeartbeat() {
	hbCtx, cancel := context.WithCancel(context.Background())
	s.heartbeatCancel = cancel
	done := make(chan struct{})
	s.heartbeatDone = done
	utils.Go(func() {
		defer close(done)
		ticker := time.NewTicker(s.lockRenewInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				err := s.lock.RenewSchedulerLock(ctx)
				cancel()
				if err == nil {
					continue
				}
				if errors.Is(err, ErrLockNotHeld) {
					// 锁已丢失，停止 cron 避免 split-brain。
					// 不调用 releaseLock（会等待自身退出导致死锁），仅停止 cron 并退出。
					logger.Log.Error("scheduler: lock lost, stopping cron to avoid split-brain")
					s.stopCron(context.Background())
					s.cronStarted.Store(false)
					return
				}
				logger.Log.Warn("scheduler: failed to renew lock", zap.Error(err))
			case <-hbCtx.Done():
				return
			}
		}
	})
}

// releaseLock 停止心跳并释放分布式锁。仅在持锁时执行。
func (s *Scheduler) releaseLock(ctx context.Context) {
	if !s.lockOwned {
		return
	}
	// 先停止心跳，避免释放后还在续期
	if s.heartbeatCancel != nil {
		s.heartbeatCancel()
		<-s.heartbeatDone // 等待心跳 goroutine 退出
		s.heartbeatCancel = nil
		s.heartbeatDone = nil
	}
	if err := s.lock.ReleaseSchedulerLock(ctx); err != nil {
		if errors.Is(err, ErrLockNotHeld) {
			// 锁已丢失（心跳已检测到并停止 cron），Release 返回 ErrLockNotHeld 是预期行为，不告警
			logger.Log.Info("scheduler: lock already lost before release, skip")
		} else {
			logger.Log.Warn("scheduler: failed to release lock", zap.Error(err))
		}
	}
	s.lockOwned = false
}

// Stop 优雅停止 cron 调度，等待运行中的任务完成或超时。
// 若当前实例持有分布式锁，同时释放锁，允许其他实例接管。
// 若心跳已因锁丢失而停止 cron，则 Stop 跳过 stopCron 仅释放锁（幂等）。
func (s *Scheduler) Stop(ctx context.Context) error {
	if !s.cronStarted.Load() {
		// cron 未启动或已被心跳停止，仅释放锁（若持有）
		s.releaseLock(ctx)
		return nil
	}
	s.stopCron(ctx)
	s.cronStarted.Store(false)
	s.releaseLock(ctx)
	return nil
}

// stopCron 停止 cron 并等待运行中的任务完成，最多等待 5 秒。
func (s *Scheduler) stopCron(ctx context.Context) {
	stopCtx := s.cron.Stop()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	case <-timer.C:
		logger.Log.Warn("cron scheduler stop timeout after 5s")
	}
}

// AddFunc 注册一个 cron 任务，返回 entry ID。
// 供 Job 在 Init 中调用以注册自身的定时任务。
func (s *Scheduler) AddFunc(expr string, fn func()) (cron.EntryID, error) {
	return s.cron.AddFunc(expr, fn)
}

// Remove 移除一个 cron 任务。
// 供 Job 在 Reload 时移除旧的 entry。
func (s *Scheduler) Remove(id cron.EntryID) {
	s.cron.Remove(id)
}

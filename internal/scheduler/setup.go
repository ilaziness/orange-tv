package scheduler

import "github.com/ilaziness/orange-tv/internal/repository"

// Deps 收集所有调度任务所需的依赖。
// app 组装层构造后传入 Setup，由 scheduler 包内部创建并注册任务，
// 避免 app 层逐个拼装调度器。新增任务只需在此扩展。
type Deps struct {
	CollectRepo   repository.CollectRepository
	CollectRunner JobRunner
}

// Setup 在 scheduler 包内创建所有调度任务并注册到 Manager。
// 应在 Manager.Start 之前调用。
func Setup(mgr *Manager, deps Deps) {
	if deps.CollectRepo != nil && deps.CollectRunner != nil {
		mgr.Register(NewCollectScheduler(mgr, deps.CollectRepo, deps.CollectRunner))
	}
}

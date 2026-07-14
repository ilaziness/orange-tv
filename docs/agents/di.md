# 依赖注入说明

本项目使用**手动构造函数注入**。所有依赖在 `internal/app` 包中显式组装，不使用 Fx 或 Wire。

## 组装入口

| 文件 | 职责 |
|------|------|
| `internal/app/app.go` | `New()` 串联 infra/http；`Run()` 管理生命周期 |
| `internal/app/lifecycle.go` | `Hook`、`addHook`、`startHooks`、`stopHooks` |
| `internal/app/infra.go` | logger、db、cache、event、tracing、metrics、auth 等基础设施 |
| `internal/app/http.go` | 组装 `router.Handlers`、HTTPServer |
| `internal/server/http.go` | 接收 `*router.Handlers`，调用 `router.RegisterRoutes` 注册路由 |

## 生命周期 Hook

语义对齐 [uber-go/fx Lifecycle](https://pkg.go.dev/go.uber.org/fx#Lifecycle)（不引入 Fx 依赖）：

| 阶段 | 时机 | 顺序 |
|------|------|------|
| `OnStart` | `Run()` 内，日志与对外服务启动之前 | 注册顺序 FIFO |
| `OnStop` | `shutdown()` 内（服务器停止之后） | 注册顺序 LIFO |
| 启动失败回滚 | 某 `OnStart` 失败 | 对已成功的 Hook 逆序调用 `OnStop` |

`OnStart` / `OnStop` 均可为 nil：nil 则跳过对应阶段。可选 `Name` 字段用于日志与错误信息。

成对 Hook 仅在 `OnStart` 成功执行后才会在 `OnStop` 阶段清理；启动失败回滚与 `shutdown` 不会重复调用同一 Hook 的 `OnStop`。

### 注册方式

在 `wire*` 方法中通过 `addHook` 注册（见 `internal/app/lifecycle_example_test.go`）。

**成对初始化 + 清理**（后台 worker、订阅、预热）：

```go
worker := jobs.NewSyncWorker(a.db, a.log)
a.addHook(Hook{
    Name: "order_sync_worker",
    OnStart: func(ctx context.Context) error {
        a.log.Info("Starting sync worker")
        return worker.Start(ctx)
    },
    OnStop: func(ctx context.Context) error {
        a.log.Info("Stopping sync worker")
        return worker.Stop(ctx)
    },
})
```

**仅释放资源**（DB、Cache 等，构造在 `New()` 完成、无需启动逻辑）：

```go
a.addHook(Hook{
    Name:   "cache",
    OnStop: func(ctx context.Context) error { return a.cache.Close() },
})
```

### 约定

- 构造依赖放在 `New()` / `wire*()`；延迟到进程 run 的逻辑放 `OnStart`
- 有清理需求务必提供 `OnStop`；避免只有 `OnStart` 无 `OnStop`
- `New()` 失败时按 LIFO 执行已注册 Hook 的 `OnStop` 回滚

## 约定

- 各包只提供 `NewXxx(deps...)` 构造函数，通过参数声明依赖关系
- 业务依赖禁止全局变量；配置加载后会有 `config.Cfg`，但业务层应通过构造函数注入获取依赖
- 新增业务模块时，在对应 `wire*` 方法中追加组装；不使用 Fx/Wire，也不新增 `module.go`

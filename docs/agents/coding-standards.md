# 编码规范（详情）

> 通用核心规则见 [AGENTS.md](../../AGENTS.md#核心规则始终遵守)。本文档补充数据库、配置、运维端点与基础设施约定。

## 数据库规范

必须使用 Bun 官方组件：

- `github.com/uptrace/bun` - 核心 ORM
- `github.com/uptrace/bun/dialect/mysqldialect` - MySQL 方言
- `github.com/uptrace/bun/dialect/pgdialect` - PostgreSQL 方言
- `github.com/uptrace/bun/dialect/sqlitedialect` - SQLite 方言
- `github.com/uptrace/bun/driver/sqliteshim` - SQLite 驱动
- `github.com/uptrace/bun/driver/pgdriver` - PostgreSQL 驱动
- `github.com/go-sql-driver/mysql` - MySQL 驱动
- `github.com/uptrace/bun/extra/bundebug` - 调试钩子

## 配置管理

加载逻辑见 `internal/config/config.go` 的 `LoadWithEnv`：

1. **环境判定**：`-e/--env` > 环境变量 `APP_ENV` > 默认 `dev`
2. **dotenv**：按环境加载 `.env.{env}`、`configs/.env.{env}`、`.env`、`configs/.env`（文件缺失不报错）
3. **主配置**：指定 `-c/--config` 时加载该文件；未指定时加载 `configs/config.yaml`
4. **环境覆盖**：合并 `configs/config.{env}.yaml`（存在则 merge）
5. **敏感字段**：`DATABASE_PASSWORD`、`REDIS_PASSWORD` 通过 Viper 绑定环境变量覆盖

支持的环境名：`dev`、`prod`、`test`。

## 健康检查端点

注册于 `internal/router/system.go`；路径常量与 `SystemPaths` 见 `internal/router/paths.go`：

| 端点 | 语义 | 检查依赖 |
|------|------|----------|
| `GET /health` | 轻量存活探测 | 否（委托 `/liveness`） |
| `GET /liveness` | 存活探测 | 否 |
| `GET /readiness` | 就绪探测 | 是（通过 `health.Checker`） |
| `GET /version` | 应用版本 | 否 |

响应使用统一 `{code, message, data}` envelope；K8s/Docker HTTP 探测仅依赖状态码（2xx/503）。

**扩展依赖检查**：在 `internal/app/http.go` 向 `HealthHandler` 注册 checker，实现 `internal/health.Checker`（`Check(ctx context.Context) error`）。示例：`health.NewDatabaseChecker(db)`。多个 checker 顺序执行并共享 `readinessCheckTimeout`（默认 5s）；依赖较多或较慢时可增大超时，或对单个 checker 使用独立 deadline，必要时用 `errgroup` 并行。

**`/health` 响应变更**：`data.status` 为 `"alive"`（与 `/liveness` 一致），不再返回 `"ok"`。

**运行时指标**：启用 `metrics.enabled` 后使用 `GET /metrics`（Prometheus），不再提供 `/stats`。

**CLI**：`orange-tv health` 默认探测 `/readiness`，供 Docker `HEALTHCHECK` 使用；`--probe` 可选 `health`/`liveness`/`readiness`。

## 缓存和事件系统

- **缓存**：内存（Ristretto）、Redis、多级缓存 — 示例见 `internal/cache/example_test.go`
- **事件总线**：发布/订阅，支持同步/异步、优先级、通配符 — 示例见 `internal/event/example_test.go`

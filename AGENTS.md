# agents.md — Orange TV 影视系统 AI 编码指南

> 本文件仅保留每次任务都需要的关键信息。详细内容按主题拆分至 [`docs/agents/`](docs/agents/README.md)，**仅在需要时阅读对应文档**，避免重复加载。

## 项目概述

影视系统后端应用，模块路径 `github.com/ilaziness/orange-tv`，Go 1.26.4。

| 类别 | 技术 |
| ------ | ------ |
| Web | Gin |
| 依赖组装 | 手动构造函数注入（`internal/app`） |
| CLI | Cobra |
| 配置 | Viper |
| 日志 | Zap + Lumberjack |
| ORM | Bun（MySQL / PostgreSQL / SQLite） |
| 迁移 | 内置 migrate 命令 |

支持 HTTP 服务，通过 `configs/config.yaml` 中 `enabled` 控制。

开发默认 **MySQL**；迁移为 MySQL DDL；业务模型优先用 `orange-tv gen model` 从库表生成。
API 路径：用户端 `/api/client/v1`、`/api/client/v2`，管理端 `/api/admin/v1`、`/api/admin/v2`，内网 `/api/internal/v1`，开放 `/api/open/v1`。
前端 monorepo：`web/apps/client`、`web/apps/admin`、`web/packages/shared`。

## 核心规则（始终遵守）

1. **格式化与检查**：`go fmt ./...`、`go vet ./...` 必须通过
2. **错误处理**：不忽略 error；向上传递用 `fmt.Errorf("...: %w", err)`
3. **日志**：使用 `zap.Logger`，禁止 `fmt.Println` / `log.Println`，关键位置需要添加合适的日志
4. **错误码**：`internal/errcode`，格式 `{3位模块码}{4位业务码}`（100 通用 / 200 用户 / 300 认证 / 400 内容 / 900 系统）
5. **依赖注入**：构造函数注入，仅在 `internal/app` 组装，禁止全局变量
6. **数据库**：必须使用 Bun 官方包（见 [coding-standards.md](docs/agents/coding-standards.md)）
6.1 **分层分包**：DTO / Service / Handler 按 API 面分子包（`admin`、`client`、`open`）；共享类型放根包；路由只保留版本前缀常量，业务路径注册时写字符串
7. **代码设计**：遵循高内聚、低耦合原则，保持模块职责单一、边界清晰，避免跨层依赖和隐式耦合
8. **复用与扩展**：优先抽象稳定且通用的能力，避免重复实现；通过接口、组合和依赖注入支持复用与扩展，避免过度设计
9. **可读可维护**：代码应简洁、清晰、易理解，命名准确，控制函数和模块复杂度；以易编写、易测试、易维护为质量目标
10. **修改后验证**：`make fmt && make vet && make build && make test`（详见 [verification.md](docs/agents/verification.md)）
11. **外部输入边界验证**：编写业务逻辑时必须验证外部输入（HTTP 参数/Body、TCP/UDP payload、数据库/缓存读取、外部接口返回等）的边界条件（空值、范围、长度、枚举、越界等）；无效输入应返回对应 errcode，禁止未经验证的数据进入后续处理
12. `internal\service`下面的`admin`，`client`， `open`对应三端不能互相`import`逻辑，如果是需要多端通用使用的写到包`internal\service`下面，然后再具体端的包里面import调用即可
13. api接口需要在`handler`函数添加`swagger`注释，方便生成api文档
14. 更新新增数据不需要填充`updated_at`和`created_at`通过模型操作会按需自动加上这两个字段

## 前端项目指南

修改前端项目才需要，后端项目忽略，指南文件：`web/AGENTS.md`。

## 按需查阅

| 主题 | 文件 | 何时阅读 |
| ------ | ------ | ---------- |
| 文档索引 | [docs/agents/README.md](docs/agents/README.md) | 查找全部 Agent 专题文档 |
| 目录结构 | [docs/agents/structure.md](docs/agents/structure.md) | 不确定代码应放在哪个包 |
| 编码规范详情 | [docs/agents/coding-standards.md](docs/agents/coding-standards.md) | 数据库、配置、健康检查、缓存/事件 |
| 扩展 HTTP 模块 | [docs/agents/extend-http-module.md](docs/agents/extend-http-module.md) | 新增 Model/Service/Handler/路由 |
| 依赖注入 | [docs/agents/di.md](docs/agents/di.md) | 理解 `internal/app` 组装与生命周期 |
| 验证流程 | [docs/agents/verification.md](docs/agents/verification.md) | 提交前完整检查步骤 |
| 常用命令 | [docs/agents/commands.md](docs/agents/commands.md) | migrate、build、test 等 CLI |
| 模块启用/删除 | [docs/module-usage.md](docs/module-usage.md) | 只保留 HTTP |
| 可观测性 | [docs/observability.md](docs/observability.md) | 日志、指标、链路追踪 |
| 部署 | [docs/deployment.md](docs/deployment.md) | 构建与部署流程 |

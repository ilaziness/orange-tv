# 小橘TV 技术文档

> 本文件汇总项目的技术性说明，面向开发者与运维人员。业务介绍请参阅根目录 [README.md](../README.md)。

## 技术栈

| 类别 | 技术 |
| ------ | ------ |
| 语言 | Go 1.26.4 |
| Web | Gin |
| 依赖组装 | 手动构造函数注入（`internal/app`） |
| CLI | Cobra |
| 配置 | Viper |
| 日志 | Zap + Lumberjack |
| ORM | Bun（MySQL / PostgreSQL / SQLite） |
| 迁移 | 内置 migrate 命令 |
| 可观测性 | OpenTelemetry（链路追踪）、Prometheus（指标） |
| API 文档 | Swagger |

支持 HTTP 服务，通过 `configs/config.yaml` 中 `enabled` 控制。

开发默认 **MySQL**；迁移为 MySQL DDL；业务模型优先用 `orange-tv gen model` 从库表生成。

## 功能特性

- **模块化设计**：按需启用所需功能
- **依赖注入**：手动构造函数注入，仅在 `internal/app` 组装
- **配置管理**：YAML/JSON 配置，仅敏感数据使用环境变量覆盖
- **结构化日志**：Zap 日志，支持日志轮转
- **优雅关闭**：信号处理和资源清理
- **健康检查**：`/health`、`/readiness`、`/liveness` 端点
- **CLI 工具**：基于 Cobra 构建
- **数据库集成**：Bun ORM，支持 MySQL、PostgreSQL、SQLite
- **数据库迁移**：内置迁移工具管理数据库架构
- **代码生成**：从数据库表生成模型
- **可观测性**：可配置的分布式追踪（OpenTelemetry）和指标（Prometheus）

## 快速开始

### 前置要求

- Go 1.26.4 或更高版本

### 安装

```bash
# 克隆仓库
git clone https://github.com/ilaziness/orange-tv.git
cd orange-tv

# 下载依赖
make deps
```

### 运行应用

```bash
# 使用默认配置运行
make run

# 使用开发配置运行
make run-dev

# 或先构建再运行
make build
./build/orange-tv serve
```

### 可用命令

```bash
# 使用默认配置启动服务
./build/orange-tv serve

# 使用指定环境启动（dev/prod/test）
./build/orange-tv serve -e dev
./build/orange-tv serve -e prod
./build/orange-tv serve -e test

# 使用指定配置文件启动
./build/orange-tv serve -c configs/config.prod.yaml

# 显示版本信息
./build/orange-tv version

# 验证配置
./build/orange-tv config validate -e dev
./build/orange-tv config validate -c configs/config.prod.yaml

# 显示当前配置
./build/orange-tv config show -e dev
./build/orange-tv config show -c configs/config.prod.yaml

# 数据库迁移
./build/orange-tv migrate up              # 执行所有待执行的迁移
./build/orange-tv migrate down            # 回滚最后一次迁移
./build/orange-tv migrate status          # 显示迁移状态
./build/orange-tv migrate create <name>   # 创建新的迁移文件
./build/orange-tv migrate up --dry-run    # 预览迁移而不执行

# 代码生成
./build/orange-tv gen model               # 从数据库生成模型
./build/orange-tv gen model --table users --output ./internal/model

# 创建管理员账号
./build/orange-tv admin create --username <name> --password <pass>

# 显示帮助信息
./build/orange-tv --help
```

## 配置

配置可以通过以下方式提供：

1. **YAML 文件**：`configs/config.yaml`（基础配置）、`config.dev.yaml`、`config.prod.yaml`、`config.test.yaml`
2. **命令行**：`--env dev|prod|test` 或 `--config <path>`
3. **环境变量**：仅用于敏感数据（密码、API 密钥）

### 配置优先级

`--config` > `--env` > `config.yaml` 默认值

### 配置文件

- `config.yaml` - 基础配置
- `config.dev.yaml` - 开发环境（使用 `--env dev`）
- `config.prod.yaml` - 生产环境（使用 `--env prod`）
- `config.test.yaml` - 测试环境（使用 `--env test`）

### 环境变量

环境变量仅用于敏感数据：

```bash
# 数据库配置（默认 MySQL）
export DATABASE_DRIVER=mysql
export DATABASE_HOST=127.0.0.1
export DATABASE_PORT=3306
export DATABASE_USER=orange
export DATABASE_PASSWORD=orange_password
export DATABASE_DATABASE=orange_tv

# Redis 配置
export REDIS_ENABLED=false
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
```

查看 `.env.example` 获取所有可用的环境变量。

## 数据库集成

使用 Bun ORM，**开发与生产默认 MySQL**。驱动层仍支持 PostgreSQL 与 SQLite（单元测试可用 SQLite）。

### 支持的数据库

- **MySQL**：开发与生产默认
- **PostgreSQL**：可选
- **SQLite**：可选 / 单元测试

### 数据库配置

在配置文件中配置数据库：

```yaml
database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: orange_tv
  user: orange
  password: orange_password
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 300
```

### 执行迁移

```bash
# 创建新迁移
./build/orange-tv migrate create add_users_table

# 执行迁移
./build/orange-tv migrate up

# 检查迁移状态
./build/orange-tv migrate status

# 回滚最后一次迁移
./build/orange-tv migrate down
```

### 生成模型

```bash
# 从所有表生成模型
./build/orange-tv gen model

# 为特定表生成模型
./build/orange-tv gen model --table users

# 自定义输出
./build/orange-tv gen model --table users --output ./internal/model --package model
```

## 项目结构

```text
.
├── cmd/                    # 命令行接口
│   ├── root.go            # 根命令
│   ├── serve.go           # 服务命令
│   ├── config.go          # 配置管理命令
│   ├── version.go         # 版本命令
│   ├── migrate.go         # 数据库迁移命令
│   └── gen.go             # 代码生成命令
├── configs/               # 配置文件
├── migrations/            # 数据库迁移文件
│   ├── *.up.sql          # 升级迁移文件
│   └── *.down.sql        # 降级迁移文件
├── internal/              # 私有应用代码
│   ├── app/              # 应用组装与生命周期
│   ├── config/           # 配置结构
│   ├── constant/         # 应用常量
│   ├── database/         # 数据库初始化
│   ├── errcode/          # 错误码定义
│   ├── handler/          # 请求处理器
│   │   ├── http/         # HTTP 处理器
│   │   │   ├── user.go   # 用户处理器
│   │   │   └── health.go # 健康检查处理器
│   ├── logger/           # 日志封装
│   ├── middleware/       # 中间件
│   │   └── http/         # HTTP 专用中间件
│   ├── response/         # API 响应结构
│   ├── router/           # 路由注册
│   ├── server/           # 服务器实现
│   │   └── http.go       # HTTP 服务器
│   ├── service/          # 业务逻辑层
│   │   └── user.go       # 用户服务
│   ├── repository/       # 数据访问层
│   │   └── user.go       # 用户仓储
│   ├── model/            # 数据模型
│   │   └── user.go       # 用户模型
│   └── dto/              # 数据传输对象
│       └── user.go       # 用户 DTO
├── web/                   # 前端 monorepo
│   ├── apps/client/      # 用户端
│   ├── apps/admin/       # 管理后台
│   └── packages/shared/  # 共享包
├── main.go               # 应用入口
├── Makefile              # 构建命令
├── .env.example          # 环境变量示例
└── README.md             # 业务介绍（中文主文档）
```

### API 路径约定

- 用户端：`/api/client/v1`、`/api/client/v2`
- 管理端：`/api/admin/v1`、`/api/admin/v2`
- 内网：`/api/internal/v1`
- 开放：`/api/open/v1`

## 错误码

错误码遵循 `{3位模块码}{4位业务码}` 格式：

- `100xxxx` - 通用模块（参数错误、数据未找到等）
- `200xxxx` - 用户模块（用户不存在、用户已存在等）
- `300xxxx` - 认证模块（认证失败、令牌过期、权限不足等）
- `400xxxx` - 内容模块
- `900xxxx` - 系统模块（内部错误、数据库错误、缓存错误等）

示例：

- `1000001` - 参数错误
- `1000002` - 数据未找到
- `2000001` - 用户不存在
- `3000001` - 认证失败
- `3000003` - 权限不足
- `9000001` - 内部服务器错误

每个错误码都包含关联的 HTTP 状态码。

## 开发

### Make 命令

```bash
make build          # 构建应用
make run            # 运行应用
make run-dev        # 使用开发配置运行
make test           # 运行测试
make test-coverage  # 运行测试并生成覆盖率报告
make clean          # 清理构建产物
make deps           # 下载依赖
make lint           # 运行代码检查
make fmt            # 格式化代码
make vet            # 运行 go vet
```

### 修改后验证

```bash
make fmt && make vet && make build && make test
```

详见 [验证流程](agents/verification.md)。

### 健康检查端点

服务器启动后，可以访问以下端点：

- `GET /health` - 基础健康检查
- `GET /readiness` - 就绪检查（包含依赖检查）
- `GET /liveness` - 存活检查
- `GET /version` - 应用版本
- `GET /metrics` - Prometheus 指标（需 `metrics.enabled: true`）

示例：

```bash
curl http://localhost:8080/health
```

## 可观测性

应用支持可配置的分布式追踪和指标监控。详细文档请参阅 [可观测性指南](observability.md)。

### 可观测性特性

- **分布式追踪**：OpenTelemetry with OTLP 协议（支持 Jaeger、Tempo 等）
- **指标监控**：Prometheus 原生客户端，包含 HTTP、数据库、Redis 指标
- **数据关联**：trace_id 自动注入到日志和 HTTP 响应头

### 启用可观测性

在 `configs/config.yaml` 中启用可观测性：

```yaml
# 启用分布式追踪
tracing:
  enabled: true
  endpoint: localhost:4317  # OTLP gRPC 端点
  sample_rate: 1.0

# 启用指标监控
metrics:
  enabled: true
  path: /metrics
  labels:
    env: dev
    version: "1.0.0"
```

访问指标端点：<http://localhost:8080/metrics>

详细配置和用法请参阅 [可观测性指南](observability.md)。

## API 文档

应用集成了 Swagger API 文档，访问以下地址查看：

- Swagger UI: <http://localhost:8080/swagger/index.html>

### 生成 API 文档

```bash
# 生成 Swagger 文档
make swagger

# 清理生成的文档
make swagger-clean
```

## 示例代码

项目包含多个示例代码，展示如何使用各种组件：

- **HTTP 服务**：`/api/v1/users/{id}` 等接口是完整的 HTTP 服务示例
- `internal/event/example_test.go` - 事件系统使用示例
- `internal/cache/example_test.go` - 缓存使用示例

## 相关文档

- [业务介绍](../README.md) - 面向用户的业务说明
- [Agent 编码指南](../AGENTS.md) - AI 编码规范总览
- [Agent 专题文档](agents/README.md) - 目录结构、编码规范、依赖注入等
- [模块使用说明](module-usage.md) - 如何选择和删除不需要的服务模块
- [部署文档](deployment.md) - 单实例、多实例和 Docker 部署指南
- [可观测性指南](observability.md) - 分布式追踪和指标监控配置

## License

MIT License

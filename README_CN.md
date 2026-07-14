# App Template

一个灵活的 Go 应用模板框架，支持多种服务类型和可选组件集成，适合快速构建各种后端服务。

## 功能特性

- **模块化设计**：按需启用所需功能
- **依赖注入**：使用 Uber Fx
- **配置管理**：YAML/JSON 配置，仅敏感数据使用环境变量覆盖
- **结构化日志**：Zap 日志，支持日志轮转
- **优雅关闭**：信号处理和资源清理
- **健康检查**：/health、/readiness、/liveness 端点
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
# 数据库配置
export DATABASE_ENABLED=true
export DATABASE_DRIVER=sqlite
export DATABASE_DATABASE=./data/app.db
export DATABASE_USER=
export DATABASE_PASSWORD=

# Redis 配置
export REDIS_ENABLED=false
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
```

查看 `.env.example` 获取所有可用的环境变量。

## 数据库集成

本模板包含使用 Bun ORM 的数据库集成，支持 MySQL、PostgreSQL 和 SQLite。

### 支持的数据库

- **SQLite**：默认用于开发和测试
- **MySQL**：用于生产环境
- **PostgreSQL**：用于生产环境

### 数据库配置

在配置文件中配置数据库：

```yaml
database:
  enabled: true
  driver: sqlite
  host: ""
  port: 0
  database: ./data/app.db
  user: ""
  password: ""
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
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

```
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
├── main.go               # 应用入口
├── Makefile              # 构建命令
├── .env.example          # 环境变量示例
└── README.md             # 本文件
```

## 错误码

错误码遵循 `{3位模块码}{4位业务码}` 格式：

- `100xxxx` - 通用模块（参数错误、数据未找到等）
- `200xxxx` - 用户模块（用户不存在、用户已存在等）
- `300xxxx` - 认证模块（认证失败、令牌过期、权限不足等）
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

应用支持可配置的分布式追踪和指标监控。详细文档请参阅 [可观测性指南](docs/observability.md)。

### 功能特性

- **分布式追踪**：OpenTelemetry with OTLP 协议（支持 Jaeger、Tempo 等）
- **指标监控**：Prometheus 原生客户端，包含 HTTP、数据库、Redis 指标
- **数据关联**：trace_id 自动注入到日志和 HTTP 响应头

### 快速开始

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

访问指标端点：http://localhost:8080/metrics

详细配置和用法请参阅 [可观测性指南](docs/observability.md)。

## API 文档

应用集成了 Swagger API 文档，访问以下地址查看：

- Swagger UI: http://localhost:8080/swagger/index.html

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

## 文档

- [模块使用说明](docs/module-usage.md) - 如何选择和删除不需要的服务模块
- [部署文档](docs/deployment.md) - 单实例、多实例和 Docker 部署指南
- [可观测性指南](docs/observability.md) - 分布式追踪和指标监控配置

## 许可证

MIT License

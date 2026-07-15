# 模块使用说明

本文档说明项目模块结构和可选模块的启用/禁用方式。

## 模块化设计

本项目采用模块化设计，当前支持：

- **HTTP 服务**：提供 HTTP/HTTPS RESTful API 服务

通过 `configs/config.yaml` 中 `enabled` 字段控制各模块启用/禁用。

## 通过配置文件启用/禁用服务

在 `configs/config.yaml` 中配置：

```yaml
http:
  enabled: true   # 启用 HTTP 服务
  host: 0.0.0.0
  port: 8080
```

## 模块依赖关系

### HTTP 服务依赖

- `internal/handler/http/` - 公共绑定工具、health、stub
- `internal/handler/http/admin/` - 管理端处理器
- `internal/handler/http/client/` - 用户端处理器
- `internal/server/http.go` - HTTP 服务器
- `internal/router/` - 路由注册（`Handlers` 聚合 + `RegisterRoutes`；业务路径直接写字符串）
- `internal/middleware/http/` - HTTP 中间件
- `internal/service/admin`、`internal/service/client` - 业务逻辑
- `internal/repository/`、`internal/model/` - 数据访问与模型

### 共享层

以下层被各服务类型共享，不要删除：

- `internal/service/`（含 `admin/`、`client/`）- 业务逻辑层
- `internal/repository/` - 数据访问层
- `internal/model/` - 数据模型
- `internal/dto/`（含 `admin/`、`client/`）- 数据传输对象
- `internal/config/` - 配置管理
- `internal/database/` - 数据库连接
- `internal/cache/` - 缓存
- `internal/event/` - 事件系统
- `internal/logger/` - 日志
- `internal/errcode/` - 错误码定义
- `internal/response/` - 统一响应

## 可选模块

### 数据库

通过 `database.enabled` 控制。禁用时移除 `internal/database/` 相关代码。

### Redis

通过 `redis.enabled` 控制。禁用时移除 `internal/cache/` 中 Redis 相关代码。

### 链路追踪

通过 `tracing.enabled` 控制。禁用时移除 `internal/tracing/` 相关代码。

### 指标监控

通过 `metrics.enabled` 控制。禁用时移除 `internal/metrics/` 相关代码。

## 删除模块后的清理步骤

### 1. 更新 go.mod

删除模块后，运行以下命令清理未使用的依赖：

```bash
go mod tidy
```

### 2. 更新依赖组装

在 `internal/app/` 中，删除或注释对应模块的 wiring 逻辑。

### 3. 更新文档

更新 README.md 和其他文档，移除对已删除模块的说明。

## 示例配置

```yaml
# configs/config.yaml
app:
  name: orange-tv
  version: 1.0.0

http:
  enabled: true
  host: 0.0.0.0
  port: 8080

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: orange_tv
  user: orange
  password: orange_password

redis:
  enabled: false
```

## 总结

- 通过配置文件控制模块启用/禁用
- 删除不需要的模块可以减小应用体积
- 删除模块后需要清理依赖和配置
- 共享层（service、repository、model）不要删除

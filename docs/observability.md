# 可观测性配置指南

本项目支持通过配置文件启用/禁用分布式链路追踪和指标监控功能。

## 功能概述

- **链路追踪**：使用 OpenTelemetry + OTLP 协议，支持 Jaeger、Tempo 等后端
- **指标监控**：使用 Prometheus native client，提供 HTTP、数据库、Redis 等指标
- **数据关联**：trace_id 自动注入到日志和 HTTP 响应头，实现链路追踪与日志关联

## 配置说明

### 链路追踪配置

在 `configs/config.yaml` 中配置：

```yaml
tracing:
  enabled: false              # 是否启用链路追踪，默认 false
  endpoint: localhost:4317   # OTLP gRPC 端点地址
  sample_rate: 1.0            # 采样率（0-1），1.0 表示全采样
```

**支持的 OTLP 后端**：
- Jaeger Collector: `localhost:4317`
- Grafana Tempo: `localhost:4317`
- 其他支持 OTLP 的后端

**采样率建议**：
- 开发环境：`1.0`（全采样）
- 生产环境：`0.1`（10% 采样，降低性能影响）

### 指标监控配置

在 `configs/config.yaml` 中配置：

```yaml
metrics:
  enabled: false              # 是否启用指标监控，默认 false
  path: /metrics              # Prometheus 抓取端点路径
  labels:                     # 全局标签
    env: dev
    version: "1.0.0"
```

**预定义指标**：
- HTTP：请求总数、请求耗时、并发请求数
- 数据库：活跃连接数、空闲连接数、查询耗时
- Redis：缓存命中/未命中、操作耗时
- Go runtime：内存、GC、goroutine 等
- 进程：CPU、文件描述符等

## 使用示例

### 启用链路追踪

1. 修改 `configs/config.yaml`：

```yaml
tracing:
  enabled: true
  endpoint: localhost:4317
  sample_rate: 1.0
```

2. 启动 Jaeger Collector（使用 Docker）：

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest
```

3. 启动应用后，访问 Jaeger UI：http://localhost:16686

### 启用指标监控

1. 修改 `configs/config.yaml`：

```yaml
metrics:
  enabled: true
  path: /metrics
  labels:
    env: dev
    version: "1.0.0"
```

2. 启动应用后，访问指标端点：http://localhost:8080/metrics

指标端点包含 Go runtime（goroutine、内存、GC）与进程指标，可替代原 `/stats` 端点。健康探测（`/health`、`/liveness`、`/readiness`）不计入 `http_requests_total`，避免污染业务指标。

3. 配置 Prometheus 抓取：

```yaml
scrape_configs:
  - job_name: 'orange-tv'
    static_configs:
      - targets: ['localhost:8080']
```

### 数据关联

启用链路追踪后，trace_id 会自动注入到：

1. **HTTP 响应头**：`X-Trace-ID`
2. **日志字段**：`trace_id`

示例日志：

```json
{
  "level": "info",
  "msg": "Request processed",
  "request_id": "abc123",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "timestamp": "2026-05-23T10:00:00Z"
}
```

## 性能影响

- **链路追踪**：采样率可配置，生产环境建议 0.1，性能影响 < 5%
- **指标监控**：内存占用约 10-20MB，CPU 影响可忽略
- **禁用时**：零性能影响，相关依赖不加载

## 故障排查

### 链路追踪未上报

1. 检查 endpoint 配置是否正确
2. 检查 OTLP 后端是否正常运行
3. 查看应用日志是否有 exporter 错误

### 指标端点返回 503

1. 检查 `metrics.enabled` 是否为 true
2. 检查配置文件是否正确加载

### trace_id 未注入

1. 确认 `tracing.enabled` 为 true
2. 检查 tracing 中间件是否正确注册

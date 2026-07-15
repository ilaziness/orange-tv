# 部署文档

本文档提供应用的部署说明，包括单实例部署、多实例部署和 Docker 部署。

## 前置要求

### 系统要求

- Go 1.26.4 或更高版本
- Linux/macOS/Windows 操作系统
- 至少 100MB 可用内存（空闲状态）

### 依赖服务（可选）

- **数据库**（三选一）：
  - MySQL 5.7+
  - PostgreSQL 12+
  - SQLite 3+（无需额外安装）
- **Redis**（可选，用于分布式限流和缓存）：
  - Redis 6.0+

## 生产建议（第四阶段）

- **JWT**：通过环境变量设置 `JWT_SECRET`（或配置 `jwt.secret`），勿使用开发默认。
- **缓存**：建议 `cache.enabled: true`；单实例可用 `memory`，多实例用 `redis`/`multi`。
- **限流**：建议生产开启 `rate_limit.enabled`；登录接口额外始终启用 IP 限流（5 rps）。
- **资源站**：若对外提供采集，务必配置 `resource_api_key`，并按需关闭 `enable_third_party_collect`。
- **设置种子**：执行迁移后会写入站点/API 相关 `system_settings` 键。

## 单实例部署

### 1. 编译应用

```bash
# 克隆仓库
git clone https://github.com/ilaziness/orange-tv.git
cd orange-tv

# 下载依赖
make deps

# 编译应用
make build
```

编译后的二进制文件位于 `build/orange-tv`。

### 2. 配置文件准备

复制配置文件示例并根据环境修改：

```bash
cp configs/config.yaml configs/config.prod.yaml
```

编辑 `configs/config.prod.yaml`：

```yaml
app:
  name: my-app
  version: 1.0.0

http:
  enabled: true
  host: 0.0.0.0
  port: 8080

database:
  enabled: true
  driver: mysql  # 或 postgresql/sqlite
  host: localhost
  port: 3306
  database: myapp
  user: myapp
  password: your_password

redis:
  enabled: true
  host: localhost
  port: 6379
  password: ""
```

### 3. 环境变量配置

创建 `.env` 文件（仅用于敏感信息）：

```bash
cp .env.example .env
```

编辑 `.env`：

```bash
DATABASE_PASSWORD=your_secure_password
REDIS_PASSWORD=your_redis_password
```

### 4. 数据库迁移

```bash
# 执行数据库迁移
./build/orange-tv migrate up -c configs/config.prod.yaml

# 查看迁移状态
./build/orange-tv migrate status -c configs/config.prod.yaml
```

### 5. 启动应用

```bash
# 使用生产配置启动
./build/orange-tv serve -c configs/config.prod.yaml
```

### 6. 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# 就绪检查
curl http://localhost:8080/readiness

# 版本信息
curl http://localhost:8080/version
```

## 多实例部署

多实例部署需要使用负载均衡器和共享存储（Redis）。

### 1. 负载均衡配置

使用 Nginx 作为负载均衡器：

```nginx
upstream app_backend {
    server 192.168.1.10:8080;
    server 192.168.1.11:8080;
    server 192.168.1.12:8080;
}

server {
    listen 80;
    server_name myapp.example.com;

    location / {
        proxy_pass http://app_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 2. 共享存储配置

所有实例必须使用同一个数据库和 Redis。

**配置文件**（所有实例相同）：

```yaml
database:
  enabled: true
  driver: mysql
  host: db.example.com  # 共享数据库
  port: 3306
  database: myapp
  user: myapp
  password: your_password

redis:
  enabled: true
  host: redis.example.com  # 共享 Redis
  port: 6379
  password: your_redis_password
```

### 3. 分布式限流配置

启用 Redis 限流以实现多实例间的限流共享：

```yaml
ratelimit:
  enabled: true
  store: redis  # 使用 Redis 存储限流状态
  global_rps: 10000
  ip_rps: 100
  user_rps: 50
```

### 4. 启动多个实例

在每个服务器上：

```bash
# 编译应用
make build

# 启动应用
./build/orange-tv serve -c configs/config.prod.yaml
```

### 5. 健康检查配置

在负载均衡器中配置健康检查：

```nginx
upstream app_backend {
    server 192.168.1.10:8080 max_fails=3 fail_timeout=30s;
    server 192.168.1.11:8080 max_fails=3 fail_timeout=30s;
    server 192.168.1.12:8080 max_fails=3 fail_timeout=30s;
}
```

## Docker 部署

### 1. 构建镜像

```bash
# 构建镜像
docker build -t orange-tv:1.0.0 .
```

### 2. 运行容器

```bash
# 运行容器
docker run -d \
  --name orange-tv \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/data:/app/data \
  -e DATABASE_PASSWORD=your_password \
  orange-tv:1.0.0
```

### 3. 环境变量配置

通过 `-e` 参数传递环境变量：

```bash
docker run -d \
  --name orange-tv \
  -p 8080:8080 \
  -e DATABASE_ENABLED=true \
  -e DATABASE_DRIVER=mysql \
  -e DATABASE_HOST=db \
  -e DATABASE_PORT=3306 \
  -e DATABASE_DATABASE=myapp \
  -e DATABASE_USER=myapp \
  -e DATABASE_PASSWORD=your_password \
  -e REDIS_ENABLED=true \
  -e REDIS_HOST=redis \
  -e REDIS_PORT=6379 \
  orange-tv:1.0.0
```

### 4. 数据持久化

使用数据卷持久化数据库文件：

```bash
docker run -d \
  --name orange-tv \
  -p 8080:8080 \
  -v app-data:/app/data \
  orange-tv:1.0.0
```

### 5. 健康检查

Docker 会自动执行容器内定义的健康检查：

```bash
# 查看容器健康状态
docker inspect --format='{{.State.Health.Status}}' orange-tv
```

## Docker Compose 部署

### 1. 使用 docker-compose.yml

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 2. 服务编排

docker-compose.yml 包含以下服务：

- **app**：主应用
- **mysql**：MySQL 数据库（可选）
- **postgres**：PostgreSQL 数据库（可选）
- **redis**：Redis 缓存（可选）

### 3. 网络配置

所有服务在同一个 Docker 网络中，可以通过服务名互相访问：

```yaml
# 应用配置
database:
  host: mysql  # 使用服务名
  port: 3306

redis:
  host: redis  # 使用服务名
  port: 6379
```

### 4. 数据卷配置

数据卷用于持久化数据：

```yaml
volumes:
  mysql-data:
  postgres-data:
  redis-data:
  app-data:
```

## 健康检查配置

### 应用健康检查端点

- `GET /health` - 基础健康检查
- `GET /readiness` - 就绪检查（包含依赖检查）
- `GET /liveness` - 存活检查
- `GET /version` - 应用版本
- `GET /metrics` - Prometheus 指标（需 `metrics.enabled: true`）

### 负载均衡器健康检查

配置负载均衡器定期检查 `/readiness` 端点：

```bash
# 每 10 秒检查一次
curl http://localhost:8080/readiness
```

### Kubernetes 健康检查

```yaml
livenessProbe:
  httpGet:
    path: /liveness
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readiness
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## 日志收集配置

### 日志文件位置

默认日志文件位于 `logs/` 目录：

```
logs/
├── app.log          # 应用日志
├── app-error.log    # 错误日志
└── app-2024-01-01.log  # 按日期轮转的日志
```

### 日志轮转配置

在 `configs/config.yaml` 中配置：

```yaml
logger:
  level: info
  output: file
  filename: logs/app.log
  max_size: 100      # 单个文件最大 100MB
  max_backups: 30    # 保留最近 30 个文件
  max_age: 30        # 保留 30 天内的日志
  compress: true     # 压缩旧日志
```

### 日志收集工具

推荐使用以下工具收集日志：

- **Filebeat**：发送到 Elasticsearch
- **Fluentd**：发送到 Elasticsearch 或其他存储
- **Loki**：Grafana 日志系统

## 监控告警配置建议

### Prometheus 指标

启用 Prometheus 指标导出：

```yaml
metrics:
  enabled: true
  path: /metrics
  labels:
    env: prod
    version: "1.0.0"
```

访问指标端点：

```bash
curl http://localhost:8080/metrics
```

### Grafana 仪表板

建议监控以下指标：

- **应用指标**：
  - HTTP 请求总数（按状态码、路径、方法分类）
  - 请求响应时间（P50、P90、P95、P99）
  - 请求错误率
  - 并发请求数
  - Goroutine 数量

- **数据库指标**：
  - 连接池状态（活跃连接、空闲连接、等待连接）
  - 查询执行时间
  - 查询总数

- **Redis 指标**：
  - 缓存命中率
  - 连接池状态
  - 操作耗时
  - 操作总数

### 告警规则

建议配置以下告警：

- 应用错误率超过 5%
- P99 响应时间超过 1 秒
- 数据库连接池耗尽
- Redis 连接失败
- Goroutine 数量超过 10000
- 内存使用超过 80%

## 常见部署问题排查

### 问题 1：应用无法启动

**检查步骤**：
1. 查看日志文件 `logs/app-error.log`
2. 检查配置文件语法是否正确
3. 验证数据库连接配置
4. 检查端口是否被占用

```bash
# 检查端口占用
netstat -tuln | grep 8080
```

### 问题 2：数据库连接失败

**检查步骤**：
1. 验证数据库服务是否运行
2. 检查数据库用户名和密码
3. 验证网络连接
4. 检查数据库防火墙规则

```bash
# 测试数据库连接
mysql -h localhost -u myapp -p
```

### 问题 3：Redis 连接失败

**检查步骤**：
1. 验证 Redis 服务是否运行
2. 检查 Redis 密码配置
3. 验证网络连接
4. 检查 Redis 防火墙规则

```bash
# 测试 Redis 连接
redis-cli -h localhost -p 6379 ping
```

### 问题 4：健康检查失败

**检查步骤**：
1. 访问 `/readiness` 端点查看详细信息
2. 检查依赖服务状态
3. 查看应用日志
4. 验证配置文件中的依赖配置

### 问题 5：内存占用过高

**检查步骤**：
1. 访问 `/metrics` 端点查看 Prometheus 指标（需启用 `metrics.enabled`）
2. 检查 Goroutine 数量
3. 检查数据库连接池配置
4. 检查缓存配置

## 性能优化建议

### 连接池优化

**数据库连接池**：

```yaml
database:
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
```

**Redis 连接池**：

```yaml
redis:
  pool_size: 10
  min_idle_conns: 5
```

### Goroutine 管理

避免创建过多的 Goroutine，使用工作池模式限制并发数。

### 缓存策略优化

- 使用多级缓存（内存 + Redis）
- 设置合理的 TTL
- 监控缓存命中率

## 安全建议

### 1. 使用 HTTPS

在生产环境中启用 HTTPS：

```yaml
http:
  enabled: true
  tls:
    enabled: true
    cert_file: /path/to/cert.pem
    key_file: /path/to/key.pem
```

### 2. 敏感信息加密

- 不要在配置文件中明文存储密码
- 使用环境变量传递敏感信息
- 考虑使用密钥管理服务（如 HashiCorp Vault）

### 3. 网络隔离

- 数据库和 Redis 不应暴露在公网
- 使用防火墙限制访问
- 使用 VPN 或私有网络

### 4. 定期更新

- 定期更新依赖包
- 关注安全公告
- 及时修复已知漏洞

## 备份建议

### 数据库备份

定期备份数据库：

```bash
# MySQL 备份
mysqldump -u myapp -p myapp > backup_$(date +%Y%m%d).sql

# PostgreSQL 备份
pg_dump -U myapp myapp > backup_$(date +%Y%m%d).sql

# SQLite 备份
cp data/app.db backup/app_$(date +%Y%m%d).db
```

### 配置文件备份

备份配置文件：

```bash
tar -czf config_backup_$(date +%Y%m%d).tar.gz configs/
```

## 总结

- 单实例部署适合小型应用
- 多实例部署需要负载均衡器和共享存储
- Docker 部署简化环境配置
- 配置健康检查和监控确保服务稳定
- 定期备份重要数据
- 遵循安全最佳实践

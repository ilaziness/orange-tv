# 多阶段构建 Dockerfile
# 阶段 1: 构建阶段
FROM golang:1.26.4-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git make ca-certificates

# 设置工作目录
WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用 Docker 缓存）
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
ARG VERSION=1.0.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X github.com/ilaziness/orange-tv/cmd.version=${VERSION}" \
    -o orange-tv .

# 阶段 2: 运行阶段（使用 distroless 提高安全性）
FROM gcr.io/distroless/static-debian12

# 复制 CA 证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 复制二进制文件
COPY --from=builder /build/orange-tv /app/orange-tv

# 复制配置文件
COPY configs/ /app/configs/

# 复制迁移文件
COPY migrations/ /app/migrations/

# 暴露端口
EXPOSE 8080

# 健康检查（探测 /readiness，含数据库等依赖就绪状态）
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/orange-tv", "health"]

# 启动应用
CMD ["/app/orange-tv", "serve", "-c", "/app/configs/config.yaml"]

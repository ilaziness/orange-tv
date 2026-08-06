# Orange TV 一体化镜像 Dockerfile（基于 pack 产物构建）
# 镜像内含：后端二进制 + 前端静态资源 + nginx，单容器即可对外服务。
#
# pack 产物同时包含 linux 与 windows 二进制，Docker 镜像仅使用 linux 二进制。
#
# 构建方式：
#   docker build -t orange-tv:<version> .
#
# 运行：
#   docker run -d -p 80:80 -p 81:81 \
#     -e DATABASE_HOST=... -e DATABASE_USER=... -e DATABASE_PASSWORD=... -e DATABASE_NAME=... \
#     -e JWT_SECRET=... \
#     orange-tv:<version>

FROM nginx:alpine

# 安装 ca-certificates（后端 HTTPS 调用外部资源站需要）与时区
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 后端二进制与运行时资源（仅 linux 二进制，windows 二进制不用于 Docker）
COPY orange-tv        /app/orange-tv
COPY configs          /app/configs
COPY migrations       /app/migrations

# 前端静态资源
COPY web/client       /usr/share/nginx/html/client
COPY web/admin        /usr/share/nginx/html/admin

# nginx 配置
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

# 入口脚本：自动迁移 + 后台启动后端 + 前台运行 nginx
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /app/orange-tv /docker-entrypoint.sh

# 工作目录设为 /app，便于容器内执行命令时定位配置与迁移文件
WORKDIR /app

# 用户端 80，管理端 81
EXPOSE 80 81

ENTRYPOINT ["/docker-entrypoint.sh"]

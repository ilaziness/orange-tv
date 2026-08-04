# Orange TV 发布包

本目录由 `make pack` 生成，包含运行 Orange TV 所需的全部资源。修改配置后即可直接启动。

## 目录结构

```
.
├── orange-tv                  # 后端二进制（默认 linux/amd64）
├── configs/                   # 配置文件
│   ├── config.yaml            # 默认/示例配置（含完整字段说明，开发用）
│   └── config.prod.yaml       # 生产环境配置（裸机与 Docker 共用，可通过环境变量覆盖）
├── migrations/                # 数据库迁移脚本
├── web/
│   ├── client/                # 用户端前端构建产物
│   └── admin/                 # 管理端前端构建产物
├── nginx/
│   └── nginx.conf             # nginx 示例配置（用户端 80 / 管理端 81）
├── Dockerfile                 # 一体化镜像构建文件（后端 + 前端 + nginx）
├── docker-compose.yml         # 一键 Docker Compose（应用 + MySQL）
├── docker-entrypoint.sh       # 容器入口脚本
└── README.md
```

## 方式一：裸机运行

1. 修改配置：编辑 `configs/config.prod.yaml`，确认数据库连接等信息正确（或通过环境变量覆盖）。

2. 执行数据库迁移：
   ```sh
   ./orange-tv migrate up -c configs/config.prod.yaml
   ```

3. 启动后端：
   ```sh
   ./orange-tv serve -c configs/config.prod.yaml
   ```

4. 用 nginx 托管前端（参考 `nginx/nginx.conf`）：
   - 将 `nginx.conf` 拷贝/软链到 nginx 配置目录（如 `/etc/nginx/conf.d/`）
   - 确认静态资源路径与配置一致（默认指向本目录下的 `web/client` 与 `web/admin`，请按实际部署路径调整）
   - reload nginx：`nginx -s reload`

## 方式二：Docker Compose 一键部署（推荐）

使用 `docker-compose.yml` 一键启动应用（后端 + 前端 + nginx）和 MySQL 数据库：

```sh
# 1.（可选）创建 .env 文件自定义密码等变量，或直接使用默认值
#    可配置的变量：
#      DB_USER, DB_PASSWORD, DB_NAME  — 数据库账号（同步注入 MySQL 和 app）
#      MYSQL_ROOT_PASSWORD            — MySQL root 密码
#      JWT_SECRET                     — JWT 签名密钥（生产环境必填，至少 32 字节）
#      HTTP_PORT, ADMIN_PORT, MYSQL_PORT — 端口映射

# 2. 启动所有服务
docker compose up -d

# 3. 首次启动执行数据库迁移
docker compose exec app /app/orange-tv migrate up -c /app/configs/config.prod.yaml

# 4. 查看日志
docker compose logs -f app

# 停止
docker compose down

# 停止并删除数据卷（清空数据库）
docker compose down -v
```

启动后：
- 用户端：`http://<host>:80`
- 管理端：`http://<host>:81`

> Docker 容器内后端监听 `0.0.0.0:8080`，由同容器 nginx 反代访问，无需对外暴露。配置文件为 `configs/config.prod.yaml`，数据库连接通过环境变量注入。

## 方式三：Docker 单独构建（可选）

如不使用 docker-compose，也可单独构建一体化镜像（需自行准备 MySQL）：

```sh
docker build -t orange-tv:1.0.0 .

docker run -d -p 80:80 -p 81:81 \
  -e DATABASE_HOST=<mysql-host> -e DATABASE_USER=<user> \
  -e DATABASE_PASSWORD=<password> -e DATABASE_NAME=orange_tv \
  -e JWT_SECRET=<your-jwt-secret-at-least-32-bytes> \
  orange-tv:1.0.0
```

如需自定义配置，挂载配置目录覆盖：

```sh
docker run -d -p 80:80 -p 81:81 \
  -v $(pwd)/configs:/app/configs \
  -e CONFIG_FILE=/app/configs/config.prod.yaml \
  -e JWT_SECRET=<your-jwt-secret-at-least-32-bytes> \
  orange-tv:1.0.0
```

## 访问地址

| 端口 | 用途 |
| ---- | ---- |
| 80   | 用户端（nginx） |
| 81   | 管理端（nginx） |
| 8080 | 后端 HTTP（容器内 / nginx 反代，无需对外暴露） |

## 备注

- 生产环境（裸机与 Docker）统一使用 `configs/config.prod.yaml`，通过环境变量注入 `DATABASE_HOST` / `DATABASE_PORT` / `DATABASE_USER` / `DATABASE_PASSWORD` / `DATABASE_NAME` / `JWT_SECRET` 等敏感信息，无需修改配置文件。
- 如需 HTTPS，自行在 nginx 中添加 ssl 配置并替换 listen 端口。
- 后端二进制默认为 `linux/amd64`，打包时可通过 `make pack PACK_GOOS=<os> PACK_GOARCH=<arch>` 覆盖。

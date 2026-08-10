# Orange TV 发布包

## 目录结构

```shell
.
├── orange-tv                  # 后端二进制（linux/amd64）
├── orange-tv.exe              # 后端二进制（windows/amd64）
├── configs/                   # 配置文件
│   ├── config.yaml            # 默认/示例配置（含完整字段说明，开发用）
│   └── config.prod.yaml       # 生产环境配置（裸机与 Docker 共用，可通过环境变量覆盖）
├── migrations/                # 数据库迁移脚本
├── web/
│   ├── client/                # 用户端前端构建产物
│   └── admin/                 # 管理端前端构建产物
├── nginx/
│   └── nginx.conf             # nginx 示例配置（用户端 80 / 管理端 81）
├── .env.example               # Docker Compose 环境变量示例
├── Dockerfile                 # 一体化镜像构建文件（后端 + 前端 + nginx）
├── docker-compose.yml         # 一键 Docker Compose（应用 + MySQL）
├── docker-entrypoint.sh       # 容器入口脚本（自动迁移 + 启动后端 + nginx）
└── README.md
```

## 方式一：Docker Compose 一键部署（推荐）

> 前置条件：服务器已安装 Docker 和 Docker Compose（Docker 20.10+）。

使用 `docker-compose.yml` 一键启动应用（后端 + 前端 + nginx）和 MySQL 数据库。

```sh
# 1. 复制环境变量示例文件并按需修改
cp .env.example .env
#    可配置的变量：
#      DB_USER, DB_PASSWORD, DB_NAME  — 数据库账号（同步注入 MySQL 和 app）
#      MYSQL_ROOT_PASSWORD            — MySQL root 密码
#      JWT_SECRET                     — JWT 签名密钥（生产环境必填，至少 32 字节）
#      HTTP_PORT, ADMIN_PORT, MYSQL_PORT — 端口映射
#      AUTO_MIGRATE                   — 是否自动迁移（true/false，默认 true）

# 2. 启动所有服务（首次启动会自动创建数据库并执行迁移）
docker compose up -d

# 3. 等待 app 容器启动完成，查看日志确认无报错
docker compose logs -f app

# 4. 创建管理员账号（见下方「初始化管理员账号」章节），否则无法登录管理端

# 停止
docker compose down

# 停止并删除数据卷（清空数据库，慎用）
docker compose down -v
```

app 容器启动时会自动执行数据库迁移（`AUTO_MIGRATE=true`），无需手动运行 migrate 命令。

启动后（将 `<host>` 替换为服务器 IP，本机访问用 `localhost`）：

- 用户端：`http://<host>:80`（默认，可通过 `HTTP_PORT` 修改）
- 管理端：`http://<host>:81`（默认，可通过 `ADMIN_PORT` 修改）

## 方式二：裸机运行（Linux）

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

## 方式三：裸机运行（Windows）

1. 修改配置：编辑 `configs\config.prod.yaml`，确认数据库连接等信息正确。

2. 执行数据库迁移：

   ```sh
   orange-tv.exe migrate up -c configs\config.prod.yaml
   ```

3. 启动后端：

   ```sh
   orange-tv.exe serve -c configs\config.prod.yaml
   ```

4. 用 IIS 或其他 Web 服务器托管 `web\client` 与 `web\admin` 下的前端静态资源，将 `/api` 反向代理到 `http://127.0.0.1:8080`。

## 方式四：Docker 单独构建（可选）

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

## 初始化管理员账号

服务启动且数据库迁移完成后，需要创建一个超级管理员账号才能登录管理端。`admin create` 命令会创建一个 `super_admin` 角色的启用账号。

**参数：**

- `--username`（必填，3-50 字符）
- `--password`（必填，6-72 字符）
- `--email`（可选）

> 下面的命令中请将 `your_password` 替换为你自己的密码。

### Docker Compose

```sh
docker compose exec app /app/orange-tv admin create --username admin --password your_password --email admin@example.com
```

### 裸机（Linux）

```sh
./orange-tv admin create --username admin --password your_password --email admin@example.com -c configs/config.prod.yaml
```

### 裸机（Windows）

```sh
orange-tv.exe admin create --username admin --password your_password --email admin@example.com -c configs\config.prod.yaml
```

创建成功后输出类似 `Created super_admin: id=1 username=admin`，即可用该账号登录管理端 `http://<host>:81`。

## 访问地址

| 端口 | 用途 |
| ---- | ---- |
| 80   | 用户端（nginx） |
| 81   | 管理端（nginx） |
| 8080 | 后端 HTTP（容器内 / nginx 反代，无需对外暴露） |

## 项目源码

[github.com/ilaziness/orange-tv](https://github.com/ilaziness/orange-tv)

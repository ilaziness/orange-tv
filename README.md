# 小橘TV 影视系统

> English version: [README_EN.md](README_EN.md) · 技术文档：[docs/README.md](docs/README.md)

一个开箱即用的影视站点系统，用来搭建面向观众的在线影视网站。支持影视采集、内容管理、播放和后台运营，前端自适应手机、平板和电脑。

## 这个系统能做什么？

面向普通观众的观影网站。观众打开网站就能：

- 在**首页**看到推荐影视、分类入口和最新内容
- 按**分类**（电影、电视剧、综艺、动漫等）浏览影视
- 按**地区、年份、语言**筛选和排序影视
- 进入**影视详情页**查看海报、简介、导演、演员、评分、播放源和剧集列表
- 在**播放页**流畅观看视频，支持切换播放源、选集、快进快退、全屏
- 通过**搜索**快速找到想看的影视
- 在手机、平板、电脑上都有良好的浏览体验（自适应屏幕）
- 切换明暗主题

## 谁来使用？

系统分为两个独立的部分，分别面向不同的人：

| 部分 | 使用者 | 用来做什么 |
| ------ | ------ | ------ |
| **用户端**（前台网站） | 普通观众 | 浏览、搜索、观看影视 |
| **管理后台** | 站长 / 运营 / 录入员 | 管理影视内容、配置采集、管理用户、设置站点 |

## 管理后台能做的事

- **首页概况**：一眼看到今日新增影视数、总影视数、采集任务状态等运营数据
- **内容管理**
  - 分类管理：多级分类的增删改查、排序、启用/禁用
  - 影视管理：新增、编辑、删除、上下架、批量操作
  - 直播管理：电视直播频道的增删改查、分类和播放地址配置
  - 导演 / 演员 / 标签管理
  - 播放源管理：全局播放源和剧集播放链接
  - Banner 管理：首页轮播图配置
  - 数据采集：配置采集源、定时采集、分类映射、播放源绑定、查看采集日志
- **用户管理**：管理员账号、用户组（角色）、普通用户、登录日志
- **系统设置**：站点名称、Logo、版权、备案号、SEO 关键词、API 配置、系统日志

## 数据采集

如果你不想一部一部影视手动录入，可以使用采集功能自动拉取其他站点的影视数据：

- 支持**默认格式**（系统自定义）和**苹果 CMS 格式**两种采集源
- 支持**定时采集**
- 也支持**手动一键触发**某个采集源立即采集
- 支持**字段映射**和**分类映射**，把外部数据对齐到你的站点结构
- 采集到的播放链接会自动归入你绑定的播放源
- 完整的**采集日志**，方便排查问题

## 快速开始（下载发布包）

不想从源码构建？可以直接下载已打包好的发布包，解压后按以下步骤运行。发布包是自包含的，包含后端二进制、前端构建产物、配置文件、数据库迁移脚本和 Docker 部署文件，修改配置后即可直接启动。

### 1. 下载发布包

前往 [Releases 页面](https://github.com/ilaziness/orange-tv/releases) 下载对应平台的压缩包：

- Linux：`orange-tv-<版本>-linux-amd64.tar.gz`
- Windows：`orange-tv-<版本>-windows-amd64.zip`

### 2. 解压

```sh
# Linux
tar -xzf orange-tv-<版本>-linux-amd64.tar.gz
cd orange-tv-<版本>/

# Windows
# 右键压缩包 → 解压到当前文件夹，进入解压后的目录
```

解压后的目录结构：

```text
.
├── orange-tv                  # 后端二进制（Linux/macOS）
├── orange-tv.exe              # 后端二进制（Windows）
├── configs/                   # 配置文件
│   ├── config.yaml            # 默认/示例配置（含完整字段说明）
│   └── config.prod.yaml       # 生产环境配置（可通过环境变量覆盖）
├── migrations/                # 数据库迁移脚本
├── web/
│   ├── client/                # 用户端前端构建产物
│   └── admin/                 # 管理端前端构建产物
├── nginx/nginx.conf           # nginx 示例配置（用户端 80 / 管理端 81）
├── .env.example               # Docker Compose 环境变量示例
├── Dockerfile                 # 一体化镜像构建文件
├── docker-compose.yml         # 一键 Docker Compose（应用 + MySQL）
└── docker-entrypoint.sh       # 容器入口脚本
```

### 3. 运行

#### 方式一：Docker Compose 一键部署（推荐）

> 前置条件：服务器已安装 Docker 和 Docker Compose（Docker 20.10+）。

```sh
# 复制环境变量示例并按需修改（DB 账号、JWT_SECRET、端口等）
cp .env.example .env

# 启动应用 + MySQL（首次启动会自动创建数据库并执行迁移）
docker compose up -d

# 查看日志确认启动正常
docker compose logs -f app
```

启动后：

- 用户端：`http://<host>:80`
- 管理端：`http://<host>:81`

#### 方式二：裸机运行（Linux）

1. 修改 `configs/config.prod.yaml` 中的数据库连接等信息（或通过环境变量覆盖）。
2. 执行数据库迁移：

   ```sh
   ./orange-tv migrate up -c configs/config.prod.yaml
   ```

3. 启动后端：

   ```sh
   ./orange-tv serve -c configs/config.prod.yaml
   ```

4. 用 nginx 托管前端：参考 `nginx/nginx.conf`，调整静态资源路径后 `nginx -s reload`。

#### 方式三：裸机运行（Windows）

1. 修改 `configs\config.prod.yaml` 中的数据库连接等信息。
2. 执行数据库迁移：

   ```sh
   orange-tv.exe migrate up -c configs\config.prod.yaml
   ```

3. 启动后端：

   ```sh
   orange-tv.exe serve -c configs\config.prod.yaml
   ```

4. 用 IIS 或其他 Web 服务器托管 `web\client` 与 `web\admin` 下的前端静态资源，将 `/api` 反向代理到 `http://127.0.0.1:8080`。

### 4. 初始化管理员账号

服务启动且数据库迁移完成后，需要创建一个超级管理员账号才能登录管理端。将 `your_password` 替换为你自己的密码。

```sh
# Docker Compose
docker compose exec app /app/orange-tv admin create --username admin --password your_password --email admin@example.com

# 裸机（Linux）
./orange-tv admin create --username admin --password your_password --email admin@example.com -c configs/config.prod.yaml

# 裸机（Windows）
orange-tv.exe admin create --username admin --password your_password --email admin@example.com -c configs\config.prod.yaml
```

创建成功后输出类似 `Created super_admin: id=1 username=admin`，即可用该账号登录管理端 `http://<host>:81`，开始录入或采集影视内容。

> 发布包内附带的 `README.md` 包含更详细的部署说明，可一并参考。

## 我想用起来，怎么开始？（从源码构建）

如果你是开发者或运维人员，希望从源码构建，请按以下步骤部署：

1. 准备一台服务器，安装好 Go 1.26.4+ 和 MySQL 8.x/9.x
2. 克隆本仓库：`git clone https://github.com/ilaziness/orange-tv.git`
3. 按照 [部署文档](docs/deployment.md) 完成配置、迁移和启动
4. 用命令行创建第一个管理员账号，然后登录管理后台开始录入或采集影视

详细的安装、配置、命令和技术说明请查阅 [技术文档](docs/README.md)。

## 文档导航

| 文档 | 内容 |
| ------ | ------ |
| [技术文档](docs/README.md) | 项目结构、配置、命令、数据库、可观测性等技术说明 |
| [部署文档](docs/deployment.md) | 单实例、多实例、Docker 部署指南 |
| [模块使用说明](docs/module-usage.md) | 如何选择和删除不需要的服务模块 |
| [可观测性指南](docs/observability.md) | 日志、链路追踪、指标监控配置 |

## 免责声明

本项目仅供学习交流和技术研究使用，不提供、不存储、不传播任何影视内容。

- 系统中的采集功能仅为技术演示，使用者需自行确保采集行为符合相关法律法规及源站点的使用条款
- 使用者需对所采集内容的版权合法性、所部署站点的运营行为自行承担全部责任
- 项目作者不对任何因使用或滥用本系统而产生的法律纠纷负责

## License

本项目基于 [MIT License](LICENSE) 开源。

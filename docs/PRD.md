# 影视系统产品需求文档 (PRD)

## 1. 文档概述

### 1.1 文档目的

本文档定义影视系统的完整产品需求，分用户端和管理后台两大模块进行编写，用于指导产品设计、开发和验收。

### 1.2 目标读者

- 产品经理
- UI/UX设计师
- 前后端开发工程师
- 测试工程师
- 项目经理

### 1.3 术语定义

- **影视站模式**：面向普通用户的观影平台，提供完整的前端界面、播放器、推荐等功能
- **资源站模式**：向其他站点提供数据源的模式，主要通过API接口输出影视数据
- **苹果CMS格式**：国内影视网站广泛使用的JSON数据交换格式
- **主题系统**：可提供给第三方开发的前端界面定制方案

### 1.4 系统模块划分

本系统划分为两个独立产品模块：

- **用户前端**：面向终端用户的影视内容浏览和观看平台
- **管理后台**：面向站点运营管理员的内容管理、采集配置和系统设置平台

## 2. 用户端产品需求

### 2.1 产品目标

为用户提供一个功能完善、界面美观、播放流畅的影视观影平台。平台支持多主题展示（由管理后台配置当前使用主题）、多类型内容展示、多来源播放，并且方便主题开发者进行定制化。

### 2.2 用户画像

主要使用人群：

- 普通观影用户：追剧、观看影片、浏览影视信息
- 潜在第三方主题开发者：希望快速开发和布置自定义主题

### 2.3 功能需求

#### 2.3.1 多主题系统

##### 2.3.1.1 功能描述

用户端需支持多主题展示，当前展示主题由管理后台统一配置。用户端无需提供主题切换入口，按照管理后台配置的主题渲染页面。主题支持通过配置文件定义布局、样式、颜色等属性，并允许第三方开发主题。

主题配置分两层：主题包内 `theme.json` 的 `config` 字段为主题默认配置；管理员可在后台覆盖配置项，覆盖后的配置存储在数据库 `themes.config` 字段中。用户端通过 API 获取的是合并后的最终配置。

##### 2.3.1.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-U-001 | 主题渲染 | 用户端按照管理后台配置的主题渲染页面，用户无主题切换入口 | P1 |
| FR-U-002 | 主题包存储结构 | 主题需以标准化文件夹结构存储，含配置文件、模板、样式、静态资源 | P1 |
| FR-U-003 | 第三方开发支持 | 第三方开发者可以按照规范开发独立主题并导入系统 | P2 |
| FR-U-004 | 主题自定义配置 | 主题可通过变量定义颜色、字体、布局参数等 | P2 |

##### 2.3.1.3 主题包结构

```text
themes/
├── default/                 # 默认主题
│   ├── theme.json          # 主题定义配置
│   ├── styles/             # 主题样式
│   │   ├── variables.css   # CSS变量
│   │   ├── components.css  # 组件样式
│   │   └── layout.css      # 布局样式
│   ├── templates/          # 页面模板
│   │   ├── home.html       # 首页
│   │   ├── category.html   # 分类页
│   │   ├── detail.html     # 详情页
│   │   └── player.html     # 播放页
│   └── assets/             # 静态资源
│       ├── images/
│       ├── fonts/
│       └── js/
└── dark-blue/              # 第三方自定义主题示例
    ├── theme.json
    ├── styles/
    ├── templates/
    └── assets/

```
##### 2.3.1.4 主题配置文件示例

```json

{
  "name": "默认主题",
  "version": "1.0.0",
  "author": "系统",
  "description": "系统默认主题",
  "preview": "/themes/default/preview.jpg",
  "config": {
    "primary_color": "#1890ff",
    "secondary_color": "#52c41a",
    "background_color": "#f0f2f5",
    "text_color": "#262626",
    "header_height": "64px",
    "sidebar_width": "240px",
    "enable_dark_mode": false,
    "custom_fonts": []
  },
  "templates": {
    "home": "home.html",
    "category": "category.html",
    "detail": "detail.html",
    "player": "player.html"
  },
  "custom_css": "",
  "custom_js": ""
}

```
#### 2.3.2 首页分类展示

##### 2.3.2.1 功能描述

首页应展示影视分类，并且以卡片形式展示各分类下的影视内容。首页支持顶部导航、轮播图、分类入口和内容卡片列表。

##### 2.3.2.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-U-005 | 顶部导航 | 包含Logo、分类菜单、搜索框、用户入口 | P1 |
| FR-U-006 | 轮播横幅 | 展示推荐影视的横幅图片和播放按钮 | P1 |
| FR-U-007 | 分类入口 | 展示主要分类（如电影、电视剧、综艺、动漫） | P1 |
| FR-U-008 | 影视卡片列表 | 以卡片形式展示影视封面、标题、年份、评分 | P1 |
| FR-U-009 | 响应式布局 | 首页在PC、平板、移动端都能正常展示 | P1 |

##### 2.3.2.3 首页布局草图

```text
┌─────────────────────────────────────────────────────────────┐
│                        顶部导航栏                            │
│  Logo | 分类菜单 | 搜索框 | 用户登录                       │
├─────────────────────────────────────────────────────────────┤
│                        轮播横幅                            │
│              [推荐影视轮播图 + 播放按钮]                     │
├─────────────────────────────────────────────────────────────┤
│                      分类展示区                             │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │  电影   │ │  电视剧  │ │  综艺   │ │  动漫   │           │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │
├─────────────────────────────────────────────────────────────┤
│                      内容展示区                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   影视卡片   │ │   影视卡片   │ │   影视卡片   │           │
│  │  [封面图]   │ │  [封面图]   │ │  [封面图]   │           │
│  │   标题      │ │   标题      │ │   标题      │           │
│  │   年份/评分 │ │   年份/评分 │ │   年份/评分 │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘

```
#### 2.3.3 分类页面展示

##### 2.3.3.1 功能描述

分类页面展示某一分类下的影视列表，支持按地区、年份、语言等条件筛选和排序。

##### 2.3.3.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-U-010 | 面包屑导航 | 显示当前所在分类路径 | P2 |
| FR-U-011 | 筛选条件 | 支持按地区、年份、语言、排序筛选 | P1 |
| FR-U-012 | 网格列表 | 展示影视卡片网格，支持分页加载 | P1 |
| FR-U-013 | 空状态提示 | 筛选无结果时给出提示 | P2 |

##### 2.3.3.3 分类页布局草图

```text
┌─────────────────────────────────────────────────────────────┐
│                        顶部导航栏                            │
├─────────────────────────────────────────────────────────────┤
│                      面包屑导航                             │
│  首页 > 电影 > 动作片                                       │
├─────────────────────────────────────────────────────────────┤
│                      筛选工具栏                             │
│  [地区:全部▼] [年份:全部▼] [语言:全部▼] [排序:最新▼]         │
├─────────────────────────────────────────────────────────────┤
│                      内容网格                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   影视卡片   │ │   影视卡片   │ │   影视卡片   │           │
│  │  [封面图]   │ │  [封面图]   │ │  [封面图]   │           │
│  │   标题      │ │   标题      │ │   标题      │           │
│  │   年份/评分 │ │   年份/评分 │ │   年份/评分 │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│                        分页组件                             │
└─────────────────────────────────────────────────────────────┘

```
#### 2.3.4 详情页展示

##### 2.3.4.1 功能描述

详情页展示影视的详细信息，包括影视标题、副标题、年份、地区、语言、时长、上映日期、导演、演员、标签、简介、评分、播放源和剧集列表。

##### 2.3.4.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-U-014 | 影视基本信息 | 展示海报图、标题、副标题、年份、地区、语言、时长、上映日期、评分 | P1 |
| FR-U-015 | 创作人信息 | 展示导演、演员、标签 | P1 |
| FR-U-016 | 剧情简介 | 展示影视简介文本 | P1 |
| FR-U-017 | 播放源列表 | 列出所有可用播放源 | P1 |
| FR-U-018 | 剧集列表 | 连载类影视展示剧集选择 | P1 |
| FR-U-019 | 相关推荐 | 展示相关影视推荐 | P2 |

##### 2.3.4.3 详情页布局草图

```text
┌─────────────────────────────────────────────────────────────┐
│                        顶部导航栏                            │
├─────────────────────────────────────────────────────────────┤
│                      影视信息区                             │
│  ┌─────────────┐ ┌───────────────────────────────────────┐ │
│  │             │ │  标题：电影名称                         │ │
│  │   海报图    │ │  副标题：电影副标题                     │ │
│  │             │ │  年份：2024 | 地区：美国 | 语言：英语   │ │
│  │             │ │  时长：120分钟 | 上映日期：2024-01-01     │ │
│  │             │ │  评分：8.5                              │ │
│  └─────────────┘ │  导演：导演名称                         │ │
│                  │  演员：演员1, 演员2, 演员3              │ │
│                  │  标签：动作, 科幻, 冒险                 │ │
│                  │  简介：电影简介内容...                   │ │
│                  └───────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      播放源选择                             │
│  [播放源1] [播放源2] [播放源3] [下载源]                     │
├─────────────────────────────────────────────────────────────┤
│                      剧集列表（连载类）                     │
│  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐         │
│  │ 第1集│ │ 第2集│ │ 第3集│ │ 第4集│ │ 第5集│ │ 第6集│         │
│  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘         │
├─────────────────────────────────────────────────────────────┤
│                      相关推荐                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   推荐影片   │ │   推荐影片   │ │   推荐影片   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘

```
#### 2.3.5 播放页展示

##### 2.3.5.1 功能描述

播放页应提供流畅的视频播放体验，前端播放器采用Video.js，支持HLS、MP4、DASH、FLV等常见播放链接格式，支持多个播放源切换、多集展示、快进快退、全屏播放等常用播放功能。

##### 2.3.5.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-U-020 | 视频播放器 | 前端播放器采用Video.js，支持HLS、MP4、DASH、FLV等常见播放链接格式 | P1 |
| FR-U-021 | 播放源切换 | 用户可切换不同播放源 | P1 |
| FR-U-022 | 选集/换集 | 连载类影视可快速切换剧集 | P1 |
| FR-U-023 | 播放控制 | 播放/暂停、快进快退、音量、清晰度、全屏 | P1 |
| FR-U-024 | 播放记忆 | 记录上次播放进度，重新打开时继续播放 | P2 |

##### 2.3.5.3 播放页布局草图

```text
┌─────────────────────────────────────────────────────────────┐
│                        顶部导航栏                            │
├─────────────────────────────────────────────────────────────┤
│                      影视标题                               │
│  电影名称 - 第1集                                           │
├─────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────────┐ │
│  │                   视频播放器                          │ │
│  └───────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      播放控制                               │
│  [上一集] [播放/暂停] [下一集] [选集] [清晰度] [全屏]       │
├─────────────────────────────────────────────────────────────┤
│                      播放源切换                             │
│  当前播放源：播放源1 [切换播放源▼]                         │
└─────────────────────────────────────────────────────────────┘

```

#### 2.3.6 加载过渡体验

##### 2.3.6.1 功能描述

页面切换和数据加载时需展示加载过渡动画，防止网络延迟导致页面静止，提升用户体验。

##### 2.3.6.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-U-025 | 页面加载过渡 | 路由切换时展示页面级加载动画，数据加载完成后渐入内容 | P1 |
| FR-U-026 | 骨架屏 | 列表页、详情页等数据驱动页面在加载期间展示骨架屏占位 | P1 |
| FR-U-027 | 按钮加载态 | 表单提交、采集触发等异步操作按钮展示加载状态，防止重复提交 | P1 |
| FR-U-028 | 图片懒加载 | 影视封面、海报等图片采用懒加载，加载期间展示占位图 | P2 |

### 2.4 用户端API接口

#### 2.4.0 响应约定

统一响应结构（与 `internal/response` 一致）：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- **成功**：业务码 `code = 0`，HTTP 状态通常为 200。
- **失败**：`code` 为业务错误码（见 `internal/errcode`），`message` 为错误说明；HTTP 状态与错误类型对应（如 400/401/403/404/409/422/500）。
- **分页**：列表接口 `data` 直接包含：

```json
{
  "list": [],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

分页查询参数兼容 `page` + `page_size`（或 `limit`），默认 `page=1`、`page_size=20`，`page_size` 最大 100。

#### 2.4.1 基础接口

```go
// 路径前缀约定（与代码 internal/router/paths.go 一致）：
// 仅定义版本前缀：用户端 /api/client/v1，管理端 /api/admin/v1，内网 /api/internal/v1
// 业务路径在路由注册时直接写字符串，不在 paths.go 再定义常量

// 获取分类列表（仅启用、未软删除；树形，叶子 children 为空数组）
GET /api/client/v1/categories
Response: {
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "电影",
      "parent_id": 0,
      "sort_order": 0,
      "status": 1,
      "children": [
        {"id": 11, "name": "动作片", "parent_id": 1, "sort_order": 0, "status": 1, "children": []},
        {"id": 12, "name": "喜剧片", "parent_id": 1, "sort_order": 0, "status": 1, "children": []}
      ]
    }
  ]
}

// 获取影视列表（仅上架、未软删除）
GET /api/client/v1/videos?category_id=1&page=1&page_size=20&year=2024&region=美国&language=英语&sort=created_at_desc
Response: {
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}

// 获取影视详情（仅上架；聚合导演/演员/标签及启用播放源与剧集）
GET /api/client/v1/videos/{id}
Response: {
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "title": "电影标题",
    "subtitle": "副标题",
    "description": "简介",
    "category_id": 1,
    "cover": "封面地址",
    "poster": "海报地址",
    "year": 2024,
    "region": "美国",
    "language": "英语",
    "duration": 120,
    "release_date": "2024-01-01",
    "rating": 8.5,
    "serial_status": 2,
    "view_count": 0,
    "directors": [{"id": 1, "name": "导演名"}],
    "actors": [{"id": 1, "name": "演员1", "role": "主角"}],
    "tags": [{"id": 1, "name": "动作"}],
    "sources": [
      {
        "id": 1,
        "name": "播放源1",
        "episodes": [
          {"episode": 1, "title": "第1集", "url": "播放地址", "quality": "1080p", "format": "hls"}
        ]
      }
    ]
  }
}

// 搜索影视（关键词匹配标题/副标题/简介；仅上架）
GET /api/client/v1/search?keyword=关键词&page=1&page_size=20
Response: {
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 50,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
  }
}

// 获取直播频道列表（第三阶段实现）
GET /api/client/v1/live?category=新闻&page=1&page_size=20
Response: {
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "频道名称",
        "category": "新闻",
        "stream_url": "直播流地址",
        "logo": "Logo地址",
        "description": "频道简介"
      }
    ],
    "total": 30,
    "page": 1,
    "page_size": 20,
    "total_pages": 2
  }
}
```

#### 2.4.2 站点配置接口

```go
// 获取当前主题配置（由管理后台配置，用户端不可切换主题；第三阶段完整实现）
GET /api/client/v1/theme/current
Response: {
  "code": 0,
  "message": "success",
  "data": {
    "name": "默认主题",
    "identifier": "default",
    "config": {
      "primary_color": "#1890ff",
      "secondary_color": "#52c41a",
      "background_color": "#f0f2f5",
      "text_color": "#262626",
      "header_height": "64px",
      "sidebar_width": "240px",
      "enable_dark_mode": false,
      "custom_fonts": []
    },
    "templates": {
      "home": "home.html",
      "category": "category.html",
      "detail": "detail.html",
      "player": "player.html"
    },
    "custom_css": "",
    "custom_js": ""
  }
}
```
### 2.5 用户端非功能需求

#### 2.5.1 性能需求

- 首页首屏加载时间不超过3秒
- 视频列表页加载时间不超过2秒
- 图片应使用懒加载
- 应支持缓存和离线缓存

#### 2.5.2 兼容性需求

- 支持Chrome、Firefox、Safari、Edge等主流浏览器
- 支持PC、平板、手机等多端屏幕
- 支持移动端横竖屏切换

#### 2.5.3 安全需求

- 访问控制通过API安全验证
- 防止XSS和CSRF攻击
- 图片和静态资源验证

## 3. 管理后台产品需求

### 3.1 产品目标

为站点运营管理员提供一套完善的影视内容管理工具，支持影视采集、内容管理、主题配置、系统设置等功能，并且可将站点作为资源站供其他站点采集。

### 3.2 用户画像

主要使用人群：

- 站点运营管理员：管理影视内容、配置采集源和站点参数
- 内容录入人员：人工录入影视数据和播放源
- 资源站管理员：配置API接口和数据输出格式

### 3.3 功能需求

#### 3.3.1 影视采集功能

##### 3.3.1.1 功能描述

管理后台应支持影视数据采集，采集源支持两种格式：默认格式（系统自定义格式）和苹果CMS格式。支持定时采集（cron表达式配置）、采集日志查看、采集分类映射、播放源绑定等功能。

##### 3.3.1.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-001 | 采集格式支持 | 支持默认格式（系统格式）和苹果CMS格式两种采集源 | P1 |
| FR-A-002 | 采集源配置 | 支持配置采集地址、密钥、采集规则 | P1 |
| FR-A-003 | 定时采集 | 支持通过cron表达式配置定时采集，空值表示不开启 | P2 |
| FR-A-004 | 手动采集 | 支持管理员手动触发某个采集源的采集 | P1 |
| FR-A-005 | 数据映射 | 支持配置采集源字段与本系统字段的映射 | P1 |
| FR-A-006 | 采集日志 | 记录采集执行日志，包含成功、失败等状态 | P2 |
| FR-A-007 | 分类映射 | 支持采集源外部分类映射到系统内分类 | P1 |
| FR-A-008 | 播放源绑定 | 采集源绑定播放源，采集到的播放链接存入对应播放源 | P1 |

##### 3.3.1.3 采集源配置页面布局草图

```text
┌─────────────────────────────────────────────────────────────┐
│                      采集源配置                             │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   添加源    │ │   编辑源    │ │   删除源    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
├─────────────────────────────────────────────────────────────┤
│                      采集源列表                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ 源名称 | 类型 | 状态 | 最后采集 | 操作                │ │
│  │ ───────────────────────────────────────────────────── │ │
│  │ 苹果CMS源 | 苹果CMS | 启用 | 2024-01-01 12:00 | [编辑] │ │
│  │ 自定义源 | 默认格式 | 禁用 | 2024-01-01 10:00 | [编辑] │ │
│  └───────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      采集配置表单                           │
│  源名称: [________________]                                │
│  源类型: [默认格式 ▼]                                    │
│  采集地址: [________________]                            │
│  API密钥: [________________]                              │
│  定时采集: [*/60 * * * *]  (cron表达式，空表示不开启)      │
│  绑定播放源: [播放源1 ▼]                                  │
│  高级配置: [展开配置▼]                                    │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ 分类映射: [外部分类→系统分类映射表]                    │ │
│  │ 采集字段: [标题,封面,简介,播放地址]                    │ │
│  │ 数据映射: [配置JSON字段映射]                          │ │
│  │ 过滤规则: [设置采集过滤条件]                          │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

```

##### 3.3.1.4 苹果CMS数据格式示例

```javascript

// 苹果CMS API响应格式示例
const appleCmsFormat = {
  "code": 1,
  "msg": "数据列表",
  "page": 1,
  "pagecount": 100,
  "limit": "20",
  "total": 2000,
  "list": [
    {
      "vod_id": "123",
      "type_id": "1",
      "type_id_1": "0",
      "group_id": "0",
      "vod_name": "电影名称",
      "vod_sub": "副标题",
      "vod_en": "英文名",
      "vod_status": "1",
      "vod_letter": "A",
      "vod_color": "",
      "vod_tag": "标签",
      "class": "分类",
      "pic": "封面图",
      "actor": "演员",
      "director": "导演",
      "writer": "编剧",
      "behind": "幕后",
      "blurb": "简介",
      "remarks": "备注",
      "pubdate": "2024-01-01",
      "vod_year": "2024",
      "vod_area": "美国",
      "vod_lang": "英语",
      "vod_duration": "120",
      "total": "10",
      "serial": "1",
      "note": "备注",
      "douban_id": "123456",
      "douban_score": "8.5",
      "imdb_id": "tt1234567",
      "imdb_score": "8.0",
      "jumpurl": "",
      "tpl": "",
      "vod_play_from": "播放源名称",
      "vod_play_server": "",
      "vod_play_note": "",
      "vod_play_url": "第1集$播放地址1#第2集$播放地址2"
    }
  ]
};

// 数据映射配置
const dataMapping = {
  "title": "vod_name",
  "subtitle": "vod_sub",
  "description": "blurb",
  "cover": "pic",
  "category": "type_id",
  "director": "director",
  "actors": "actor",
  "year": "vod_year",
  "region": "vod_area",
  "language": "vod_lang",
  "release_date": "pubdate",
  "duration": "vod_duration",
  "rating": "douban_score",
  "play_sources": "vod_play_from",
  "play_urls": "vod_play_url"
};

```
#### 3.3.2 站点运行模式

##### 3.3.2.1 功能描述

管理后台可设置站点运行模式。默认为影视站，可选切换为资源站模式。资源站下可设置数据输出格式，默认为自定义JSON格式，可开启苹果CMS兼容格式。

##### 3.3.2.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-009 | 影视站模式 | 默认运行模式，面向普通用户观影 | P1 |
| FR-A-010 | 资源站模式 | 向第三方站点提供影视数据源 | P1 |
| FR-A-011 | API输出格式 | 资源站默认自定义JSON格式，可选苹果CMS兼容 | P1 |
| FR-A-012 | 第三方采集开关 | 可配置是否允许第三方采集 | P2 |
| FR-A-013 | 访问密钥 | 资源站API可配置访问密钥 | P2 |

##### 3.3.2.3 站点模式配置页面草图

```text
┌─────────────────────────────────────────────────────────────┐
│                      站点运行模式配置                       │
├─────────────────────────────────────────────────────────────┤
│  站点模式: [影视站 ▼]                                         │
│                                                                  │
│  资源站设置:                                                    │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 输出格式: [自定义JSON ▼]                          │  │
│  │ 开启苹果CMS兼容: [开关]                               │  │
│  │ 允许第三方采集: [开关]                                │  │
│  │ API密钥: [________________]                          │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

```
#### 3.3.3 内容管理菜单

##### 3.3.3.1 功能描述

管理后台需提供完整的内容管理功能，包括分类管理、影视管理、直播管理（电视直播）和数据采集。

##### 3.3.3.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-014 | 分类管理 | 支持多级分类的增删改查、排序与启用/禁用 | P1 |
| FR-A-015 | 影视列表 | 展示影视数据列表，支持搜索、筛选、分页 | P1 |
| FR-A-016 | 影视编辑 | 支持新增、编辑、删除影视基本信息 | P1 |
| FR-A-017 | 播放源管理 | 支持全局播放源管理和剧集播放链接管理 | P1 |
| FR-A-018 | 直播管理 | 支持电视直播频道的增删改查、分类和播放地址配置 | P1 |
| FR-A-019 | 数据采集 | 支持采集源配置（含定时采集、分类映射、播放源绑定）、采集任务、采集日志 | P1 |
| FR-A-020 | 批量操作 | 支持影视的批量上下架、删除 | P2 |

#### 3.3.4 首页概况

##### 3.3.4.1 功能描述

管理后台首页无二级菜单，直接在下部主体展示站点概况信息，包括今日新增影视数、总影视数、用户访问量、采集任务状态和系统健康状态等。

##### 3.3.4.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-021 | 数据概览 | 展示核心运营数据卡片，如今日新增影视、总影视数、在线人数 | P1 |
| FR-A-022 | 快捷入口 | 提供常用功能快捷入口，如新增影视、分类管理、采集任务 | P1 |
| FR-A-023 | 系统公告 | 展示系统公告、采集任务状态和待处理事项 | P2 |

#### 3.3.5 用户管理功能

##### 3.3.5.1 功能描述

管理后台需提供用户管理功能，包括管理员账号、用户组和普通用户的管理，支持权限分配和状态控制。

##### 3.3.5.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-024 | 管理员管理 | 支持管理员账号的增删改查、角色分配和启用/禁用 | P1 |
| FR-A-025 | 用户组管理 | 支持用户组（角色）的增删改查和权限配置 | P1 |
| FR-A-026 | 普通用户管理 | 支持普通用户的列表、搜索、禁用、重置密码等操作 | P1 |
| FR-A-027 | 登录日志 | 记录管理员和普通用户的登录日志 | P2 |

#### 3.3.6 主题管理功能

##### 3.3.6.1 功能描述

管理后台需提供主题管理功能，管理员可配置和切换用户端展示的主题。主题切换权限仅保留在管理后台，用户端按照管理后台配置的当前主题自动渲染页面。

##### 3.3.6.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-028 | 主题选择与切换 | 管理员可在已上线主题中选择当前用户端展示的主题，切换后用户端实时生效 | P1 |
| FR-A-029 | 主题上传 | 支持上传第三方主题包并导入系统 | P2 |
| FR-A-030 | 主题预览 | 管理员可预览主题在用户端的展示效果 | P2 |
| FR-A-031 | 主题参数配置 | 管理员可配置主题自定义参数（颜色、字体、布局等） | P2 |

#### 3.3.7 系统设置功能

##### 3.3.7.1 功能描述

管理后台需提供系统设置功能，支持站点名称、站点Logo、站点版权、API配置、系统日志等参数配置。主题选择已归入3.3.6主题管理功能。

##### 3.3.7.2 需求详情

| 序号 | 需求点 | 描述 | 优先级 |
| ------ | -------- | ------ | -------- |
| FR-A-032 | 站点设置 | 支持站点名称、Logo、版权、备案号、SEO关键词、站点描述等基础信息配置 | P1 |
| FR-A-033 | API配置 | 支持资源站模式API开关、第三方采集API开关和API密钥配置 | P2 |
| FR-A-034 | 系统日志 | 支持查看系统运行日志、管理员操作日志和异常日志 | P2 |
| FR-A-035 | 邮件/通知 | 支持系统通知渠道配置，如邮件、站内信等 | P2 |

#### 3.3.8 管理后台布局

##### 3.3.8.1 布局需求

管理后台布局分为上下两部分：

- 顶部：展示一级菜单，包括首页、内容管理、用户管理、系统设置；右侧展示当前登录管理员头像，点击后弹出下拉菜单（账号信息、退出等）。
- 下部：页面主体。点击有二级菜单的一级菜单后，左侧展示二级菜单；右侧页面主体展示对应功能内容。无二级菜单的菜单（如首页）直接在下部主体展示概况内容。

##### 3.3.8.2 一级菜单分类

```javascript

const mainMenus = [
  {
    id: 'home',
    name: '首页',
    icon: 'home',
    children: []
  },
  {
    id: 'content',
    name: '内容管理',
    icon: 'video',
    children: [
      { id: 'category', name: '分类管理', path: '/content/category' },
      { id: 'video', name: '影视管理', path: '/content/video' },
      { id: 'live', name: '直播管理', path: '/content/live' },
      { id: 'collect', name: '数据采集', path: '/content/collect' }
    ]
  },
  {
    id: 'user',
    name: '用户管理',
    icon: 'user',
    children: [
      { id: 'admin', name: '管理员', path: '/user/admin' },
      { id: 'group', name: '用户组管理', path: '/user/group' },
      { id: 'regular', name: '普通用户管理', path: '/user/regular' }
    ]
  },
  {
    id: 'system',
    name: '系统设置',
    icon: 'setting',
    children: [
      { id: 'site-config', name: '站点设置', path: '/system/site' },
      { id: 'theme-config', name: '主题管理', path: '/system/theme' },
      { id: 'api-config', name: 'API配置', path: '/system/api' },
      { id: 'system-log', name: '系统日志', path: '/system/log' }
    ]
  }
];

```
##### 3.3.8.3 管理后台布局草图

```text
┌───────────────────────────────────────────────────────────────────┐
│                          顶部导航栏                                │
│  Logo | 首页 | 内容管理 | 用户管理 | 系统设置 |         [头像] │
├───────────────────────────────────────────────────────────────────┤
│                │                                                  │
│   二级菜单      │               页面主体内容区                      │
│                │                                                  │
│  ┌──────────┐  │  ┌────────────────────────────────────────────┐  │
│  │ 分类管理  │  │  │                                            │  │
│  └──────────┘  │  │            具体功能页面内容                 │  │
│  ┌──────────┐  │  │                                            │  │
│  │ 影视管理  │  │  │                                            │  │
│  └──────────┘  │  │                                            │  │
│  ┌──────────┐  │  │                                            │  │
│  │ 直播管理  │  │  │                                            │  │
│  └──────────┘  │  │                                            │  │
│  ┌──────────┐  │  │                                            │  │
│  │ 数据采集  │  │  │                                            │  │
│  └──────────┘  │  └────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘

```
### 3.4 管理端API接口

#### 3.4.1 认证接口

```go
// 管理员登录与认证（第二阶段）
// 登录成功返回 access_token；受保护接口使用 Authorization: Bearer <token>
// 第二阶段仅允许 super_admin；每个受保护请求实时查询管理员/用户组，不信任 JWT 内角色
// 初始管理员通过 CLI 创建，不提供默认密码：
//   orange-tv admin create --username <name> --password <pass>
POST   /api/admin/v1/auth/login       // 管理员登录
POST   /api/admin/v1/auth/logout      // 管理员退出（无状态，客户端清除 token）
GET    /api/admin/v1/auth/profile     // 获取登录管理员信息
```

#### 3.4.2 内容管理接口

```go
// 分类管理
POST   /api/admin/v1/categories      // 创建分类
PUT    /api/admin/v1/categories/{id} // 更新分类
DELETE /api/admin/v1/categories/{id} // 删除分类（有有效子分类或影视时拒绝）
GET    /api/admin/v1/categories      // 获取分类列表（树形）

// 影视管理
POST   /api/admin/v1/videos          // 创建影视（可同时维护导演/演员/标签关联）
PUT    /api/admin/v1/videos/{id}     // 更新影视
DELETE /api/admin/v1/videos/{id}     // 删除影视（软删除，并清理关联）
GET    /api/admin/v1/videos          // 获取影视列表（搜索/筛选/分页）
GET    /api/admin/v1/videos/{id}     // 获取影视详情（编辑回填用）

// 播放源管理（全局播放源）
POST   /api/admin/v1/play-sources        // 添加播放源
PUT    /api/admin/v1/play-sources/{id}   // 更新播放源
DELETE /api/admin/v1/play-sources/{id}   // 删除播放源（仍有有效剧集时拒绝）
GET    /api/admin/v1/play-sources        // 获取播放源列表

// 播放集数管理
POST   /api/admin/v1/play-episodes        // 添加播放集数
PUT    /api/admin/v1/play-episodes/{id}   // 更新播放集数
DELETE /api/admin/v1/play-episodes/{id}   // 删除播放集数
GET    /api/admin/v1/play-episodes        // 获取播放集数列表（必填 video_id + source_id 筛选）

// 导演管理（名称全局唯一，含软删除占用；仍被未删除影视引用时拒绝删除）
GET    /api/admin/v1/directors           // 获取导演列表（支持搜索）
POST   /api/admin/v1/directors           // 添加导演
PUT    /api/admin/v1/directors/{id}      // 更新导演
DELETE /api/admin/v1/directors/{id}      // 删除导演

// 演员管理（规则同导演）
GET    /api/admin/v1/actors             // 获取演员列表（支持搜索）
POST   /api/admin/v1/actors             // 添加演员
PUT    /api/admin/v1/actors/{id}        // 更新演员
DELETE /api/admin/v1/actors/{id}        // 删除演员

// 标签管理（规则同导演）
GET    /api/admin/v1/tags               // 获取标签列表（支持搜索）
POST   /api/admin/v1/tags               // 添加标签
PUT    /api/admin/v1/tags/{id}          // 更新标签
DELETE /api/admin/v1/tags/{id}          // 删除标签

// 直播管理
POST   /api/admin/v1/live            // 创建直播频道
PUT    /api/admin/v1/live/{id}       // 更新直播频道
DELETE /api/admin/v1/live/{id}       // 删除直播频道
GET    /api/admin/v1/live            // 获取直播频道列表

// 数据采集
POST   /api/admin/v1/collect-sources              // 添加采集源
PUT    /api/admin/v1/collect-sources/{id}         // 更新采集源
DELETE /api/admin/v1/collect-sources/{id}         // 删除采集源
GET    /api/admin/v1/collect-sources              // 获取采集源列表
POST   /api/admin/v1/collect-sources/{id}/categories  // 配置采集源分类映射
GET    /api/admin/v1/collect-sources/{id}/categories  // 获取采集源分类映射
POST   /api/admin/v1/collect/{source_id}/start        // 开始采集
POST   /api/admin/v1/collect/{source_id}/stop         // 停止采集
GET    /api/admin/v1/collect/logs                     // 获取采集日志

```
#### 3.4.3 用户管理接口

```go

// 管理员管理
POST   /api/admin/v1/admins          // 创建管理员
PUT    /api/admin/v1/admins/{id}     // 更新管理员
DELETE /api/admin/v1/admins/{id}     // 删除管理员
GET    /api/admin/v1/admins          // 获取管理员列表

// 用户组/角色管理
POST   /api/admin/v1/groups          // 创建用户组
PUT    /api/admin/v1/groups/{id}     // 更新用户组
DELETE /api/admin/v1/groups/{id}     // 删除用户组
GET    /api/admin/v1/groups          // 获取用户组列表

// 普通用户管理
GET    /api/admin/v1/users           // 获取用户列表
PUT    /api/admin/v1/users/{id}      // 更新用户状态
DELETE /api/admin/v1/users/{id}      // 删除用户
GET    /api/admin/v1/users/{id}/login-logs // 获取用户登录日志

```

#### 3.4.4 系统配置接口

```go

// 系统设置
GET    /api/admin/v1/settings          // 获取系统设置
PUT    /api/admin/v1/settings          // 更新系统设置

// 主题管理
GET    /api/admin/v1/themes            // 获取主题列表
POST   /api/admin/v1/themes            // 上传主题
PUT    /api/admin/v1/themes/{id}       // 更新主题
DELETE /api/admin/v1/themes/{id}       // 删除主题
POST   /api/admin/v1/themes/{id}/activate // 激活主题

```
### 3.5 管理端非功能需求

#### 3.5.1 权限需求

- 管理后台应支持登录认证（JWT Bearer）
- 支持角色和权限控制
  - **第二阶段**：仅预置角色 `super_admin`（用户组名），拥有全部后台权限；不提供用户组 CRUD
  - 授权策略：JWT 仅携带管理员 ID；每个受保护请求实时查库校验启用状态与角色
  - 后续阶段再扩展可配置用户组/细粒度权限
- 应记录管理员操作日志（系统日志完整能力在第四阶段完善）

#### 3.5.2 性能需求

- 采集任务异步执行，不阻塞前端操作
- 影视列表应支持大数据量分页
- 系统设置变更应实时生效

#### 3.5.3 安全需求

- 管理后台接口应进行身份认证
- 应防止SQL注入和跨站脚本攻击
- 密码应加密存储

## 4. 数据库设计

### 4.0 设计约定

- **目标库**：MySQL 8.x/9.x（开发默认 MySQL；迁移脚本使用 MySQL DDL）。
- **模型生成**：迁移执行后使用 CLI `orange-tv gen model` 从数据库表结构生成 Bun 模型，避免手写全量 model。
- **软删除**：
  - 业务主实体表增加字段 `deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'`。
  - **不为 `deleted_at` 单独创建索引**（不添加 `idx_deleted_at` 等）。
  - 列表/详情查询默认过滤 `deleted_at IS NULL`（业务层约定）。
  - 关联表、日志表、系统配置键值表不做软删除。
- **适用软删除的表**：`categories`、`videos`、`directors`、`actors`、`tags`、`play_sources`、`play_episodes`、`live_channels`、`collect_sources`、`themes`、`admins`、`users`、`user_groups`。
- **不做软删除的表**：`video_directors`、`video_actors`、`video_tags`、`collect_source_categories`、`collect_logs`、`login_logs`、`system_logs`、`system_settings`。
- **名称唯一策略**：
  - `directors.name` / `actors.name` / `tags.name`：数据库全局唯一索引（含软删除记录占用名称）。
  - `categories.name` / `play_sources.name`：**不建**数据库唯一索引；业务层仅对未软删除记录做同名冲突校验（软删除后允许复用名称）。
  - `play_episodes (source_id, video_id, episode_number)`：数据库唯一索引；创建/更新时若键被软删除记录占用，允许恢复或回收该键。
- **删除冲突策略**（第二阶段业务规则）：
  - 分类：存在未删除子分类或未删除影视时拒绝删除。
  - 导演/演员/标签：仍被未删除影视引用时拒绝删除。
  - 播放源：仍有未删除剧集时拒绝删除。
  - 删除影视时：软删除影视，并清理导演/演员/标签关联，避免元数据被“幽灵引用”占用。
- **管理员初始化**：不向迁移写入默认账号密码；使用 CLI `orange-tv admin create` 创建并绑定 `super_admin` 用户组（迁移预置该用户组）。

### 4.1 实体关系

```text
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│  Categories │       │    Videos   │       │  Directors  │
│  (分类表)   │       │  (影视表)   │       │  (导演表)   │
└─────────────┘       └─────────────┘       └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│VideoDirectors│      │    Actors   │       │ VideoActors │
│(导演关联表) │       │  (演员表)   │       │(演员关联表) │
└─────────────┘       └─────────────┘       └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    Tags     │       │  VideoTags  │       │ PlaySources │
│  (标签表)   │       │(标签关联表) │       │ (播放源表)  │
└─────────────┘       └─────────────┘       └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│PlayEpisodes │       │CollectSources│      │CollectSource│
│ (播放集数表)│       │ (采集源配置) │       │ Categories  │
└─────────────┘       └─────────────┘       │(采集分类映射)│
                                            └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│ CollectLogs │       │LiveChannels │       │    Themes   │
│ (采集日志)  │       │ (直播频道)  │       │   (主题表)  │
└─────────────┘       └─────────────┘       └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│SystemSettings│      │    Admins   │       │  UserGroups │
│ (系统设置表) │       │  (管理员表) │       │ (用户组表)  │
└─────────────┘       └─────────────┘       └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    Users    │       │  LoginLogs  │       │ SystemLogs  │
│ (普通用户表)│       │ (登录日志表)│       │ (系统日志表)│
└─────────────┘       └─────────────┘       └─────────────┘

```
### 4.2 影视内容表

```sql

-- 影视分类表
CREATE TABLE categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '分类名称',
    parent_id BIGINT NOT NULL DEFAULT 0 COMMENT '父分类ID',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_parent (parent_id)
);

-- 影视作品表
CREATE TABLE videos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL COMMENT '标题',
    subtitle VARCHAR(255) NOT NULL DEFAULT '' COMMENT '副标题',
    description TEXT COMMENT '描述',
    category_id BIGINT NOT NULL DEFAULT 0 COMMENT '分类ID',
    publish_status TINYINT NOT NULL DEFAULT 0 COMMENT '上下架状态：1上架 0下架',
    serial_status TINYINT NOT NULL DEFAULT 1 COMMENT '连载状态：1连载中 2已完结 3即将上线',
    cover_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '封面图',
    poster_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '海报图',
    year INT NOT NULL DEFAULT 0 COMMENT '年份',
    region VARCHAR(50) NOT NULL DEFAULT '' COMMENT '地区',
    rating DECIMAL(3,1) NOT NULL DEFAULT 0.0 COMMENT '评分',
    view_count INT NOT NULL DEFAULT 0 COMMENT '播放次数',
    duration INT NOT NULL DEFAULT 0 COMMENT '时长（分钟）',
    language VARCHAR(50) NOT NULL DEFAULT '' COMMENT '语言',
    release_date DATE NULL COMMENT '上映日期',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_category (category_id),
    INDEX idx_year (year),
    FULLTEXT idx_search (title, subtitle, description)
);

-- 导演表
CREATE TABLE directors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '导演名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

-- 影视导演关联表
CREATE TABLE video_directors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_id BIGINT NOT NULL COMMENT '影视ID',
    director_id BIGINT NOT NULL COMMENT '导演ID',
    INDEX idx_video (video_id),
    INDEX idx_director (director_id)
);

-- 演员表
CREATE TABLE actors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '演员名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

-- 影视演员关联表
CREATE TABLE video_actors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_id BIGINT NOT NULL COMMENT '影视ID',
    actor_id BIGINT NOT NULL COMMENT '演员ID',
    role VARCHAR(100) NOT NULL DEFAULT '' COMMENT '角色名',
    INDEX idx_video (video_id),
    INDEX idx_actor (actor_id)
);

-- 标签表
CREATE TABLE tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '标签名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

-- 影视标签关联表
CREATE TABLE video_tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    video_id BIGINT NOT NULL COMMENT '影视ID',
    tag_id BIGINT NOT NULL COMMENT '标签ID',
    INDEX idx_video (video_id),
    INDEX idx_tag (tag_id)
);

-- 播放源表（全局播放源，如"播放源1"、"采集源A"，不绑定单个影视）
CREATE TABLE play_sources (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '源名称（如"播放源1"、"采集源A"）',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

-- 播放集数表（每个播放源下有多部影视的多集播放链接）
CREATE TABLE play_episodes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_id BIGINT NOT NULL COMMENT '播放源ID',
    video_id BIGINT NOT NULL COMMENT '影视ID',
    episode_number INT NOT NULL DEFAULT 1 COMMENT '集数编号',
    title VARCHAR(255) NOT NULL DEFAULT '' COMMENT '集标题（如"第1集"）',
    play_url VARCHAR(1000) NOT NULL COMMENT '播放地址',
    quality VARCHAR(50) NOT NULL DEFAULT '' COMMENT '清晰度',
    format VARCHAR(20) NOT NULL DEFAULT '' COMMENT '格式（hls/mp4/dash/flv）',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    UNIQUE uk_source_video_episode (source_id, video_id, episode_number),
    INDEX idx_video (video_id)
);

-- 采集源配置表
CREATE TABLE collect_sources (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '采集源名称',
    type TINYINT NOT NULL DEFAULT 1 COMMENT '采集源格式：1默认(系统格式) 2苹果CMS格式',
    collect_url VARCHAR(500) NOT NULL COMMENT '采集地址',
    api_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'API密钥',
    config JSON COMMENT '采集配置',
    cron_expr VARCHAR(100) NOT NULL DEFAULT '' COMMENT '定时采集cron表达式，空表示未开启定时采集',
    play_source_id BIGINT NOT NULL DEFAULT 0 COMMENT '绑定播放源ID，采集到的播放链接存入该播放源',
    last_collect_at TIMESTAMP NULL COMMENT '最后采集时间',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

-- 采集源分类映射表（采集源返回的外部分类映射到系统内分类）
CREATE TABLE collect_source_categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_id BIGINT NOT NULL COMMENT '采集源ID',
    external_category VARCHAR(100) NOT NULL COMMENT '外部分类名称（采集源返回的分类）',
    category_id BIGINT NOT NULL COMMENT '系统内分类ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE uk_source_external (source_id, external_category),
    INDEX idx_category (category_id)
);

-- 采集日志表
CREATE TABLE collect_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    source_id BIGINT NOT NULL COMMENT '采集源ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '采集状态：1成功 2失败 3部分成功',
    total_count INT NOT NULL DEFAULT 0 COMMENT '采集总数',
    success_count INT NOT NULL DEFAULT 0 COMMENT '成功数',
    failed_count INT NOT NULL DEFAULT 0 COMMENT '失败数',
    error_message TEXT COMMENT '错误信息',
    duration_ms INT NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_source (source_id),
    INDEX idx_created (created_at)
);

-- 直播频道表
CREATE TABLE live_channels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '频道名称',
    category VARCHAR(50) NOT NULL DEFAULT '' COMMENT '频道分类',
    stream_url VARCHAR(1000) NOT NULL COMMENT '直播流地址',
    logo VARCHAR(500) NOT NULL DEFAULT '' COMMENT '频道Logo',
    description TEXT COMMENT '频道描述',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

```
### 4.3 主题系统表

```sql

-- 主题表
CREATE TABLE themes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '主题名称',
    identifier VARCHAR(100) NOT NULL UNIQUE COMMENT '主题标识',
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0' COMMENT '版本',
    author VARCHAR(100) NOT NULL DEFAULT '' COMMENT '作者',
    description TEXT COMMENT '描述',
    preview_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '预览图',
    config JSON COMMENT '主题配置（管理员覆盖后的最终配置，合并自theme.json默认值）',
    custom_css TEXT COMMENT '自定义CSS',
    custom_js TEXT COMMENT '自定义JS',
    is_default TINYINT NOT NULL DEFAULT 0 COMMENT '是否默认',
    is_active TINYINT NOT NULL DEFAULT 0 COMMENT '是否当前启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

```
### 4.4 系统配置表

```sql

-- 系统设置表
CREATE TABLE system_settings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    setting_key VARCHAR(100) NOT NULL UNIQUE COMMENT '设置键',
    setting_value TEXT COMMENT '设置值',
    setting_type TINYINT NOT NULL DEFAULT 1 COMMENT '设置类型：1string 2number 3boolean 4json',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 站点模式配置
INSERT INTO system_settings (setting_key, setting_value, setting_type, description) VALUES
('site_mode', 'video_site', 1, '站点模式：video_site(影视站) resource_site(资源站)'),
('api_output_format', 'custom', 1, 'API输出格式：custom(自定义) apple_cms(苹果CMS)'),
('enable_third_party_collect', '1', 3, '是否允许第三方采集'),
('active_theme_id', '1', 2, '当前激活主题ID（与themes.is_active互为冗余，以themes表为准）');

```
### 4.5 用户管理表

```sql

-- 管理员表
CREATE TABLE admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码（加密存储）',
    email VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
    avatar VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
    group_id BIGINT NOT NULL DEFAULT 0 COMMENT '用户组ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_group (group_id)
);

-- 用户组表
CREATE TABLE user_groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '用户组名称',
    permissions JSON COMMENT '权限列表',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

-- 普通用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码（加密存储）',
    email VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邮箱',
    avatar VARCHAR(500) NOT NULL DEFAULT '' COMMENT '头像',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间'
);

```
### 4.6 日志表

```sql

-- 登录日志表
CREATE TABLE login_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_type TINYINT NOT NULL DEFAULT 1 COMMENT '用户类型：1管理员 2普通用户',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    ip_address VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    user_agent VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'User-Agent',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '登录状态：1成功 2失败',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_type, user_id),
    INDEX idx_created (created_at)
);

-- 系统日志表
CREATE TABLE system_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    level TINYINT NOT NULL DEFAULT 1 COMMENT '日志级别：1info 2warning 3error 4critical',
    module VARCHAR(50) NOT NULL DEFAULT '' COMMENT '模块',
    action VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作',
    admin_id BIGINT NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
    content TEXT COMMENT '日志内容',
    ip_address VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_module_created (module, created_at)
);

```
## 5. 技术栈选型

### 5.1 后端技术栈

- **Web框架**：Gin (高性能HTTP框架)
- **ORM**：Bun (高性能ORM)
- **数据库**：MySQL 8.x/9.x 为主开发与生产目标（本阶段默认 MySQL）；PostgreSQL 可选；Redis 7.x
- **模型生成**：迁移后使用 CLI `orange-tv gen model` 从库表生成 Bun 模型
- **API 路径约定**：用户端 `/api/client/v{1,2}`，管理端 `/api/admin/v{1,2}`，内网 `/api/internal/v1`
- **配置管理**：Viper
- **日志系统**：Zap + Lumberjack
- **任务调度**：cron
- **文件存储**：本地存储 + 云存储支持

### 5.2 前端技术栈

- **框架**：React 19 + TypeScript 5.x
- **UI组件库**：逐步引入 shadcn/ui（第二阶段可用原生/轻量组件完成管理页与用户页）
- **状态管理**：Zustand 5
- **路由**：React Router 7
- **样式方案**：可使用 Tailwind CSS / 组件级 CSS（以 monorepo 实际工程为准）
- **构建工具**：Vite 6.x
- **视频播放器**：Video.js 8（第三阶段集成）
- **工程结构**：`web/` monorepo，`apps/client`、`apps/admin`、`packages/shared`

## 6. 开发计划

### 6.1 开发阶段

#### 第一阶段：基础架构搭建（2周）

- [x] Go项目初始化和基础配置
- [x] 数据库设计和迁移（MySQL DDL + `--bun:split`；主实体 `deleted_at` 软删除且不单独建索引；`gen model` 已生成 `internal/model`）
- [x] 基础API框架搭建（路径 `/api/client/v1`、`/api/admin/v1`；模板 User 示例已清除；域路由骨架已注册）
- [x] 前端项目初始化（web monorepo：client + admin + shared）

#### 第二阶段：核心功能开发（4周）

- [x] 用户认证和权限系统（管理员登录 + 仅 `super_admin`；JWT 仅存 ID，请求时查库授权；CLI `admin create`）
- [x] 影视内容管理模块（影视/导演/演员/标签/播放源/剧集）
- [x] 分类管理功能
- [x] 基础的前端页面（管理端内容管理页 + 用户端首页/分类/详情）
- [x] 用户端基础检索（`/api/client/v1/search` 关键词分页；推荐算法不在本阶段）

#### 第三阶段：高级功能开发（3周）

- [ ] 数据采集系统
- [ ] 主题系统
- [ ] 播放器集成
- [ ] 推荐与高级搜索增强（相关推荐、更完善筛选/排序策略等）

#### 第四阶段：系统完善（2周）

- [ ] 系统配置（站点设置、API 配置等）与优化
- [ ] 性能优化
- [ ] 安全加固（含更完整的操作审计/系统日志等）
- [ ] 文档完善

### 6.2 部署方案

#### 6.2.1 开发环境

```bash
# 数据库
docker-compose up -d mysql redis

# 迁移与初始管理员（需配置 JWT_SECRET / jwt.secret）
make build
./build/orange-tv migrate up
./build/orange-tv admin create --username admin --password 'your-password'

# 后端启动
make run-dev
# 或
./build/orange-tv serve

# 前端启动（分别启动两端）
cd web
npm run dev:client   # 用户端，默认 5173
npm run dev:admin    # 管理端，默认 5174
```
#### 6.2.2 生产环境

```bash

# 构建应用

make build

# Docker部署

docker-compose -f docker-compose.prod.yml up -d

# 或传统部署

./orange-tv server --config configs/config.prod.yaml

```
## 7. 附录

### 7.1 苹果CMS兼容格式

```javascript

// 自定义JSON格式
const customFormat = {
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "123",
        "title": "影视标题",
        "cover": "封面地址",
        "category": "分类",
        "year": 2024,
        "rating": 8.5,
        "sources": [
          {
            "name": "播放源1",
            "episodes": [
              {"episode": 1, "url": "播放地址1"},
              {"episode": 2, "url": "播放地址2"}
            ]
          }
        ]
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100
    }
  }
};

// 苹果CMS兼容格式
const appleCmsCompatible = {
  "code": 1,
  "msg": "数据列表",
  "page": 1,
  "pagecount": 5,
  "limit": "20",
  "total": 100,
  "list": [
    // 苹果CMS标准格式
  ]
};

```
### 7.2 影视系统简要流程

```text
用户前端: 首页 -> 分类页 -> 详情页 -> 播放页

管理后台: 影视采集 -> 内容管理 -> 主题配置 -> 系统设置

资源站: 本地影视数据 -> API接口输出 -> 第三方站点采集

```

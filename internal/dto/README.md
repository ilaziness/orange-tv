# DTO Layer

请求/响应数据传输对象。与 `internal/model` 分离，不直接暴露数据库模型。

## 结构

```text
dto/
├── common.go            # 通用类型（分页、ID URI、NamedItem 等与业务数据无关的共享结构）
├── settings.go          # 配置类 DTO（站点/广告/功能开关，admin 编辑 & client 读取共用同一结构）
├── admin/               # 管理端 DTO（字段全量暴露，含 status / sort_order / stream_url / 时间戳等）
│   ├── auth.go
│   ├── category.go      # 分类：CreateCategoryRequest / UpdateCategoryRequest / CategoryResponse
│   ├── video.go         # 影视：VideoListRequest / CreateVideoRequest / UpdateVideoRequest / VideoListItem / VideoDetailResponse / VideoSourceGroup ...
│   ├── metadata.go
│   ├── play.go
│   ├── live.go          # 直播：LiveListRequest / CreateLiveRequest / UpdateLiveRequest / LiveChannelItem / LiveSyncResult
│   └── collect.go       # 采集：CollectSourceListRequest / CreateCollectSourceRequest / CollectSourceItem / CollectLogItem ...
├── client/              # 用户端 DTO（仅暴露对外必要字段，不含 publish_status / stream_url / 时间戳等）
│   ├── category.go      # 分类：CategoryResponse（仅展示字段）
│   ├── video.go         # 影视：VideoListRequest / SearchRequest / VideoListItem / VideoDetailResponse / PlayEpisodeResponse ...
│   ├── live.go          # 直播：LiveListRequest / LiveChannelItem（不含 stream_url / status）
│   ├── settings.go
│   └── user.go
└── open/                # 开放 API DTO（第三方资源站接口，独立字段集）
    ├── category.go
    └── resource.go
```

## 约定

- **按 API 面分包**：管理端在 `dto/admin`，用户端在 `dto/client`，开放 API 在 `dto/open`
- **按业务分类拆文件**：每个文件按业务模块命名（video.go / category.go / live.go ...），同一业务的 Request 和 Response 类型成对放在同一文件中，方便查看和修改
- **业务数据 DTO 不跨端共用**：管理端需要全量字段（含管理属性），对外端只暴露必要字段，避免多余字段泄露
- **共享类型**仅限与业务数据无关的通用结构（分页、ID URI、NamedItem），放在 `dto` 根包
- **配置类 DTO**（settings.go）因 admin 编辑和 client 读取使用相同结构，可共用
- 入参用 `validate` 标签，由 handler 的 `BindAndValidate` 校验
- 不在 DTO 中返回密码等敏感字段

## 响应

HTTP 统一外壳见 `internal/response`：

- 成功：`code = 0`
- 分页：`data.list` / `data.total` / `data.page` / `data.page_size` / `data.total_pages`

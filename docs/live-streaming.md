# 用户直播（主播开播）运维说明

本文说明 Orange TV **用户主播直播** 与外置媒体服务器的分工、SRS 最小部署、管理端推流配置、OBS 填写方式，以及从开功能到观众观看的完整流程。

> **与「电视直播」区分**：内容管理中的 IPTV「电视直播」（`livetv_*`、`/content/livetv`）是另一套业务，本文不涉及。下文「直播」均指用户主播开播（`live` / `streamer`）。

## 1. 架构：业务 API vs 媒体服务器

Orange TV **不经手音视频码流**，只负责：

- 主播资格、申请审核、房间/场次状态
- `stream_key` 生成与重置
- 按模板拼装观众侧播放 URL
- 功能开关 `live_enabled`

外置媒体服务器负责：

- 接收 OBS（等推流端）的 **RTMP 推流**
- 向观众提供 **HLS / HTTP-FLV** 等拉流

```text
OBS  ──RTMP publish──►  媒体服务器  ◄──HTTP pull (HLS/FLV)──  观众浏览器
                              ▲
                              │ 不传码流
                         Orange TV API
                    （密钥、房间状态、解析后的播放 URL）
```

| 方向 | 谁连谁 | 协议（推荐） |
|------|--------|--------------|
| 推流 | OBS → 媒体服 | RTMP（如 `rtmp://media.example.com/live`） |
| 拉流 | 观众 → 媒体服 | HLS（`.m3u8`）和/或 HTTP-FLV（`.flv`） |
| 业务 | 工作室/观众 → Orange TV | HTTPS API；仅返回密钥或**已解析**的播放地址 |

公开观看接口只下发解析后的 `play_hls_url` / `play_flv_url`；**永不**向普通观众返回 `stream_key` 或推流服务器地址。推流服务器与 `stream_key` 仅出现在主播工作室（及管理端）。播放 URL **模板本身不下发客户端**。

## 2. 推荐：SRS 最小部署

以 [SRS](https://ossrs.io/) 为例，最小目标：**RTMP 收流 + HLS / HTTP-FLV 播放**。

### 2.1 功能与端口

| 用途 | 协议 | 常见端口 | 说明 |
|------|------|----------|------|
| OBS 推流 | RTMP | `1935` | 防火墙对主播侧开放 |
| 观众拉流 | HTTP（HLS / HTTP-FLV） | `8080`（或前面 Nginx/HTTPS） | 浏览器访问；生产建议走 HTTPS |

SRS 示例配置思路（路径名可按环境调整，但须与管理端三项配置一致）：

```conf
listen              1935;
max_connections     1000;
srs_log_tank        console;

http_server {
    enabled         on;
    listen          8080;
}

vhost __defaultVhost__ {
    # HTTP-FLV
    http_remux {
        enabled     on;
        mount       [vhost]/[app]/[stream].flv;
    }
    # HLS
    hls {
        enabled         on;
        hls_path        ./objs/nginx/html;
        hls_fragment    2;
        hls_window      10;
    }
}
```

推流应用名常用 `live`，则：

- OBS Server：`rtmp://<媒体服主机>/live`
- Stream Key：房间密钥（如 `a1b2c3...`）
- HLS 播放：`http(s)://<媒体服主机>:<端口>/live/<stream_key>.m3u8`
- FLV 播放：`http(s)://<媒体服主机>:<端口>/live/<stream_key>.flv`

具体 URL 形态以你本机 SRS / 反代规则为准；填入 Orange TV 的模板必须能正确替换 `{stream_key}` 后被浏览器拉到流。

### 2.2 防火墙

- **1935/TCP**：仅需主播推流可达（可限制来源 IP 或 VPN）。
- **HTTP(S) 播放端口**：观众与用户端站点均可访问。
- Orange TV 业务端口与媒体服端口相互独立，不要混为一谈。

### 2.3 HTTPS 与混合内容

若用户端站点是 **HTTPS**，播放地址也必须是 **HTTPS**，否则浏览器会拦截混合内容（黑屏/无法加载）。

建议：

1. 媒体服 HTTP 播放端口前挂 Nginx / Caddy，配置 TLS。
2. 管理端 `live_play_*_template` 使用 `https://...`。
3. 本地纯 HTTP 联调可以暂时用 HTTP 站点 + HTTP 播放地址，上线前再统一为 HTTPS。

CORS：若播放域名与站点域名不同，需在媒体服或反代上为播放路径配置合适的 `Access-Control-Allow-Origin`（HLS/FLV 跨域常见）。

## 3. 管理端推流配置（settings 组 `live`）

管理端：**直播管理 → 推流配置**，对应 API `GET/PUT /api/admin/v1/settings?group=live`。

该组为 **admin-only**：不会出现在用户端 settings 的 `group` 枚举中，客户端拿不到推流服务器或播放模板原文。

| 配置键 | 含义 | 示例 |
|--------|------|------|
| `live_push_server` | OBS「服务器」字段 | `rtmp://media.example.com/live` |
| `live_play_hls_template` | HLS 播放 URL 模板 | `https://media.example.com/live/{stream_key}.m3u8` |
| `live_play_flv_template` | FLV 播放 URL 模板（可空） | `https://media.example.com/live/{stream_key}.flv` |

约定：

- 占位符 **仅** `{stream_key}`；服务端替换后下发解析 URL。
- 开播要求：`live_push_server` 已配置，且 HLS / FLV 模板**至少一个**非空。
- 模板路径中的应用名（如 `/live/`）须与 SRS 推流应用名、反代规则一致。

与 SRS 的对应关系示例：

```text
OBS Server          = live_push_server
                      rtmp://media.example.com/live

HLS 观众 URL        = live_play_hls_template 替换 {stream_key}
                      https://media.example.com/live/「房间密钥」.m3u8

FLV 观众 URL        = live_play_flv_template 替换 {stream_key}
                      https://media.example.com/live/「房间密钥」.flv
```

## 4. OBS 填写方式

主播在用户端 **工作室**（需已是主播）可查看：

- **服务器** = 管理端配置的 `live_push_server`
- **推流码 / 串流密钥** = 该房间的 `stream_key`

OBS 设置示例：

1. 设置 → 推流 → 服务选「自定义」
2. 服务器：粘贴工作室中的推流服务器（即 `live_push_server`）
3. 串流密钥：粘贴工作室中的 `stream_key`（不要带路径前缀）
4. 可先点 OBS「开始推流」做连通性测试；**正式上架列表仍须在工作室点击「开播」**（见 FAQ）

未开播时也可查看密钥，便于提前配置 OBS。直播中不可重置密钥。

## 5. 功能开关与业务流程

### 5.1 功能开关 `live_enabled`

- 所在组：`feature`（与 `livetv_enabled` 等同一套功能设置）
- 形态：四端平台开关（`PlatformFlags`），默认关闭
- 用户端 `/api/client/v1/live/*` 受 `LiveFeatureMiddleware` 保护：关闭时对用户端返回不可用（与电视直播开关行为类似）
- **管理端**直播管理接口不受该中间件阻挡；关闭功能不会自动掐断已在播场次

### 5.2 推荐上线顺序

1. 部署并验证媒体服（推流 + 拉流）
2. 管理端填写 `group=live` 三项推流配置
3. 功能设置中打开对应端的 `live_enabled`
4. 准备至少一个启用的直播分类（迁移会种子若干默认分类）
5. 用户申请主播 → 管理端审核通过 → 用户进入工作室
6. 配置 OBS → 工作室点「开播」→ 观众在 `/live` 列表与 `/live/:id` 观看

流程简图：

```text
开 live_enabled → 申请主播 → 管理端通过 → 工作室开播 → 观众 /live
         ↑              ↑
    推流配置已填     分类可用
```

## 6. FAQ

### 只推流、未点「开播」，列表里没有？

正常。列表只展示业务状态为「直播中」的房间。OBS 推到媒体服 ≠ Orange TV 场次已开始。必须在工作室点击开播后，观众端 `/live` 才会出现该房间。

### 观看黑屏 / 播不出来？

依次检查：

1. **模板**：`live_play_hls_template` / `live_play_flv_template` 是否与媒体服真实路径一致，`{stream_key}` 是否写对
2. **HTTPS 混合内容**：站点 HTTPS 时播放 URL 是否也是 HTTPS
3. **CORS**：跨域播放是否被浏览器拦截
4. **是否真的在推**：媒体服是否已有该 `stream_key` 的流；业务侧是否已点开播（有播放 URL 但媒体侧无流也会黑屏）

### 推流密钥泄露了怎么办？

在工作室（非直播中）使用「重置密钥」；旧密钥立即失效。重置后更新 OBS 串流密钥再推流。直播中需先停播再重置。

### 关了 `live_enabled` 还在播？

功能关闭只挡住用户端直播 API/入口，**不会**自动结束已在播场次。需要停播请主播自行下播，或管理端强制下播/禁播。

### 和「电视直播」配置搞混了？

- 用户主播推流：settings 组 **`live`**（`live_push_server` 等）
- IPTV 电视直播：组 **`livetv`**（如 `livetv_sync_source_url`），与主播开播无关

## 7. 相关路径（便于联调）

| 角色 | 入口 |
|------|------|
| 管理端推流配置 | 直播管理 → 推流配置（`settings?group=live`） |
| 管理端功能开关 | 功能设置 → `live_enabled` |
| 用户端列表 / 观看 | `/live`、`/live/:id` |
| 主播工作室 | `/live/studio`（需登录且已是主播） |

联调检查清单：媒体服可达 → 填 `live` 组 → 开 `live_enabled` → 申请审核 → OBS 推流 → 工作室开播 → 观众拉流 → 管理端强制下播/禁播/撤销（按需验证）。

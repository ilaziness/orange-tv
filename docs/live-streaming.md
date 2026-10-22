# 搭建用户直播

选一种方式：自建 SRS、阿里云直播或七牛直播。管理端路径：**直播 → 推流配置**。

有房间正在直播时不能保存这组配置。先让主播停播，再改厂商或密钥。

主播先在网页点「开始直播」，工作室才会给出推流密钥（含签名时是整串，不要只填流名）。密钥快到期时，在工作室点「刷新推流密钥」，再更新 OBS；页面不会自动改掉已经复制出去的密钥。

密钥输入框留空表示不修改。推流有效期默认 12 小时，播放有效期默认 2 小时，允许 600–86400 秒。

站点用 HTTPS 时，播放地址也必须是 HTTPS。

## 自建 SRS

安装见 [SRS 文档](https://ossrs.io/lts/zh-cn/docs/v5/doc/getting-started)。媒体服和应用分开部署。应用名用 `live`，和推流地址路径一致，例如 `rtmp://media.example.com/live`。

| 用途 | 端口 | 谁能访问 |
|------|------|----------|
| 推流 RTMP | 1935 | 主播 |
| 播放 HTTP | 8080 | 观众，或只给本机 Nginx |
| SRS HTTP API | 1985 | 只绑 `127.0.0.1`，不要映射到公网 |

### 1. 写入 SRS 配置

`http://127.0.0.1:8080` 是 Orange TV 的地址，不是 SRS 的 8080。先在管理端保存配置，再把页面上的回调 URL 贴进 `http_hooks`。

未开鉴权时，删掉下面的 `http_hooks` 整段即可。

```conf
listen              1935;
max_connections     1000;
srs_log_tank        console;

http_api {
    enabled         on;
    listen          127.0.0.1:1985;
}

http_server {
    enabled         on;
    listen          127.0.0.1:8080;
}

vhost __defaultVhost__ {
    http_remux {
        enabled     on;
        mount       [vhost]/[app]/[stream].flv;
    }
    hls {
        enabled         on;
        hls_path        ./objs/nginx/html;
        hls_fragment    2;
        hls_window      10;
    }
    http_hooks {
        enabled         on;
        on_publish      http://127.0.0.1:8080/api/open/v1/live/hooks/srs?hook_secret=管理端回调URL里的密钥;
        on_unpublish    http://127.0.0.1:8080/api/open/v1/live/hooks/srs?hook_secret=管理端回调URL里的密钥;
        on_play         http://127.0.0.1:8080/api/open/v1/live/hooks/srs?hook_secret=管理端回调URL里的密钥;
    }
}
```

改完后重载 SRS。回调没配好时，不带密钥的流名仍然能推。

### 2. 播放鉴权二选一

- 观众直接访问 SRS 的 FLV/HLS：保留上面的 `on_play`，不要再配 Nginx 鉴权。
- 观众访问前面的 Nginx：删掉 `on_play`，只给 m3u8 和 flv 配 `auth_request`。不要给 `.ts` 做鉴权。

Nginx 示例（播放域名和站点不同时需要 CORS）：

```nginx
server {
    listen 443 ssl;
    server_name media.example.com;

    location /live/ {
        auth_request /orange_live_auth;
        alias /usr/local/srs/objs/nginx/html/live/;
        # 播放由 SRS HTTP 提供时，改为：
        # proxy_pass http://127.0.0.1:8080/live/;
        add_header Access-Control-Allow-Origin *;
    }

    location = /orange_live_auth {
        internal;
        proxy_pass http://127.0.0.1:8080/api/open/v1/live/auth/play?hook_secret=回调密钥;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
        proxy_set_header X-Original-URI $request_uri;
    }

    location ~* \.ts$ {
        root /usr/local/srs/objs/nginx/html;
    }
}
```

`proxy_pass` 里的地址同样是 Orange TV，不是 SRS 的 8080。`hook_secret` 用管理端的回调密钥。

### 3. 填写管理端

未开鉴权：

| 字段 | 示例 |
|------|------|
| 厂商 | 自建 SRS |
| 开启自建鉴权 | 关 |
| 推流服务器 | `rtmp://media.example.com/live` |
| HLS 播放模板 | `https://media.example.com/live/{stream_key}.m3u8` |
| FLV 播放模板 | `https://media.example.com/live/{stream_key}.flv` |

打开鉴权后再填：

| 字段 | 填什么 |
|------|--------|
| 签名密钥 | 一长串随机字符。留空表示不改 |
| 回调密钥 | 另一串随机字符，不能和签名密钥相同 |
| SRS API 地址 | `http://127.0.0.1:1985`，Orange TV 要能访问 |
| SRS 回调 URL | 保存后整段复制到 SRS 的 `http_hooks` |
| Nginx 播放鉴权 URL | 只有走 Nginx 鉴权时才用 |

公开站点 URL 在 SEO 设置里。没填时，回调 URL 只有路径，前面自己补上 Orange TV 的对外地址。

### 4. OBS

1. 网页点「开始直播」。
2. OBS → 设置 → 推流 → 服务选自定义。
3. 服务器填推流服务器，例如 `rtmp://media.example.com/live`。
4. 串流密钥填工作室里的整串。开了鉴权时是 `流名?token=...`。

### 5. 检查

- 鉴权和回调都生效后，不带 `token` 推不上去。
- 网页停播后，OBS 断开。
- 改错或改过期播放地址里的 `token`，下一次请求 m3u8 或 flv 失败。
- 页面提示媒体连接未断开时，确认 Orange TV 能访问 1985，或在 SRS 上手动断开该路推流。

## 阿里云直播

### 1. 控制台

1. 开通视频直播。准备推流域名和播放域名，完成 CNAME。域名不要带 `https://` 或路径。
2. 推流和播放都打开 URL 鉴权，方式选 **A**。推流、播放各用一把 Key。
3. 控制台允许的鉴权有效时长不要小于管理端的推流有效期和播放有效期。
4. 建一个 RAM 子账号，只授 `live:ForbidLiveStream` 和 `live:ResumeLiveStream`，记下 AccessKeyId 和 AccessKeySecret。
5. 地域和管理端一致，默认 `cn-shanghai`。AppName 常用 `live`，两边填同一个。

不需要配置阿里云推流回调。

### 2. 填写管理端

| 字段 | 示例 |
|------|------|
| 厂商 | 阿里云直播 |
| 推流域名 | `push.example.com` |
| 播放域名 | `play.example.com` |
| AppName | `live` |
| 地域 | `cn-shanghai` |
| 推流鉴权 Key | 控制台推流鉴权 Key |
| 播放鉴权 Key | 控制台播放鉴权 Key |
| AccessKeyId / AccessKeySecret | 上面那个 RAM 子账号的密钥 |

### 3. OBS

1. 网页点「开始直播」。
2. 服务器：`rtmp://推流域名/AppName`，例如 `rtmp://push.example.com/live`。
3. 串流密钥：工作室里的整串，形如 `流名?auth_key=...`。

### 4. 检查

- 错误的 `auth_key` 推不上去。
- 停播后再用同一条地址重连，会被拒绝。
- 再次开播后可以推。若开播提示恢复失败，房间仍是空闲，到阿里云控制台看该流是否还在禁推列表。
- 页面提示媒体未断开时，到直播控制台手动断开。

## 七牛直播

### 1. 控制台

1. 创建直播空间，记下 Hub 名称。
2. 绑定推流域名和播放域名，完成 CNAME。域名不要带协议和路径。
3. 推流鉴权使用限时管理凭证（expiry sk）。
4. 播放打开时间戳防盗链。防盗链密钥填成和七牛账号 **SecretKey** 相同的值。

不需要配置七牛回调。

### 2. 填写管理端

| 字段 | 示例 |
|------|------|
| 厂商 | 七牛直播 |
| Hub | 控制台空间名 |
| 推流域名 | `push.example.com` |
| 播放域名 | `play.example.com` |
| AccessKey / SecretKey | 七牛账号密钥。SecretKey 与播放防盗链密钥相同 |

### 3. OBS

1. 网页点「开始直播」。
2. 服务器：`rtmp://推流域名/Hub`，例如 `rtmp://push.example.com/你的Hub`。
3. 串流密钥：工作室里的整串，形如 `流名?e=...&token=...`。

### 4. 检查

- 错误签名推不上去。
- 停播后重连被拒绝。
- 再次开播可以推。
- 页面提示媒体未断开时，到七牛直播控制台断开该流。

## 常见问题

**没点开播就推不上去。** 先在网页点「开始直播」，再用当时给出的密钥。

**HTTPS 页面播不了。** 播放地址也要是 `https://`。自建的 HLS/FLV 模板填 `https://`。

**画面停住，但 OBS 还显示在推。** 已连上的 RTMP/FLV 不会因为签名到期自己断开。临近到期时在工作室刷新密钥，再让 OBS 重连。HLS 会在下一次拉 m3u8 时换新地址。

**网页已经停播，OBS 还在推。** 到 SRS、阿里云禁推列表或七牛控制台手动断开，并确认 Orange TV 能访问对应的 API 地址、密钥正确。

# 搭建用户直播

选一种方式：自建 SRS、阿里云直播或七牛直播。管理端路径：**直播 → 推流配置**。

有房间正在直播时不能保存这组配置。先让主播停播，再改厂商或密钥。

站点用 HTTPS 时，播放地址也必须是 HTTPS。

## 自建 SRS

以下按 **SRS 6** 说明，安装见 [SRS 6 文档](https://ossrs.io/lts/zh-cn/docs/v6/doc/getting-started)，推荐 Docker 部署：

```sh
docker run -d --name srs --restart always \
  -p 1935:1935 -p 8080:8080 \
  -v ./srs.conf:/usr/local/srs/conf/srs.conf \
  ossrs/srs:6
```

> SRS HTTP API（1985）只在 Orange TV 与 SRS 分机部署时才需要映射端口，并要用防火墙只放行 Orange TV 服务器，不要暴露到公网。

本节假定 Orange TV 的服务地址为 `https://tv.example.com`（站点对外地址，经 Nginx 等反代到 443），下文所有回调、鉴权地址都按它填写，替换成你的实际地址即可。SRS 6 全部使用默认端口：RTMP 1935、HTTP API 1985、HTTP Server 8080。

媒体服（SRS）和应用（Orange TV）分开部署。应用名用 `live`，和推流地址路径一致，例如 `rtmp://media.example.com/live`。

| 用途         | 端口 | 说明                                                                    |
| ------------ | ---- | ----------------------------------------------------------------------- |
| 推流 RTMP    | 1935 | 主播访问                                                                |
| 播放 HTTP    | 8080 | 观众直连时绑 `0.0.0.0`；只给本机 Nginx 反代时绑 `127.0.0.1`             |
| SRS HTTP API | 1985 | Orange TV 用它断开播流。分机部署时绑定内网地址并只放行 Orange TV 服务器 |

> Orange TV 默认端口也是 8080。假定 Orange TV 经域名对外服务时不会与 SRS 的默认 8080 冲突；如果两者部署在同一台机器且直接按端口访问，必须改掉其中一个端口，并同步更新播放模板和下文地址。

### 1. 写入 SRS 配置

先在管理端保存配置，再把页面给出的回调 URL 贴进 `http_hooks`。回调 URL 的主机部分必须填 **SRS 所在机器能访问到的 Orange TV 服务地址**，本节假定为 `https://tv.example.com`。

```conf
daemon off;         # Docker 必须前台运行；源码部署后台运行时可去掉或改为 on
listen              1935;
max_connections     1000;
srs_log_tank        console;

http_api {
    enabled         on;
    listen          1985;
}

http_server {
    enabled         on;
    listen          8080;
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
        on_publish      https://tv.example.com/api/open/v1/live/hooks/srs?hook_secret=管理端回调URL里的密钥;
        on_unpublish    https://tv.example.com/api/open/v1/live/hooks/srs?hook_secret=管理端回调URL里的密钥;
        on_play         https://tv.example.com/api/open/v1/live/hooks/srs?hook_secret=管理端回调URL里的密钥;
    }
}
```

改完后重载 SRS：源码安装 `./etc/init.d/srs reload`，Docker 直接 `docker restart srs`。回调没配好时，不带密钥的流名仍然能推。

### 2. 播放鉴权二选一

- 观众直接访问 SRS 的 FLV/HLS：播放模板直接写 SRS 的播放地址，保留上面的 `on_play`，不要再配 Nginx 鉴权。
- 观众访问前面的 Nginx：删掉 `on_play`，只给 m3u8 和 flv 配 `auth_request`。`.ts` 分片不做鉴权（鉴权接口也不校验 `.ts`），由 m3u8 保护。

Nginx 示例（播放域名和站点不同时需要 CORS）：

```nginx
server {
    listen 443 ssl;
    server_name media.example.com;

    # m3u8 / flv：先鉴权，再转发给 SRS
    location /live/ {
        auth_request /orange_live_auth;
        proxy_pass http://127.0.0.1:8080;    # SRS 的 http_server
        add_header Access-Control-Allow-Origin *;
    }

    # ts 分片：不鉴权，直接转发给 SRS
    location ~* \.ts$ {
        proxy_pass http://127.0.0.1:8080;    # SRS 的 http_server
        add_header Access-Control-Allow-Origin *;
    }

    location = /orange_live_auth {
        internal;
        proxy_pass https://tv.example.com/api/open/v1/live/auth/play?hook_secret=回调密钥;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
        proxy_set_header X-Original-URI $request_uri;
    }
}
```

两个 `proxy_pass` 不是同一个服务：播放和 ts 分片转发的是 SRS 的 http_server（默认 8080），鉴权转发的是前面假定的 Orange TV 服务地址 `https://tv.example.com`。`hook_secret` 用管理端的回调密钥。

### 3. 填写管理端

| 字段               | 填什么                                                                                  |
| ------------------ | ---------------------------------------------------------------------------------------- |
| 厂商               | 自建 SRS                                                                                 |
| 开启自建鉴权       | 开                                                                                       |
| 推流服务器         | `rtmp://media.example.com/live`                                                          |
| HLS 播放模板       | `https://media.example.com/live/{stream_key}.m3u8`                                       |
| FLV 播放模板       | `https://media.example.com/live/{stream_key}.flv`                                        |
| 签名密钥           | 一长串随机字符。留空表示不改                                                             |
| 回调密钥           | 另一串随机字符，不能和签名密钥相同                                                       |
| 本服务 API 地址    | Orange TV 后端对外可达根地址，如 `https://api.example.com`（不要带路径）                   |
| SRS API 地址       | 同机填 `http://127.0.0.1:1985`，分机填 `http://192.168.1.10:1985`。Orange TV 必须能访问   |
| SRS 回调 URL       | 按本服务 API 地址自动生成，保存后整段复制到 SRS 的 `http_hooks`。`hook_secret` 参数就是上面的回调密钥 |
| Nginx 播放鉴权 URL | 按本服务 API 地址自动生成，只有走 Nginx 鉴权时才用。`hook_secret` 参数同样是回调密钥     |

### 4. OBS

1. 网页点「开始直播」。
2. OBS → 设置 → 推流 → 服务选自定义。
3. 服务器填推流服务器，例如 `rtmp://media.example.com/live`。
4. 串流密钥填工作室里的整串，形如 `流名?token=...`。

### 5. 检查

- 鉴权和回调都生效后，不带 `token` 推不上去。
- 网页停播后，OBS 断开。
- 改错或改过期播放地址里的 `token`，下一次请求 m3u8 或 flv 失败。
- 页面提示媒体连接未断开时，确认 Orange TV 能访问 SRS API 地址（1985），或在 SRS 上手动断开该路推流。

## 阿里云直播

### 1. 控制台

1. 开通视频直播。准备推流域名和播放域名，完成 CNAME。域名不要带 `https://` 或路径。
2. 推流和播放都打开 URL 鉴权，方式选 **A**。推流、播放各用一把 Key。
3. 控制台允许的鉴权有效时长不要小于管理端的推流有效期和播放有效期。
4. 建一个 RAM 子账号，只授 `live:ForbidLiveStream` 和 `live:ResumeLiveStream`，记下 AccessKeyId 和 AccessKeySecret。
5. 地域和管理端一致，默认 `cn-shanghai`。AppName 常用 `live`，两边填同一个。

不需要配置阿里云推流回调。

### 2. 填写管理端

| 字段                          | 示例                      |
| ----------------------------- | ------------------------- |
| 厂商                          | 阿里云直播                |
| 推流域名                      | `push.example.com`        |
| 播放域名                      | `play.example.com`        |
| AppName                       | `live`                    |
| 地域                          | `cn-shanghai`             |
| 推流鉴权 Key                  | 控制台推流鉴权 Key        |
| 播放鉴权 Key                  | 控制台播放鉴权 Key        |
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

| 字段                  | 示例                                         |
| --------------------- | -------------------------------------------- |
| 厂商                  | 七牛直播                                     |
| Hub                   | 控制台空间名                                 |
| 推流域名              | `push.example.com`                           |
| 播放域名              | `play.example.com`                           |
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

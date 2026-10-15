# 用户直播部署说明

用户主播开播依赖**外置媒体服务器**：Orange TV 不处理音视频码流，只拼装推流地址与观众播放 URL。推荐使用 [SRS](https://ossrs.io/)。

```text
主播 OBS  ──RTMP──►  媒体服务器  ◄──HTTPS 拉流──  观众浏览器
```

## 1. 部署媒体服务器

最小目标：**RTMP 收流 + HLS（和/或 HTTP-FLV）播放**。

### 1.1 端口与防火墙

| 用途 | 协议 | 常见端口 | 防火墙 |
|------|------|----------|--------|
| 主播推流 | RTMP | `1935` | 对主播可达（可限制来源 IP / VPN） |
| 观众拉流 | HTTP(S) | `8080`，或经反代 `443` | 对公网观众开放 |

与 Orange TV 业务端口相互独立，不要混用。

### 1.2 SRS 示例配置

应用名常用 `live`，须与后文管理端三项配置中的路径一致：

```conf
listen              1935;
max_connections     1000;
srs_log_tank        console;

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
}
```

对应 URL：

| 用途 | 地址 |
|------|------|
| OBS 服务器 | `rtmp://<媒体服主机>/live` |
| OBS 串流密钥 | 房间 `stream_key`（不加路径前缀） |
| HLS 播放 | `https://<媒体服主机>/live/<stream_key>.m3u8` |
| FLV 播放 | `https://<媒体服主机>/live/<stream_key>.flv` |

Docker 快速验证见 [SRS 官方文档](https://ossrs.io/lts/zh-cn/docs/v5/doc/getting-started)；生产使用固定配置与持久化目录。

### 1.3 HTTPS 与 CORS

站点为 **HTTPS** 时，播放地址也必须是 **HTTPS**，否则浏览器拦截混合内容。

1. 在媒体服 HTTP 播放端口前挂 Nginx / Caddy，配置 TLS。
2. 播放域名与站点域名不同时，为播放路径加 CORS。
3. 本地联调可用全 HTTP；上线前统一为 HTTPS。

Nginx 示例：

```nginx
server {
    listen 443 ssl;
    server_name media.example.com;
    # ssl_certificate / ssl_certificate_key ...

    location /live/ {
        proxy_pass http://127.0.0.1:8080/live/;
        proxy_http_version 1.1;
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods 'GET, OPTIONS';
        if ($request_method = OPTIONS) { return 204; }
    }
}
```

RTMP `1935` 一般不经 HTTP 反代，防火墙直接放行即可。

### 1.4 自检

1. OBS 推到 `rtmp://<主机>/live`，串流密钥用测试串（如 `testkey`）。
2. 浏览器能打开 `https://<主机>/live/testkey.m3u8`（或 `.flv`）并播出画面。

## 2. 写入管理端推流配置

媒体服自检通过后，在 **直播管理 → 推流配置** 填写（须与上节 URL 一致）：

| 配置键 | 含义 | 示例 |
|--------|------|------|
| `live_push_server` | OBS「服务器」 | `rtmp://media.example.com/live` |
| `live_play_hls_template` | HLS 模板 | `https://media.example.com/live/{stream_key}.m3u8` |
| `live_play_flv_template` | FLV 模板（可空） | `https://media.example.com/live/{stream_key}.flv` |

- 占位符仅 `{stream_key}`。
- `live_push_server` 必填；HLS / FLV 至少一个非空。
- 模板中的 `/live/` 须与 SRS 应用名、反代 `location` 一致。

然后在功能设置中打开 `live_enabled`。

## 3. 常见问题

### 观看黑屏 / 播不出来？

1. 播放模板是否与媒体服真实路径一致  
2. 站点 HTTPS 时播放 URL 是否也是 HTTPS  
3. 跨域是否被 CORS 拦截  
4. 媒体服是否已有该 `stream_key` 的流  

### 推流通了但观众列表没有？

OBS 推到媒体服不等于业务已开播；主播还需在工作室点击「开播」。

# 资源站开放 API

第四阶段提供第三方采集/资源站数据输出接口。

## 路径前缀

- `GET /api/open/v1/videos` — 影视列表
- `GET /api/open/v1/videos/:id` — 影视详情（含播放源/剧集）
- `GET /api/open/v1/categories` — 启用中的分类树

JWT 不校验；密钥由 `system_settings.resource_api_key` 控制。

## 访问控制

| 条件 | 行为 |
|------|------|
| `enable_third_party_collect=false` | 返回业务码「资源站 API 已关闭」 |
| 已配置 `resource_api_key` | 必须携带密钥 |
| 未配置密钥 | 开放访问（仍受开关控制） |

密钥传递方式（任选其一）：

- Header：`X-API-Key: <key>`
- Query：`?key=<key>` 或 `?api_key=<key>`

## 输出格式

Query `format`：

| 值 | 说明 |
|------|------|
| **不传 / 空 / `default`** | **系统默认格式**（本站自有 JSON） |
| `apple_cms` | 苹果 CMS 兼容列表/详情 |

其它非空 `format` 值将返回参数错误。

后台「API 输出格式」`api_output_format` 仅允许：`default` | `apple_cms`，作为**请求未带 format 时**的默认。

示例：

```bash
# 系统默认格式（推荐：不传 format）
curl "http://localhost:8080/api/open/v1/videos?page=1&page_size=20" \
  -H "X-API-Key: your-key"

# 或显式 default
curl "http://localhost:8080/api/open/v1/videos?page=1&page_size=20&format=default" \
  -H "X-API-Key: your-key"

# 苹果 CMS 兼容
curl "http://localhost:8080/api/open/v1/videos?page=1&limit=20&format=apple_cms" \
  -H "X-API-Key: your-key"

curl "http://localhost:8080/api/open/v1/videos/1?format=apple_cms" \
  -H "X-API-Key: your-key"
```

## 管理端配置

- 页面：系统设置 → API 配置
- API：`GET/PUT /api/admin/v1/settings`
- 字段：`api.site_mode`、`api.api_output_format`（`default` / `apple_cms`）、`api.enable_third_party_collect`、`api.resource_api_key`（更新时空串表示不修改）

## 相关

- 站点公开信息：`GET /api/client/v1/site`
- 缓存：列表/详情短 TTL（约 2 分钟）；配置变更会失效 settings 缓存

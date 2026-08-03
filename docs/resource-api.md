# 资源站开放 API

第三方采集接口

## 影视列表

```bash
GET /api/open/v1/videos?page=1&page_size=20
```

参数：

- `page`：页码，默认 `1`
- `page_size` / `limit`：每页数量，默认 `20`，最大 `100`

响应：

```json
{
  "code": 0,
  "data": {
    "list": [
      { "id": 1, "title": "标题", "category_id": 2, "created_at": "2026-08-03 10:00:00" }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

## 影视详情

```bash
GET /api/open/v1/videos/detail?id=1&id=2&id=3
```

参数：

- `id`：视频 id，可重复，必填，最多 50 个

响应：

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "title": "标题",
      "subtitle": "",
      "cover": "...",
      "category_id": 2,
      "year": 2026,
      "rating": 8.5,
      "release_date": "2026-08-01",
      "region": "中国大陆",
      "language": "国语",
      "description": "简介",
      "directors": ["导演"],
      "actors": ["演员"],
      "sources": [
        {
          "id": 1,
          "name": "默认源",
          "episodes": [
            { "episode": 1, "title": "第1集", "url": "..." }
          ]
        }
      ],
      "created_at": "2026-08-03 10:00:00"
    }
  ]
}
```

## 分类列表

```bash
GET /api/open/v1/categories
```

响应：

```json
{
  "code": 0,
  "data": [
    { "id": 1, "name": "电影", "parent_id": 0 },
    { "id": 2, "name": "动作", "parent_id": 1 }
  ]
}
```

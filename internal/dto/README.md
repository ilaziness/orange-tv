# DTO Layer

请求/响应数据传输对象。与 `internal/model` 分离，不直接暴露数据库模型。

## 结构

```text
dto/
├── common.go            # 共享类型（分页、IDURI、分类/影视公共响应等）
├── admin/               # 管理端 DTO
│   ├── auth.go
│   ├── category.go
│   ├── video.go
│   ├── metadata.go
│   └── play.go
└── client/              # 用户端 DTO
    └── video.go
```

## 约定

- **按 API 面分包**：管理端写操作在 `dto/admin`，用户端只读查询在 `dto/client`
- **共享类型**放 `dto` 根包（如分页、通用列表/详情卡片）
- 入参用 `validate` 标签，由 handler 的 `BindAndValidate` 校验
- 不在 DTO 中返回密码等敏感字段

## 响应

HTTP 统一外壳见 `internal/response`：

- 成功：`code = 0`
- 分页：`data.list` / `data.total` / `data.page` / `data.page_size` / `data.total_pages`

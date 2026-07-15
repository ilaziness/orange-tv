# Service Layer

业务逻辑层：协调 repository、执行规则校验与事务。

## 结构

```text
service/
├── admin/               # 管理端业务
│   ├── auth.go
│   ├── category.go
│   ├── video.go
│   ├── metadata.go
│   ├── play.go
│   ├── live.go
│   ├── collect.go       # 采集源/任务/日志 + cron 调度
│   └── theme.go
└── client/              # 用户端业务（只读浏览/搜索/播放辅助）
    ├── category.go
    ├── video.go         # 含 related / 搜索增强
    ├── live.go
    └── theme.go
```

## 约定

- 构造函数注入 repository / JWT 等依赖；在 `internal/app` 组装
- 不依赖 Gin；不直接写 HTTP 响应
- 错误使用 `internal/errcode`，数据库错误向上包装
- 列表/详情默认过滤软删除；删除冲突策略见 PRD 4.0
- 管理端与用户端能力分离，避免在同一 service 混入两侧入口

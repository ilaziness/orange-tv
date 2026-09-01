# Service Layer

业务逻辑层：协调 repository、执行规则校验与事务。

## 结构

```text
service/
├── settings.go          # 多端共享的设置读取工具函数（StrVal/IntVal/BoolVal 等）
├── admin/               # 管理端业务
│   ├── auth.go
│   ├── category.go
│   ├── video.go
│   ├── metadata.go
│   ├── play.go
│   ├── livetv.go
│   ├── settings.go      # 管理端设置服务（按组别获取/更新）
│   └── collect.go       # 采集源/任务/日志 + cron 调度
├── client/              # 用户端业务（只读浏览/搜索/播放辅助）
│   ├── category.go
│   ├── video.go         # 含 related / 搜索增强
│   ├── settings.go      # 客户端设置服务（白名单过滤）
│   └── livetv.go
└── open/                # 开放 API 业务
    └── resource.go      # 第三方资源站接口（仅判断是否允许采集）
```

## 约定

- 构造函数注入 repository / JWT 等依赖；在 `internal/app` 组装
- 不依赖 Gin；不直接写 HTTP 响应
- 错误使用 `internal/errcode`，数据库错误向上包装
- 列表/详情默认过滤软删除；删除冲突策略见相关业务文档
- 管理端与用户端能力分离，避免在同一 service 混入两侧入口

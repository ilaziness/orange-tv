# Repository Layer

数据访问层：只负责持久化查询与写入，不包含业务规则编排。

## 当前实现

```text
repository/
├── admin.go
├── category.go
├── video.go
├── metadata.go
├── play.go              # WithTx 支持采集入库同事务写剧集
├── live.go
└── collect.go
```

## 约定

- 使用 Bun 官方包访问 MySQL/PostgreSQL/SQLite
- 主实体查询默认 `deleted_at IS NULL`；软删除用 `Update` 写 `deleted_at`
- 返回 `internal/model`，不返回 DTO
- 事务通过 `RunInTx` / `WithTx`（影视关联、采集写剧集等）
- 复杂业务校验放在 service，不把领域策略堆进 repository

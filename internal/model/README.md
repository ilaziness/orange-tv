# Model Layer

数据模型定义目录。表结构由迁移创建后，通过 CLI 从数据库生成：

```bash
./build/orange-tv gen model --output ./internal/model --package model --json-tags
# 或单表
./build/orange-tv gen model --table videos
```

## 约定

- 使用 Bun 标签映射列
- JSON 标签用于 API 序列化
- 业务主实体含 `deleted_at` 软删除字段（不建单独索引）
- 勿手写/覆盖全量 model；以 gen 输出为准，必要时做薄封装

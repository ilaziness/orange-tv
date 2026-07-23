# 常用命令

构建产物路径：`build/orange-tv`（Windows 为 `build/orange-tv.exe`）。以下 CLI 示例使用 Unix 路径。

```bash
# 构建和运行
make build          # 构建应用
make run            # 运行应用
make run-dev        # 使用开发配置运行

# 测试
make test           # 运行测试
make test-coverage  # 运行测试并生成覆盖率报告

# 代码质量
make fmt            # 格式化代码
make vet            # 运行 go vet
make lint           # 运行 linter（需已安装 golangci-lint，见下方「开发工具」）
```

## Swagger

生成物输出至 [`docs/swagger/`](../../docs/swagger/)（`docs.go`、`swagger.json`、`swagger.yaml`）。`main.go` 通过 blank import 注册 Swagger 元数据。

```bash
make swagger        # 生成 Swagger 文档
make swagger-clean  # 删除 docs/swagger/ 下生成文件
```

需先安装 [swag](https://github.com/swaggo/swag)：

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## 开发工具

`make lint`、`make swagger`、`make mock` 不会自动安装工具；命令不在 PATH 时会报错并提示安装命令：

| 命令 | 工具 | 安装 |
|------|------|------|
| `make lint` | golangci-lint | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2` |
| `make swagger` | swag | `go install github.com/swaggo/swag/cmd/swag@latest` |
| `make mock` | mockgen | `go install go.uber.org/mock/mockgen@latest` |

## 数据库

默认 [`configs/config.yaml`](../../configs/config.yaml) 使用 **MySQL**；迁移脚本为 **MySQL DDL**。可用 `docker compose up -d mysql` 启动本地库。

```bash
./build/orange-tv migrate create <name>   # 创建迁移
./build/orange-tv migrate up              # 执行迁移
./build/orange-tv migrate down            # 回滚迁移
./build/orange-tv migrate status          # 查看迁移状态
./build/orange-tv gen model               # 从数据库生成 Bun 模型到 internal/model
./build/orange-tv gen model --table videos
./build/orange-tv gen model --with-relations # 同时读取数据库物理外键
```

生成每个模型文件时会先用等价于 `gofmt` 的方式格式化再写入，避免因缩进/空格差异导致每次生成都产生无意义 diff。

### 模型业务关联

数据库不建立外键时，在 [`configs/model-relations.yaml`](../../configs/model-relations.yaml) 声明逻辑关系。`gen model` 默认加载该文件，并生成 Bun 的双向关联字段：子模型的 `belongs-to` 和父模型的 `has-many`；关联字段固定为 `json:"-"`，应通过 DTO 对外输出。

```yaml
relations:
  # 简写：源表.外键列 -> 目标表.被引用列
  - videos.category_id -> categories.id

  # 同一张目标表有多个角色时，指定生成的字段名
  - source: videos.created_by
    target: users.id
    field: Creator
    reverse_field: CreatedVideos
```

通过 `--relations <path>` 可使用其他关系文件。`--with-relations` 会把物理外键与 YAML 关系合并；相同关系按一条生成，YAML 的字段名覆盖优先。当前仅支持单列关联；复合外键需要保持手工模型字段。

SQL 迁移多语句请用独立行 `--bun:split` 分隔（Bun 要求；单文件内多条 `CREATE` 不能靠分号自动拆分）。

模型应通过 `gen model` 生成，避免手写全量表结构。

失败迁移清理：Bun 失败时仍会记为已应用，需先 `migrate down` 回滚后再 `migrate up` 重试。

JWT 说明：未配置 `jwt.secret` 时全局不挂 JWT 中间件，管理端骨架接口可直接联调；配置 secret 后自动启用管理端 `RequireAuth`，同时用户端路径与登录接口走 `jwt.skip_paths`。

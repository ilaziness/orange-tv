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

默认 [`configs/config.yaml`](../../configs/config.yaml) 使用 SQLite；示例迁移 SQL 亦为 **SQLite 语法**。

```bash
./build/orange-tv migrate create <name>   # 创建迁移
./build/orange-tv migrate up              # 执行迁移
./build/orange-tv migrate down            # 回滚迁移
./build/orange-tv migrate status          # 查看迁移状态
./build/orange-tv gen model               # 从数据库生成模型
```

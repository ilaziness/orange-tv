# 代码修改后验证流程

每次完成代码修改后，执行以下步骤确保可编译且符合规范。

## 步骤

### 1. 格式化

```bash
go fmt ./...
```

### 2. 静态检查

```bash
go vet ./...
```

### 3. 编译

推荐：

```bash
make build
```

等效命令（与 Makefile 一致）：

```bash
go build -ldflags "-X github.com/ilaziness/orange-tv/cmd.version=1.0.0" -o build/orange-tv ./main.go
```

Windows 下产物为 `build/orange-tv.exe`。

### 4. 测试

```bash
go test ./...
```

或：

```bash
make test
```

### 5. Lint

需已安装 golangci-lint（见 [commands.md](commands.md#开发工具)）：

```bash
make lint
```

### 6. 一键验证

```bash
make fmt && make vet && make build && make test && make lint
```

## 验证失败处理

- **格式化错误**：运行 `go fmt ./...` 自动修复
- **vet 警告**：根据提示修改代码
- **编译错误**：检查 `internal/app` 中的依赖组装、包导入、语法错误
- **测试失败**：查看测试输出，修复逻辑错误

# 目录结构

```shell
orange-tv/
├── main.go                          # 程序入口，调用 cmd.Execute()
├── Makefile                         # 构建、测试、运行等快捷命令
├── go.mod / go.sum                  # 模块依赖
├── cmd/                             # CLI 命令层（Cobra：serve、config、migrate、gen 等）
├── configs/                         # 配置文件（config.yaml 及环境覆盖）
├── migrations/                      # 数据库迁移脚本（MySQL *.up.sql / *.down.sql）
├── docs/                            # 项目文档
│   ├── agents/                      # Agent 专题文档（从 AGENTS.md 拆分）
│   ├── PRD.md                       # 产品需求
│   └── swagger/                     # Swagger 生成代码与 OpenAPI 规范（make swagger）
├── web/                             # 前端 monorepo（apps/client、apps/admin）
├── scripts/                         # 构建、部署、测试脚本
├── data/                            # 本地运行时数据
├── test/                            # 集成测试
├── internal/                        # 私有业务代码（外部不可导入）
│   ├── app/                         # 应用组装与生命周期（手动 DI 唯一入口）
│   ├── auth/                        # JWT 认证与令牌管理
│   ├── cache/                       # 缓存抽象与实现（内存、Redis、多级）
│   ├── config/                      # 配置结构体、加载与校验
│   ├── constant/                    # 应用级常量
│   ├── crypto/                      # 密码哈希等加密工具
│   ├── database/                    # 数据库连接与 Bun ORM 初始化；gen model 生成器
│   ├── dto/                         # 请求/响应 DTO（共享 common + admin/client 子包）
│   ├── errcode/                     # 业务错误码定义
│   ├── event/                       # 发布/订阅事件总线
│   ├── handler/                     # 协议层请求处理器
│   │   └── http/                    # HTTP Handler 公共工具 + health/stub
│   │       ├── admin/               # 管理端 Handler
│   │       └── client/              # 用户端 Handler
│   ├── logger/                      # Zap 日志封装
│   ├── metrics/                     # Prometheus 指标采集与暴露
│   ├── middleware/                  # 协议层中间件
│   │   └── http/                    # HTTP 中间件（认证、限流、CORS 等）
│   ├── model/                       # Bun 数据模型（由 gen model 从库表生成）
│   ├── repository/                  # 数据访问层
│   ├── response/                    # 统一 HTTP 响应结构
│   ├── router/                      # Gin 路由注册：/api/client、/api/admin、/api/internal
│   ├── server/                      # HTTP 服务器实现
│   ├── service/                     # 业务逻辑层（admin/client 子包）
│   ├── testutil/                    # 测试辅助（业务 handler 桩等）
│   ├── tracing/                     # OpenTelemetry 链路追踪
│   ├── utils/                       # 通用工具（协程池等）
│   └── validator/                   # 请求参数校验器
└── build/                           # 编译产物目录（gitignore）
```

## 相关文档

- Agent 文档索引：[agents/README.md](README.md)
- 模块启用/删除：[module-usage.md](../module-usage.md)

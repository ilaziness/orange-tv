# 扩展 HTTP 业务模块

按层自下而上添加代码，最后在 `internal/app/http.go` 的 `wireHTTP()` 中追加 wiring（与现有 user 等模块并列，非替换整个函数）。

## 1. 创建模型（Model）

在 `internal/model/` 下新建模型文件，使用 Bun 标签：

```go
package model

import "github.com/uptrace/bun"

type Order struct {
    bun.BaseModel `bun:"table:orders,alias:o"`

    ID        int64   `bun:"id,pk,autoincrement"`
    UserID    int64   `bun:"user_id,notnull"`
    Amount    float64 `bun:"amount,notnull"`
    Status    string  `bun:"status,notnull,default:'pending'"`
    CreatedAt int64   `bun:"created_at,notnull"`
}
```

## 2. 创建 DTO（可选）

在 `internal/dto/` 下新建 DTO 文件：

```go
package dto

type CreateOrderRequest struct {
    UserID int64   `json:"user_id" validate:"required"`
    Amount float64 `json:"amount" validate:"required,gt=0"`
}

type OrderResponse struct {
    ID     int64   `json:"id"`
    UserID int64   `json:"user_id"`
    Amount float64 `json:"amount"`
    Status string  `json:"status"`
}
```

## 3. 创建 Repository

在 `internal/repository/` 下新建仓库文件：

```go
package repository

import (
    "context"

    "github.com/ilaziness/orange-tv/internal/database"
    "github.com/ilaziness/orange-tv/internal/model"
)

type OrderRepository interface {
    Create(ctx context.Context, order *model.Order) error
    GetByID(ctx context.Context, id int64) (*model.Order, error)
}

type orderRepo struct {
    db *database.DB
}

func NewOrderRepo(db *database.DB) OrderRepository {
    return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, order *model.Order) error {
    _, err := r.db.NewInsert().Model(order).Exec(ctx)
    return err
}

func (r *orderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
    order := new(model.Order)
    err := r.db.NewSelect().Model(order).Where("id = ?", id).Scan(ctx)
    return order, err
}
```

## 4. 创建 Service

在 `internal/service/` 下新建服务文件：

```go
package service

import (
    "context"

    "github.com/ilaziness/orange-tv/internal/dto"
    "github.com/ilaziness/orange-tv/internal/repository"
)

type OrderService interface {
    CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
}

type orderService struct {
    orderRepo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) OrderService {
    return &orderService{orderRepo: orderRepo}
}

func (s *orderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    // 业务逻辑实现
    return nil, nil
}
```

## 5. 创建 Handler

在 `internal/handler/http/` 下新建 Handler 文件：

```go
package http

import (
    "github.com/ilaziness/orange-tv/internal/dto"
    "github.com/ilaziness/orange-tv/internal/response"
    "github.com/ilaziness/orange-tv/internal/service"
    "github.com/gin-gonic/gin"
)

type OrderHandler struct {
    orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
    return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Create(c *gin.Context) {
    var req dto.CreateOrderRequest
    if !BindAndValidate(c, &req) {
        return
    }

    resp, err := h.orderService.CreateOrder(c.Request.Context(), &req)
    if err != nil {
        HandleServiceError(c, err)
        return
    }
    response.Success(c, resp)
}
```

## 6. 注册路由

1. 在 [`internal/router/handlers.go`](internal/router/handlers.go) 的 `Handlers` 中追加 handler 字段：

```go
type Handlers struct {
    Health *httphandler.HealthHandler
    Order  *httphandler.OrderHandler // 新增
}
```

2. 在对应 API 面的路由文件（如 [`internal/router/client.go`](internal/router/client.go)）中追加领域注册函数，并在 `registerClientRoutes` 中调用：

```go
func registerClientRoutes(engine *gin.Engine, h *Handlers) {
    v1 := engine.Group(PathClientV1)
    v2 := engine.Group(PathClientV2)
    registerClientOrderRoutes(v1, v2, h.Order) // 新增调用
}

func registerClientOrderRoutes(v1, v2 *gin.RouterGroup, order *httphandler.OrderHandler) {
    orders := v1.Group("/orders")
    orders.POST("", order.Create)
}
```

API 路径前缀常量定义于 [`internal/router/paths.go`](internal/router/paths.go)（用户端 `/api/client/v1`，管理端 `/api/admin/v1`，内网 `/api/internal/v1`）。仅用户端暴露的模块不必修改 `admin.go` / `internal.go`。

## 7. 在 app 包中组装依赖

在 `internal/app/http.go` 的 `wireHTTP()` 中组装新模块并挂到 `Handlers`：

```go
orderRepo := repository.NewOrderRepo(a.db)
orderSvc := service.NewOrderService(orderRepo)

handlers, err := router.NewHandlers(healthHandler)
if err != nil {
	return err
}
handlers.Order = httphandler.NewOrderHandler(orderSvc) // 新增字段时在此赋值

httpServer, err := server.NewHTTPServer(a.cfg, a.log, handlers, a.metrics, a.jwtMgr)
if err != nil {
	return err
}
a.httpServer = httpServer
```

## 8. 编写测试

为 Repository、Service、Handler 编写对应的 `_test.go` 文件。

## 9. 创建数据库迁移（如需要）

```bash
./build/orange-tv migrate create create_orders_table
```

编辑生成的 `migrations/*.up.sql` 和 `migrations/*.down.sql` 文件。

## 10. 配置（可选）

如需新增配置项，在 `internal/config/config.go` 补充字段与默认值。

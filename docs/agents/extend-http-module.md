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

DTO 按 API 面分子包：`internal/dto/admin`、`internal/dto/client`；共享类型放 `internal/dto`。

```go
package admin

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

Service 按 API 面分子包：`internal/service/admin`、`internal/service/client`。

```go
package admin

import (
    "context"

    admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
    "github.com/ilaziness/orange-tv/internal/repository"
)

type OrderService interface {
    CreateOrder(ctx context.Context, req *admindto.CreateOrderRequest) (*admindto.OrderResponse, error)
}

type orderService struct {
    orderRepo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) OrderService {
    return &orderService{orderRepo: orderRepo}
}
```

## 5. 创建 Handler

Handler 同样分子包：`internal/handler/http/admin`、`internal/handler/http/client`。公共绑定工具在 `internal/handler/http`。

```go
package admin

import (
    "github.com/gin-gonic/gin"
    admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
    httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
    "github.com/ilaziness/orange-tv/internal/response"
    adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

type OrderHandler struct {
    orderService adminsvc.OrderService
}

func NewOrderHandler(orderService adminsvc.OrderService) *OrderHandler {
    return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Create(c *gin.Context) {
    var req admindto.CreateOrderRequest
    if !httphandler.BindAndValidate(c, &req) {
        return
    }
    resp, err := h.orderService.CreateOrder(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, resp)
}
```

## 6. 注册路由

1. 在 [`internal/router/handlers.go`](internal/router/handlers.go) 的 `Handlers` 中追加 handler 字段（例如 `AdminOrder` / `ClientOrder`）。
2. 在对应 API 面路由文件中**直接注册**字符串路径，不要做 `if handler != nil` 回退：

```go
func registerAdminContentRoutes(v1 *gin.RouterGroup, h *Handlers) {
    v1.POST("/orders", h.AdminOrder.Create)
}
```

API 版本前缀常量定义于 [`internal/router/paths.go`](internal/router/paths.go)（仅系统路径与 `/api/{client|admin|internal}/v{1,2}`）。业务路径在注册时直接写字符串。

## 7. 在 app 包中组装依赖

在 `internal/app/http.go` 的 `wireHTTP()` 中组装并赋值到 `Handlers`：

```go
orderRepo := repository.NewOrderRepo(a.db)
orderSvc := adminsvc.NewOrderService(orderRepo)

handlers, err := router.NewHandlers(healthHandler)
if err != nil {
	return err
}
handlers.AdminOrder = adminhandler.NewOrderHandler(orderSvc)

httpServer, err := server.NewHTTPServer(a.cfg, a.log, accessLogger, handlers, a.metrics, a.jwtMgr)
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

package router

import (
	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerAdminRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathAdminV1)
	v1.Use(httpmiddleware.RequireAuth())

	registerAdminUserRoutes(v1, h.User)

	// 注册 v2 版本（如需要）
	// v2 := engine.Group(PathAdminV2)
	// v2.Use(httpmiddleware.RequireAuth())
}

func registerAdminUserRoutes(v1 *gin.RouterGroup, user *httphandler.UserHandler) {
	users := v1.Group("/users")
	users.GET("/:id", user.GetUser)
	users.POST("", user.CreateUser)
	users.PUT("/:id", user.UpdateUser)
	users.DELETE("/:id", user.DeleteUser)
	users.GET("", user.ListUsers)
}

package router

import (
	"time"

	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerClientRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathClientV1)
	// v1.Use(...) // JWT、限流等用户端中间件

	v2 := engine.Group(PathClientV2)
	// v2.Use(...)

	registerClientUserRoutes(v1, v2, h.User)
}

func registerClientUserRoutes(v1, v2 *gin.RouterGroup, user *httphandler.UserHandler) {
	v1Users := v1.Group("/users")
	{
		v1Users.GET("/:id", user.GetUser)
		v1Users.POST("", user.CreateUser)
		v1Users.PUT("/:id", user.UpdateUser)
		v1Users.DELETE("/:id", user.DeleteUser)

		sunsetDate := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
		v1Users.GET("", httpmiddleware.Deprecated(sunsetDate, PathClientV2Users), user.ListUsers)
	}

	v2Users := v2.Group("/users")
	{
		v2Users.GET("/:id", user.GetUser)
		v2Users.POST("", user.CreateUser)
		v2Users.PUT("/:id", user.UpdateUser)
		v2Users.DELETE("/:id", user.DeleteUser)
		v2Users.GET("", user.ListUsers)
	}
}

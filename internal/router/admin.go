package router

import (
	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerAdminRoutes(engine *gin.Engine, h *Handlers) {
	// 登录始终公开
	publicAuth := engine.Group(PathAdminV1Auth)
	publicAuth.POST("/login", h.Stub.NotImplemented)

	v1 := engine.Group(PathAdminV1)
	if h.RequireAdminAuth {
		v1.Use(httpmiddleware.RequireAuth())
	}
	registerAdminProtectedAuthRoutes(v1, h.Stub)
	registerAdminContentRoutes(v1, h.Stub)
	registerAdminUserRoutes(v1, h.Stub)
	registerAdminSystemRoutes(v1, h.Stub)
}

func registerAdminProtectedAuthRoutes(v1 *gin.RouterGroup, stub *httphandler.StubHandler) {
	auth := v1.Group("/auth")
	auth.POST("/logout", stub.NotImplemented)
	auth.GET("/profile", stub.NotImplemented)
}

func registerAdminContentRoutes(v1 *gin.RouterGroup, stub *httphandler.StubHandler) {
	// categories
	v1.GET("/categories", stub.EmptyArray)
	v1.POST("/categories", stub.NotImplemented)
	v1.PUT("/categories/:id", stub.NotImplemented)
	v1.DELETE("/categories/:id", stub.NotImplemented)

	// videos
	v1.GET("/videos", stub.EmptyList)
	v1.POST("/videos", stub.NotImplemented)
	v1.PUT("/videos/:id", stub.NotImplemented)
	v1.DELETE("/videos/:id", stub.NotImplemented)

	// play sources / episodes
	v1.GET("/play-sources", stub.EmptyList)
	v1.POST("/play-sources", stub.NotImplemented)
	v1.PUT("/play-sources/:id", stub.NotImplemented)
	v1.DELETE("/play-sources/:id", stub.NotImplemented)

	v1.GET("/play-episodes", stub.EmptyList)
	v1.POST("/play-episodes", stub.NotImplemented)
	v1.PUT("/play-episodes/:id", stub.NotImplemented)
	v1.DELETE("/play-episodes/:id", stub.NotImplemented)

	// directors / actors / tags
	v1.GET("/directors", stub.EmptyList)
	v1.POST("/directors", stub.NotImplemented)
	v1.PUT("/directors/:id", stub.NotImplemented)
	v1.DELETE("/directors/:id", stub.NotImplemented)

	v1.GET("/actors", stub.EmptyList)
	v1.POST("/actors", stub.NotImplemented)
	v1.PUT("/actors/:id", stub.NotImplemented)
	v1.DELETE("/actors/:id", stub.NotImplemented)

	v1.GET("/tags", stub.EmptyList)
	v1.POST("/tags", stub.NotImplemented)
	v1.PUT("/tags/:id", stub.NotImplemented)
	v1.DELETE("/tags/:id", stub.NotImplemented)

	// live
	v1.GET("/live", stub.EmptyList)
	v1.POST("/live", stub.NotImplemented)
	v1.PUT("/live/:id", stub.NotImplemented)
	v1.DELETE("/live/:id", stub.NotImplemented)

	// collect
	v1.GET("/collect-sources", stub.EmptyList)
	v1.POST("/collect-sources", stub.NotImplemented)
	v1.PUT("/collect-sources/:id", stub.NotImplemented)
	v1.DELETE("/collect-sources/:id", stub.NotImplemented)
	v1.GET("/collect-sources/:id/categories", stub.EmptyArray)
	v1.POST("/collect-sources/:id/categories", stub.NotImplemented)
	v1.POST("/collect/:source_id/start", stub.NotImplemented)
	v1.POST("/collect/:source_id/stop", stub.NotImplemented)
	v1.GET("/collect/logs", stub.EmptyList)
}

func registerAdminUserRoutes(v1 *gin.RouterGroup, stub *httphandler.StubHandler) {
	v1.GET("/admins", stub.EmptyList)
	v1.POST("/admins", stub.NotImplemented)
	v1.PUT("/admins/:id", stub.NotImplemented)
	v1.DELETE("/admins/:id", stub.NotImplemented)

	v1.GET("/groups", stub.EmptyList)
	v1.POST("/groups", stub.NotImplemented)
	v1.PUT("/groups/:id", stub.NotImplemented)
	v1.DELETE("/groups/:id", stub.NotImplemented)

	v1.GET("/users", stub.EmptyList)
	v1.PUT("/users/:id", stub.NotImplemented)
	v1.DELETE("/users/:id", stub.NotImplemented)
	v1.GET("/users/:id/login-logs", stub.EmptyList)
}

func registerAdminSystemRoutes(v1 *gin.RouterGroup, stub *httphandler.StubHandler) {
	v1.GET("/settings", stub.NotImplemented)
	v1.PUT("/settings", stub.NotImplemented)

	v1.GET("/themes", stub.EmptyList)
	v1.POST("/themes", stub.NotImplemented)
	v1.PUT("/themes/:id", stub.NotImplemented)
	v1.DELETE("/themes/:id", stub.NotImplemented)
	v1.POST("/themes/:id/activate", stub.NotImplemented)
}

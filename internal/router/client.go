package router

import (
	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
)

func registerClientRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathClientV1)
	// v1.Use(...) // JWT、限流等用户端中间件（业务阶段再挂载）

	v2 := engine.Group(PathClientV2)
	// v2.Use(...)

	registerClientContentRoutes(v1, h.Stub)
	_ = v2
}

func registerClientContentRoutes(v1 *gin.RouterGroup, stub *httphandler.StubHandler) {
	v1.GET("/categories", stub.EmptyArray)
	v1.GET("/videos", stub.EmptyList)
	v1.GET("/videos/:id", stub.NotImplemented)
	v1.GET("/search", stub.EmptyList)
	v1.GET("/live", stub.EmptyList)
	v1.GET("/theme/current", stub.NotImplemented)
}

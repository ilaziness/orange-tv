package router

import (
	"github.com/gin-gonic/gin"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerInternalRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathInternalV1)
	v1.Use(httpmiddleware.InternalServiceAuth(h.InternalServiceKey))

	// 内网业务路由按需注册
	_ = v1
}

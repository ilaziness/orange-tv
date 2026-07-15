package router

import (
	"github.com/gin-gonic/gin"
)

func registerClientRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathClientV1)
	v2 := engine.Group(PathClientV2)

	registerClientContentRoutes(v1, h)
	_ = v2
}

func registerClientContentRoutes(v1 *gin.RouterGroup, h *Handlers) {
	v1.GET("/categories", h.ClientCategory.List)
	v1.GET("/videos", h.ClientVideo.List)
	v1.GET("/videos/:id", h.ClientVideo.Get)
	v1.GET("/search", h.ClientVideo.Search)
	v1.GET("/videos/:id/related", h.ClientVideo.Related)

	v1.GET("/live", h.ClientLive.List)
	v1.GET("/theme/current", h.ClientTheme.Current)
}

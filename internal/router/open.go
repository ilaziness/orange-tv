package router

import (
	"github.com/gin-gonic/gin"
)

func registerOpenRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathOpenV1)
	v1.GET("/videos", h.OpenResource.ListVideos)
	v1.GET("/videos/:id", h.OpenResource.GetVideo)
	v1.GET("/categories", h.OpenResource.ListCategories)
}

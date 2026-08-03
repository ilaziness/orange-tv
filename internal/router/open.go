package router

import (
	"github.com/gin-gonic/gin"
	httpmw "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerOpenRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathOpenV1)
	v1.Use(httpmw.OpenResourceMiddleware(h.OpenResource.Service()))
	v1.GET("/videos", h.OpenResource.ListVideos)
	v1.GET("/videos/detail", h.OpenResource.GetVideo)
	v1.GET("/categories", h.OpenResource.ListCategories)
}

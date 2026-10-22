package router

import (
	"github.com/gin-gonic/gin"
	httpmw "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerOpenRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathOpenV1)
	// SRS callbacks and Nginx play auth are not gated by the resource-station switch
	// or the client live feature flag, and they do not use a user JWT.
	v1.POST("/live/hooks/srs", h.OpenLive.SRSHook)
	v1.GET("/live/auth/play", h.OpenLive.PlayAuth)

	resource := v1.Group("")
	resource.Use(httpmw.OpenResourceMiddleware(h.OpenResource.Service()))
	resource.GET("/videos", h.OpenResource.ListVideos)
	resource.GET("/videos/detail", h.OpenResource.GetVideo)
	resource.GET("/categories", h.OpenResource.ListCategories)
}

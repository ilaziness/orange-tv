package router

import (
	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerAdminRoutes(engine *gin.Engine, h *Handlers) {
	// login is always public
	publicAuth := engine.Group(PathAdminV1 + "/auth")
	publicAuth.POST("/login", h.AdminAuth.Login)

	v1 := engine.Group(PathAdminV1)
	if h.RequireAdminAuth {
		v1.Use(httpmiddleware.RequireSuperAdmin(func(c *gin.Context, adminID int64) error {
			_, _, err := h.AuthService.EnsureSuperAdmin(c.Request.Context(), adminID)
			return err
		}))
	}

	registerAdminProtectedAuthRoutes(v1, h)
	registerAdminContentRoutes(v1, h)
	registerAdminUserRoutes(v1, h.Stub)
	registerAdminSystemRoutes(v1, h)
}

func registerAdminProtectedAuthRoutes(v1 *gin.RouterGroup, h *Handlers) {
	auth := v1.Group("/auth")
	auth.POST("/logout", h.AdminAuth.Logout)
	auth.GET("/profile", h.AdminAuth.Profile)
}

func registerAdminContentRoutes(v1 *gin.RouterGroup, h *Handlers) {
	// categories
	v1.GET("/categories", h.AdminCategory.List)
	v1.POST("/categories", h.AdminCategory.Create)
	v1.PUT("/categories/:id", h.AdminCategory.Update)
	v1.DELETE("/categories/:id", h.AdminCategory.Delete)

	// videos
	v1.GET("/videos", h.AdminVideo.List)
	v1.GET("/videos/:id", h.AdminVideo.Get)
	v1.POST("/videos", h.AdminVideo.Create)
	v1.PUT("/videos/:id", h.AdminVideo.Update)
	v1.DELETE("/videos/:id", h.AdminVideo.Delete)

	// play sources / episodes
	v1.GET("/play-sources", h.AdminPlay.ListSources)
	v1.POST("/play-sources", h.AdminPlay.CreateSource)
	v1.PUT("/play-sources/:id", h.AdminPlay.UpdateSource)
	v1.DELETE("/play-sources/:id", h.AdminPlay.DeleteSource)
	v1.GET("/play-episodes", h.AdminPlay.ListEpisodes)
	v1.POST("/play-episodes", h.AdminPlay.CreateEpisode)
	v1.PUT("/play-episodes/:id", h.AdminPlay.UpdateEpisode)
	v1.DELETE("/play-episodes/:id", h.AdminPlay.DeleteEpisode)

	// directors / actors / tags
	v1.GET("/directors", h.AdminMetadata.ListDirectors)
	v1.POST("/directors", h.AdminMetadata.CreateDirector)
	v1.PUT("/directors/:id", h.AdminMetadata.UpdateDirector)
	v1.DELETE("/directors/:id", h.AdminMetadata.DeleteDirector)

	v1.GET("/actors", h.AdminMetadata.ListActors)
	v1.POST("/actors", h.AdminMetadata.CreateActor)
	v1.PUT("/actors/:id", h.AdminMetadata.UpdateActor)
	v1.DELETE("/actors/:id", h.AdminMetadata.DeleteActor)

	v1.GET("/tags", h.AdminMetadata.ListTags)
	v1.POST("/tags", h.AdminMetadata.CreateTag)
	v1.PUT("/tags/:id", h.AdminMetadata.UpdateTag)
	v1.DELETE("/tags/:id", h.AdminMetadata.DeleteTag)

	// live
	v1.GET("/live", h.AdminLive.List)
	v1.POST("/live", h.AdminLive.Create)
	v1.PUT("/live/:id", h.AdminLive.Update)
	v1.DELETE("/live/:id", h.AdminLive.Delete)

	v1.GET("/collect-sources", h.AdminCollect.ListSources)
	v1.POST("/collect-sources", h.AdminCollect.CreateSource)
	v1.PUT("/collect-sources/:id", h.AdminCollect.UpdateSource)
	v1.DELETE("/collect-sources/:id", h.AdminCollect.DeleteSource)
	v1.GET("/collect-sources/:id/categories", h.AdminCollect.ListCategories)
	v1.POST("/collect-sources/:id/categories", h.AdminCollect.SetCategories)
	v1.POST("/collect/:source_id/start", h.AdminCollect.Start)
	v1.POST("/collect/:source_id/stop", h.AdminCollect.Stop)
	v1.GET("/collect/logs", h.AdminCollect.ListLogs)
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

func registerAdminSystemRoutes(v1 *gin.RouterGroup, h *Handlers) {
	v1.GET("/settings", h.AdminSettings.Get)
	v1.PUT("/settings", h.AdminSettings.Update)

	v1.GET("/system-logs", h.AdminLog.ListSystemLogs)
	v1.GET("/login-logs", h.AdminLog.ListLoginLogs)

	v1.GET("/themes", h.AdminTheme.List)
	v1.POST("/themes", h.Stub.NotImplemented) // theme upload later
	v1.PUT("/themes/:id", h.AdminTheme.Update)
	v1.DELETE("/themes/:id", h.Stub.NotImplemented)
	v1.POST("/themes/:id/activate", h.AdminTheme.Activate)
}

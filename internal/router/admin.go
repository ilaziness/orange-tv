package router

import (
	"github.com/gin-gonic/gin"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerAdminRoutes(engine *gin.Engine, h *Handlers) {
	// login is always public
	publicAuth := engine.Group(PathAdminV1 + "/auth")
	publicAuth.POST("/login", h.AdminAuth.Login)

	v1 := engine.Group(PathAdminV1)
	if h.RequireAdminAuth {
		v1.Use(httpmiddleware.RequireSuperAdmin(func(c *gin.Context, adminID uint32) error {
			_, _, err := h.AuthService.EnsureSuperAdmin(c.Request.Context(), adminID)
			return err
		}))
	}

	registerAdminProtectedAuthRoutes(v1, h)
	registerAdminContentRoutes(v1, h)
	registerAdminUserRoutes(v1, h)
	registerAdminSystemRoutes(v1, h)
	registerAdminManagementRoutes(v1, h)
}

func registerAdminProtectedAuthRoutes(v1 *gin.RouterGroup, h *Handlers) {
	auth := v1.Group("/auth")
	auth.POST("/logout", h.AdminAuth.Logout)
	auth.GET("/profile", h.AdminAuth.Profile)
	auth.PUT("/profile", h.AdminAuth.UpdateProfile)
	auth.PUT("/profile/password", h.AdminAuth.ChangePassword)
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
	v1.POST("/play-episodes/batch-status", h.AdminPlay.BatchUpdateEpisodeStatus)
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
	v1.POST("/live/sync", h.AdminLive.Sync)

	// comments
	v1.GET("/comments", h.AdminComment.List)
	v1.GET("/comments/:id/parents", h.AdminComment.GetParents)
	v1.PUT("/comments/:id/status", h.AdminComment.UpdateStatus)
	v1.DELETE("/comments/:id", h.AdminComment.Delete)

	v1.GET("/collect-sources", h.AdminCollect.ListSources)
	v1.POST("/collect-sources", h.AdminCollect.CreateSource)
	v1.PUT("/collect-sources/:id", h.AdminCollect.UpdateSource)
	v1.DELETE("/collect-sources/:id", h.AdminCollect.DeleteSource)
	v1.GET("/collect-sources/:id/categories", h.AdminCollect.ListCategories)
	v1.POST("/collect-sources/:id/categories", h.AdminCollect.SetCategories)
	v1.GET("/collect-sources/:id/remote-categories", h.AdminCollect.FetchRemoteCategories)
	v1.POST("/collect-sources/:id/schedule/enable", h.AdminCollect.EnableSchedule)
	v1.POST("/collect-sources/:id/schedule/disable", h.AdminCollect.DisableSchedule)
	v1.POST("/collect-sources/:id/collect", h.AdminCollect.CollectNow)
	v1.GET("/collect/logs", h.AdminCollect.ListLogs)
}

func registerAdminUserRoutes(v1 *gin.RouterGroup, h *Handlers) {
	// Admins (A3)
	v1.GET("/admins", h.AdminMgmt.ListAdmins)
	v1.POST("/admins", h.AdminMgmt.CreateAdmin)
	v1.PUT("/admins/:id", h.AdminMgmt.UpdateAdmin)
	v1.DELETE("/admins/:id", h.AdminMgmt.DeleteAdmin)
	v1.PUT("/admins/:id/password", h.AdminMgmt.ResetAdminPassword)

	// User groups (A4)
	v1.GET("/groups", h.AdminMgmt.ListGroups)
	v1.POST("/groups", h.AdminMgmt.CreateGroup)
	v1.PUT("/groups/:id", h.AdminMgmt.UpdateGroup)
	v1.DELETE("/groups/:id", h.AdminMgmt.DeleteGroup)

	// Regular users (A5)
	v1.GET("/users", h.AdminMgmt.ListUsers)
	v1.POST("/users", h.AdminMgmt.CreateUser)
	v1.PUT("/users/:id", h.AdminMgmt.UpdateUser)
	v1.DELETE("/users/:id", h.AdminMgmt.DeleteUser)
	v1.PUT("/users/:id/password", h.AdminMgmt.ResetUserPassword)
}

func registerAdminManagementRoutes(v1 *gin.RouterGroup, h *Handlers) {
	// Dashboard (A1)
	v1.GET("/dashboard", h.AdminMgmt.Dashboard)

	// Batch video ops (A2)
	v1.POST("/videos/batch/publish-status", h.AdminMgmt.BatchUpdatePublishStatus)
	v1.POST("/videos/batch/delete", h.AdminMgmt.BatchDeleteVideos)

	// Banners (C1 admin)
	v1.GET("/banners", h.AdminMgmt.ListBanners)
	v1.POST("/banners", h.AdminMgmt.CreateBanner)
	v1.PUT("/banners/:id", h.AdminMgmt.UpdateBanner)
	v1.DELETE("/banners/:id", h.AdminMgmt.DeleteBanner)

	// Ads
	v1.GET("/ads", h.AdminAd.ListAds)
	v1.POST("/ads", h.AdminAd.CreateAd)
	v1.PUT("/ads/:id", h.AdminAd.UpdateAd)
	v1.DELETE("/ads/:id", h.AdminAd.DeleteAd)
}

func registerAdminSystemRoutes(v1 *gin.RouterGroup, h *Handlers) {
	v1.GET("/settings", h.AdminSettings.Get)
	v1.PUT("/settings", h.AdminSettings.Update)

	v1.GET("/system-logs", h.AdminLog.ListSystemLogs)
	v1.GET("/admin-login-logs", h.AdminLog.ListAdminLoginLogs)
	v1.GET("/user-login-logs", h.AdminMgmt.ListUserLoginLogs)
	v1.GET("/app-logs", h.AdminLog.ListAppLogs)

	// data management
	v1.GET("/data/backup", h.AdminData.Backup)
	v1.POST("/data/batch-update/preview", h.AdminData.BatchUpdatePreview)
	v1.POST("/data/batch-update/execute", h.AdminData.BatchUpdateExecute)
}

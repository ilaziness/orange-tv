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
	v1.GET("/site", h.ClientSite.Public)
	v1.GET("/banners", h.ClientBanner.List)

	// User auth (C5) — public
	v1.POST("/auth/register", h.ClientUser.Register)
	v1.POST("/auth/login", h.ClientUser.Login)

	// User profile (C5) — requires JWT
	v1.GET("/auth/profile", h.ClientUser.Profile)

	// Favorites (C6) — requires JWT
	v1.GET("/favorites", h.ClientUser.ListFavorites)
	v1.POST("/favorites/:id", h.ClientUser.AddFavorite)
	v1.DELETE("/favorites/:id", h.ClientUser.RemoveFavorite)

	// Play history (C6) — requires JWT
	v1.GET("/history", h.ClientUser.ListHistory)
	v1.POST("/history", h.ClientUser.UpsertHistory)
	v1.DELETE("/history/:id", h.ClientUser.DeleteHistory)
	v1.DELETE("/history", h.ClientUser.ClearHistory)

	// Comments (C6) — list is public, create/delete require JWT
	v1.GET("/videos/:id/comments", h.ClientUser.ListComments)
	v1.POST("/comments", h.ClientUser.CreateComment)
	v1.DELETE("/comments/:id", h.ClientUser.DeleteComment)
}

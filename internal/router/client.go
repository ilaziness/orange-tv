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
	v1.GET("/videos/:id/episodes/:source_id/:episode_id", h.ClientVideo.GetEpisode)
	v1.GET("/search", h.ClientVideo.Search)
	v1.GET("/videos/:id/related", h.ClientVideo.Related)

	livetv := v1.Group("/livetv", h.LiveTVFeature)
	livetv.GET("", h.ClientLiveTV.List)
	livetv.GET("/play/:id", h.ClientLiveTV.Play)
	livetv.HEAD("/play/:id", h.ClientLiveTV.Play)

	v1.GET("/settings", h.ClientSettings.GetSettings)
	v1.GET("/banners", h.ClientBanner.List)
	v1.GET("/promotions", h.ClientAd.List)

	// User auth (C5) — public
	v1.GET("/auth/captcha", h.ClientUser.Captcha)
	v1.POST("/auth/register", h.ClientUser.Register)
	v1.POST("/auth/login", h.ClientUser.Login)
	v1.POST("/auth/refresh", h.ClientUser.Refresh)

	// User profile (C5) — requires JWT
	v1.GET("/auth/profile", h.ClientUser.Profile)
	v1.PUT("/auth/profile", h.ClientUser.UpdateProfile)
	v1.PUT("/auth/profile/password", h.ClientUser.ChangePassword)
	v1.GET("/auth/login-history", h.ClientUser.LoginHistory)

	// Favorites (C6) — requires JWT
	v1.GET("/favorites", h.ClientUser.ListFavorites)
	v1.GET("/favorites/:id", h.ClientUser.CheckFavorite)
	v1.POST("/favorites/:id", h.ClientUser.AddFavorite)
	v1.DELETE("/favorites/:id", h.ClientUser.RemoveFavorite)

	// Play history (C6) — requires JWT
	v1.GET("/history", h.ClientUser.ListHistory)
	v1.GET("/history/:id", h.ClientUser.GetHistory)
	v1.POST("/history", h.ClientUser.UpsertHistory)
	v1.DELETE("/history/:id", h.ClientUser.DeleteHistory)
	v1.DELETE("/history", h.ClientUser.ClearHistory)

	// Comments (C6) — list/replies are public, create/vote require JWT
	v1.GET("/videos/:id/comments", h.ClientUser.ListComments)
	v1.GET("/comments/:id/replies", h.ClientUser.ListReplies)
	v1.POST("/comments", h.ClientUser.CreateComment)
	v1.POST("/comments/:id/vote", h.ClientUser.VoteComment)

	// Ratings (C6) — get is public, create requires JWT
	v1.GET("/ratings/:id", h.ClientUser.GetRating)
	v1.POST("/ratings/:id", h.ClientUser.RateVideo)
}

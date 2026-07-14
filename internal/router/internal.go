package router

import (
	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
)

func registerInternalRoutes(engine *gin.Engine, h *Handlers) {
	v1 := engine.Group(PathInternalV1)
	v1.Use(httpmiddleware.InternalServiceAuth(h.InternalServiceKey))

	registerInternalUserRoutes(v1, h.User)
}

func registerInternalUserRoutes(v1 *gin.RouterGroup, user *httphandler.UserHandler) {
	users := v1.Group("/users")
	users.GET("/:id", user.GetUser)
}

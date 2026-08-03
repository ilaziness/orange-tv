package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	opensvc "github.com/ilaziness/orange-tv/internal/service/open"
)

// OpenResourceMiddleware blocks the open resource API when third-party collect is disabled.
// When blocked, it returns 404 so the endpoints appear not to exist.
func OpenResourceMiddleware(svc opensvc.ResourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !svc.Enabled(c.Request.Context()) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

package http

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders returns middleware that sets baseline security response headers for APIs.
// Swagger UI pages are skipped: CSP default-src 'none' would break the UI assets.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/swagger") {
			c.Next()
			return
		}
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "0")
		// API-oriented CSP: allow nothing by default for document loads.
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}

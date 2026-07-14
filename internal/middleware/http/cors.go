package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a default CORS middleware.
// Note: By default, AllowCredentials is false to avoid conflict with AllowOrigins: ["*"].
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false, // Disabled by default for security
		MaxAge:           12 * time.Hour,
	})
}

// CORSWithConfig returns a CORS middleware with custom configuration.
func CORSWithConfig(allowOrigins []string, allowMethods []string, allowHeaders []string, exposeHeaders []string, allowCredentials bool, maxAge time.Duration) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
	})
}

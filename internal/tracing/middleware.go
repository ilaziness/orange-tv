// Package tracing provides distributed tracing support using OpenTelemetry.
package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Middleware returns a Gin middleware for OpenTelemetry tracing.
// Paths in skipPaths are excluded from span creation (e.g. health probes).
func Middleware(cfg *config.Config, skipPaths ...string) gin.HandlerFunc {
	skipSet := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skipSet[p] = struct{}{}
	}

	base := otelgin.Middleware(cfg.App.Name)
	return func(c *gin.Context) {
		if _, skip := skipSet[c.FullPath()]; skip {
			c.Next()
			return
		}
		base(c)
	}
}

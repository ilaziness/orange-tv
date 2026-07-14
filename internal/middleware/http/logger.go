package http

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a middleware that logs HTTP requests using zap.
// Paths in skipPaths are excluded from access logs (e.g. health probes).
func Logger(logger *zap.Logger, skipPaths ...string) gin.HandlerFunc {
	return ginzap.GinzapWithConfig(logger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        false,
		SkipPaths:  skipPaths,
	})
}

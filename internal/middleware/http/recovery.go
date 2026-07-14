package http

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics and logs them.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return ginzap.RecoveryWithZap(logger, true)
}

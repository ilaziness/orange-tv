// Package http provides HTTP middleware implementations.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/service"
	"go.uber.org/zap"
)

// LiveFeatureMiddleware returns a middleware that blocks client live endpoints
// when the live feature is disabled in system settings. It returns a plain
// HTTP 404 so callers see the route as non-existent.
func LiveFeatureMiddleware(settings service.SettingsService, log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	return func(c *gin.Context) {
		if settings == nil {
			log.Warn("live feature middleware: settings service is nil, blocking live requests")
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		m, err := settings.LoadMapByGroup(c.Request.Context(), constant.SettingGroupFeature)
		if err != nil {
			log.Error("live feature middleware: load feature settings failed", zap.Error(err))
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if !service.BoolVal(m, constant.SettingFeatureLiveEnabled, false) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}

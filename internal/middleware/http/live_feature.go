package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/clienttype"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/service"
	"go.uber.org/zap"
)

// LiveFeatureMiddleware returns a middleware that blocks client /live endpoints
// when the live feature is disabled for the request's client type. It returns a
// plain HTTP 404 so callers see the route as non-existent.
func LiveFeatureMiddleware(settings service.SettingsService, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		m, err := settings.LoadMapByGroup(c.Request.Context(), constant.SettingGroupFeature)
		if err != nil {
			log.Error("live feature middleware: load feature settings failed", zap.Error(err))
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		clientType := clienttype.FromContext(c.Request.Context())
		if !service.PlatformBoolVal(m, constant.SettingFeatureLiveEnabled, clientType, false) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}

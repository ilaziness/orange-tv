package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLiveFeatureMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		m          map[string]model.SystemSettings
		clientType string
		loadErr    error
		wantStatus int
	}{
		{
			name: "live enabled for web",
			m: map[string]model.SystemSettings{
				constant.SettingFeatureLiveEnabled: {SettingValue: `{"web":true,"desktop":false,"app":false,"tv":false}`},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "live disabled for web",
			m: map[string]model.SystemSettings{
				constant.SettingFeatureLiveEnabled: {SettingValue: `{"web":false,"desktop":true,"app":true,"tv":true}`},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "live enabled for app",
			m: map[string]model.SystemSettings{
				constant.SettingFeatureLiveEnabled: {SettingValue: `{"web":false,"desktop":false,"app":true,"tv":false}`},
			},
			clientType: constant.ClientTypeApp,
			wantStatus: http.StatusOK,
		},
		{
			name:       "live missing",
			m:          map[string]model.SystemSettings{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "settings load error",
			m:          nil,
			loadErr:    errcode.InternalError,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &stubSettingsService{m: tt.m, err: tt.loadErr}
			mw := LiveFeatureMiddleware(settings, zap.NewNop())

			called := false
			engine := gin.New()
			engine.Use(ClientTypeMiddleware())
			engine.Use(mw)
			engine.GET("/live", func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/live", nil)
			if tt.clientType != "" {
				req.Header.Set(constant.ClientTypeHeader, tt.clientType)
			}
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.True(t, called)
			} else {
				assert.False(t, called)
			}
		})
	}
}

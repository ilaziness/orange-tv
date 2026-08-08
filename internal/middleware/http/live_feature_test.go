package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type stubSettingsService struct {
	m   map[string]model.SystemSettings
	err error
}

func (s *stubSettingsService) LoadMapByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	return s.m, s.err
}

func (s *stubSettingsService) LoadGroupMaps(ctx context.Context, groups []string) (map[string]map[string]model.SystemSettings, error) {
	return nil, nil
}

func (s *stubSettingsService) MapGroupToResponse(group string, m map[string]model.SystemSettings) (any, error) {
	return nil, nil
}

func (s *stubSettingsService) MapGroupsToResponse(groups []string, maps map[string]map[string]model.SystemSettings) (any, error) {
	return nil, nil
}

func (s *stubSettingsService) UpsertMany(ctx context.Context, group string, upserts []repository.SettingUpsert) error {
	return nil
}

func TestLiveFeatureMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		m          map[string]model.SystemSettings
		loadErr    error
		wantStatus int
	}{
		{
			name:       "live enabled",
			m:          map[string]model.SystemSettings{constant.SettingFeatureLiveEnabled: {SettingValue: "true"}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "live disabled",
			m:          map[string]model.SystemSettings{constant.SettingFeatureLiveEnabled: {SettingValue: "false"}},
			wantStatus: http.StatusNotFound,
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
			engine.Use(mw)
			engine.GET("/live", func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/live", nil)
			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.True(t, called, "next handler should be called")
			} else {
				assert.False(t, called, "next handler should not be called")
			}
		})
	}
}

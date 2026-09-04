package server

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/config"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/ilaziness/orange-tv/internal/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHTTPServer_invalidHandlers(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{Enabled: true, Host: "127.0.0.1", Port: 8080},
	}
	logger := zap.NewNop()

	_, err := NewHTTPServer(cfg, logger, logger, nil, nil, nil)
	require.Error(t, err)
}

func TestNewHTTPServer_registersRoutes(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{Name: "test", Version: "1.0", Env: "test"},
		HTTP: config.HTTPConfig{Enabled: true, Host: "127.0.0.1", Port: 8080},
	}
	logger := zap.NewNop()
	health := httphandler.NewHealthHandler(cfg)
	handlers, err := router.NewHandlers(health)
	require.NoError(t, err)
	// business handlers are mandatory after phase-2 route registration
	b := testutil.NewBusinessHandlers()
	handlers.AuthService = b.AuthService
	handlers.AdminAuth = b.AdminAuth
	handlers.AdminCategory = b.AdminCategory
	handlers.AdminVideo = b.AdminVideo
	handlers.AdminMetadata = b.AdminMetadata
	handlers.AdminPlay = b.AdminPlay
	handlers.AdminLiveTV = b.AdminLiveTV
	handlers.AdminComment = b.AdminComment
	handlers.AdminCollect = b.AdminCollect
	handlers.AdminSettings = b.AdminSettings
	handlers.AdminLog = b.AdminLog
	handlers.AdminMgmt = b.AdminMgmt
	handlers.AdminData = b.AdminData
	handlers.AdminAd = b.AdminAd
	handlers.ClientCategory = b.ClientCategory
	handlers.ClientVideo = b.ClientVideo
	handlers.ClientLiveTV = b.ClientLiveTV
	handlers.ClientSettings = b.ClientSettings
	handlers.ClientUser = b.ClientUser
	handlers.ClientBanner = b.ClientBanner
	handlers.ClientAd = b.ClientAd
	handlers.LiveTVFeature = b.LiveTVFeature
	handlers.OpenResource = b.OpenResource
	handlers.SEO = b.SEO

	srv, err := NewHTTPServer(cfg, logger, logger, handlers, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, srv.Router())

	registered := make(map[string]bool)
	for _, route := range srv.Router().Routes() {
		registered[route.Path] = true
	}
	require.True(t, registered[router.PathHealth])
	require.True(t, registered[router.PathReadiness])
	require.True(t, registered[router.PathLiveness])
	require.True(t, registered[router.PathVersion])
}

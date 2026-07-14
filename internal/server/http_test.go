package server

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/config"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHTTPServer_invalidHandlers(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{Enabled: true, Host: "127.0.0.1", Port: 8080},
	}
	logger := zap.NewNop()

	_, err := NewHTTPServer(cfg, logger, nil, nil, nil)
	require.Error(t, err)
}

func TestNewHTTPServer_registersRoutes(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{Name: "test", Version: "1.0", Env: "test"},
		HTTP: config.HTTPConfig{Enabled: true, Host: "127.0.0.1", Port: 8080},
	}
	logger := zap.NewNop()
	health := httphandler.NewHealthHandler(cfg)
	handlers, err := router.NewHandlers(health, httphandler.NewUserHandler(nil))
	require.NoError(t, err)

	srv, err := NewHTTPServer(cfg, logger, handlers, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, srv.Router())

	registered := make(map[string]bool)
	for _, route := range srv.Router().Routes() {
		registered[route.Path] = true
	}
	require.True(t, registered[router.PathAdminV1Users+"/:id"])
	require.True(t, registered[router.PathInternalV1Users+"/:id"])
}

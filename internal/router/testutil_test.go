package router

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/config"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/testutil"
	"github.com/stretchr/testify/require"
)

func testHandlers(t *testing.T) *Handlers {
	t.Helper()
	cfg := &config.Config{App: config.AppConfig{Name: "test", Version: "1.0"}}
	h, err := NewHandlers(httphandler.NewHealthHandler(cfg))
	require.NoError(t, err)
	applyBusinessHandlers(h, testutil.NewBusinessHandlers())
	return h
}

func applyBusinessHandlers(h *Handlers, b testutil.BusinessHandlers) {
	h.AuthService = b.AuthService
	h.AdminAuth = b.AdminAuth
	h.AdminCategory = b.AdminCategory
	h.AdminVideo = b.AdminVideo
	h.AdminMetadata = b.AdminMetadata
	h.AdminPlay = b.AdminPlay
	h.ClientCategory = b.ClientCategory
	h.ClientVideo = b.ClientVideo
}

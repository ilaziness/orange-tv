package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/stretchr/testify/require"
)

func testHandlers(t *testing.T) *Handlers {
	cfg := &config.Config{App: config.AppConfig{Name: "test", Version: "1.0"}}
	h, err := NewHandlers(httphandler.NewHealthHandler(cfg), httphandler.NewUserHandler(nil))
	require.NoError(t, err)
	return h
}

func TestNewHandlers_requiresDependencies(t *testing.T) {
	cfg := &config.Config{App: config.AppConfig{Name: "test", Version: "1.0"}}
	health := httphandler.NewHealthHandler(cfg)

	_, err := NewHandlers(nil, httphandler.NewUserHandler(nil))
	require.Error(t, err)

	_, err = NewHandlers(health, nil)
	require.Error(t, err)
}

func TestRegisterRoutes_nilHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	err := RegisterRoutes(engine, nil)
	require.Error(t, err)
}

func TestRegisterRoutes_registersSystemHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	require.NoError(t, RegisterRoutes(engine, testHandlers(t)))

	for _, path := range []string{PathHealth, PathReadiness, PathLiveness, PathVersion} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, "path %s", path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterRoutes_registersClientAdminInternalUserRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	require.NoError(t, RegisterRoutes(engine, testHandlers(t)))

	registered := make(map[string]map[string]bool)
	for _, route := range engine.Routes() {
		if registered[route.Path] == nil {
			registered[route.Path] = make(map[string]bool)
		}
		registered[route.Path][route.Method] = true
	}

	expected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, PathClientV1Users + "/:id"},
		{http.MethodPost, PathClientV1Users},
		{http.MethodGet, PathClientV1Users},
		{http.MethodGet, PathClientV2Users + "/:id"},
		{http.MethodGet, PathAdminV1Users + "/:id"},
		{http.MethodPost, PathAdminV1Users},
		{http.MethodGet, PathInternalV1Users + "/:id"},
		{http.MethodGet, PathSwagger},
	}
	for _, route := range expected {
		methods, ok := registered[route.path]
		require.True(t, ok, "path %s should be registered", route.path)
		require.True(t, methods[route.method], "method %s %s should be registered", route.method, route.path)
	}
}

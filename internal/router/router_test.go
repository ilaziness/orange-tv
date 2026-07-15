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
	h, err := NewHandlers(httphandler.NewHealthHandler(cfg))
	require.NoError(t, err)
	return h
}

func TestNewHandlers_requiresDependencies(t *testing.T) {
	_, err := NewHandlers(nil)
	require.Error(t, err)

	cfg := &config.Config{App: config.AppConfig{Name: "test", Version: "1.0"}}
	health := httphandler.NewHealthHandler(cfg)
	h, err := NewHandlers(health)
	require.NoError(t, err)
	require.NotNil(t, h)
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

func TestRegisterRoutes_registersSwaggerAndClientScaffolds(t *testing.T) {
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
	require.True(t, registered[PathSwagger][http.MethodGet])
	require.True(t, registered[PathClientV1Categories][http.MethodGet])
	require.True(t, registered[PathClientV1Videos][http.MethodGet])
	require.True(t, registered[PathClientV1ThemeCurrent][http.MethodGet])
	require.True(t, registered[PathAdminV1Auth+"/login"][http.MethodPost])
	require.True(t, registered[PathAdminV1Auth+"/logout"][http.MethodPost])
	require.True(t, registered[PathAdminV1Auth+"/profile"][http.MethodGet])
	require.True(t, registered[PathAdminV1Categories][http.MethodGet])
}

func TestDefaultJWTSkipPaths_coversClientAndLogin(t *testing.T) {
	paths := DefaultJWTSkipPaths()
	require.Contains(t, paths, PathClientV1+"/*")
	require.Contains(t, paths, PathAdminV1Auth+"/login")
	require.Contains(t, paths, PathHealth)
}

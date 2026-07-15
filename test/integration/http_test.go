package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database/testutil"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationHealthHandler(t *testing.T, cfg *config.Config) *httphandler.HealthHandler {
	t.Helper()
	healthHandler := httphandler.NewHealthHandler(cfg)
	healthHandler.AddChecker(health.NewDatabaseChecker(testutil.OpenBunDB(t)))
	return healthHandler
}

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "orange-tv",
			Version: "1.0.0",
			Env:     constant.EnvTest,
		},
	}

	healthHandler := newIntegrationHealthHandler(t, cfg)
	handlers, err := router.NewHandlers(healthHandler)
	require.NoError(t, err)

	engine := gin.New()
	require.NoError(t, router.RegisterRoutes(engine, handlers))
	return engine
}

func TestHealthEndpoints(t *testing.T) {
	engine := setupTestRouter(t)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)

	for _, path := range []string{"/health", "/liveness", "/readiness", "/version"} {
		t.Run(path, func(t *testing.T) {
			resp, err := http.Get(server.URL + path)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]any
			require.NoError(t, json.Unmarshal(body, &result))
			assert.Contains(t, result, "code")
			assert.Contains(t, result, "message")
		})
	}
}

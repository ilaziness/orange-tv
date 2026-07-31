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
	dbtestutil "github.com/ilaziness/orange-tv/internal/database/testutil"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/ilaziness/orange-tv/internal/router"
	apptestutil "github.com/ilaziness/orange-tv/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationHealthHandler(t *testing.T, cfg *config.Config) *httphandler.HealthHandler {
	t.Helper()
	healthHandler := httphandler.NewHealthHandler(cfg)
	healthHandler.AddChecker(health.NewDatabaseChecker(dbtestutil.OpenBunDB(t)))
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

	// Integration tests only cover system endpoints; business handlers still required by router.
	b := apptestutil.NewBusinessHandlers()
	handlers.AuthService = b.AuthService
	handlers.AdminAuth = b.AdminAuth
	handlers.AdminCategory = b.AdminCategory
	handlers.AdminVideo = b.AdminVideo
	handlers.AdminMetadata = b.AdminMetadata
	handlers.AdminPlay = b.AdminPlay
	handlers.AdminLive = b.AdminLive
	handlers.AdminComment = b.AdminComment
	handlers.AdminCollect = b.AdminCollect
	handlers.AdminSettings = b.AdminSettings
	handlers.AdminLog = b.AdminLog
	handlers.AdminMgmt = b.AdminMgmt
	handlers.AdminData = b.AdminData
	handlers.ClientCategory = b.ClientCategory
	handlers.ClientVideo = b.ClientVideo
	handlers.ClientLive = b.ClientLive
	handlers.ClientSettings = b.ClientSettings
	handlers.ClientUser = b.ClientUser
	handlers.ClientBanner = b.ClientBanner
	handlers.OpenResource = b.OpenResource

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

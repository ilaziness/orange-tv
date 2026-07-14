package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/stretchr/testify/require"
)

type stubHealthChecker struct {
	name   string
	ready  bool
	called int
}

func (s *stubHealthChecker) Name() string { return s.name }

func (s *stubHealthChecker) Check(ctx context.Context) error {
	s.called++
	if s.ready {
		return nil
	}
	return context.Canceled
}

func TestHealthHandler_Readiness_allChecksPass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler(&config.Config{
		App: config.AppConfig{Name: "test", Version: "1.0"},
	})
	db := &stubHealthChecker{name: "database", ready: true}
	h.AddChecker(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readiness", nil)

	h.Readiness(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, float64(0), body["code"])
	require.Equal(t, 1, db.called)

	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "ready", data["status"])
}

func TestHealthHandler_Readiness_failingCheckerReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler(&config.Config{
		App: config.AppConfig{Name: "test", Version: "1.0"},
	})
	h.AddChecker(&stubHealthChecker{name: "database", ready: false})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readiness", nil)

	h.Readiness(c)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotEqual(t, float64(0), body["code"])

	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "not_ready", data["status"])

	checks, ok := data["checks"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, false, checks["database"])
}

func TestHealthHandler_Health_delegatesToLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler(&config.Config{
		App: config.AppConfig{Name: "test", Version: "1.0"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	h.Health(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "alive", data["status"])
}

func TestHealthHandler_Liveness_returnsAlive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler(&config.Config{
		App: config.AppConfig{Name: "test", Version: "1.0"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/liveness", nil)

	h.Liveness(c)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_Version_returnsAppInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler(&config.Config{
		App: config.AppConfig{Name: "test-app", Version: "2.0.0"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/version", nil)

	h.Version(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "test-app", data["name"])
	require.Equal(t, "2.0.0", data["version"])
}

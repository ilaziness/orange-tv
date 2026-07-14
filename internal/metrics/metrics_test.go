package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics_Disabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: false},
	}
	m := NewMetrics(cfg)
	assert.NotNil(t, m)
	assert.Nil(t, m.Registry())
}

func TestNewMetrics_Enabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
	m := NewMetrics(cfg)
	assert.NotNil(t, m)
	assert.NotNil(t, m.Registry())

	// Verify metrics are initialized
	assert.NotNil(t, m.HTTPRequestsTotal)
	assert.NotNil(t, m.HTTPRequestDuration)
	assert.NotNil(t, m.HTTPInFlight)
	assert.NotNil(t, m.DBConnectionsActive)
	assert.NotNil(t, m.DBConnectionsIdle)
	assert.NotNil(t, m.DBQueryDuration)
	assert.NotNil(t, m.RedisCacheHits)
	assert.NotNil(t, m.RedisCacheMisses)
	assert.NotNil(t, m.RedisOperationDuration)
}

func TestHandler_Disabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: false},
	}
	m := NewMetrics(cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	Handler(m, "/metrics")(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandler_Enabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
	m := NewMetrics(cfg)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", Handler(m, "/metrics"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_Disabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: false},
	}
	m := NewMetrics(cfg)

	middleware := Middleware(m)
	assert.NotNil(t, middleware)
}

func TestMiddleware_Enabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
	m := NewMetrics(cfg)

	middleware := Middleware(m)
	assert.NotNil(t, middleware)

	// Test middleware execution
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateDBConnectionStats_Disabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: false},
	}
	m := NewMetrics(cfg)

	// Should not panic
	m.UpdateDBConnectionStats(10, 5)
}

func TestUpdateDBConnectionStats_Enabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: true},
	}
	m := NewMetrics(cfg)

	m.UpdateDBConnectionStats(10, 5)
	// Metrics should be updated without error
}

func TestRecordDBQuery_Disabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: false},
	}
	m := NewMetrics(cfg)

	// Should not panic
	m.RecordDBQuery("select", "users", 0.1)
}

func TestRecordDBQuery_Enabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: true},
	}
	m := NewMetrics(cfg)

	m.RecordDBQuery("select", "users", 0.1)
	// Metrics should be recorded without error
}

func TestRecordRedisCacheHit_Disabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: false},
	}
	m := NewMetrics(cfg)

	// Should not panic
	m.RecordRedisCacheHit("user:")
}

func TestRecordRedisCacheHit_Enabled(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: true},
	}
	m := NewMetrics(cfg)

	m.RecordRedisCacheHit("user:")
	// Metrics should be recorded without error
}

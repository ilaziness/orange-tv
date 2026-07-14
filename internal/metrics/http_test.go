package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/router"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_SkipsSystemPaths(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: true, Path: "/metrics"},
	}
	m := NewMetrics(cfg)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Middleware(m, router.SystemPaths...))
	engine.GET(router.PathHealth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	engine.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, router.PathHealth, nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/test", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	metricFamilies, err := gatherMetrics(m)
	require.NoError(t, err)

	var healthCount, apiCount float64
	for _, mf := range metricFamilies {
		if mf.GetName() != "http_requests_total" {
			continue
		}
		for _, metric := range mf.GetMetric() {
			labels := metric.GetLabel()
			path := labelValue(labels, "path")
			switch path {
			case router.PathHealth:
				healthCount += metric.GetCounter().GetValue()
			case "/api/test":
				apiCount += metric.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(0), healthCount, "health probe should not be counted")
	assert.Equal(t, float64(1), apiCount, "business route should be counted")
}

func gatherMetrics(m *Metrics) ([]*dto.MetricFamily, error) {
	reg := m.Registry()
	if reg == nil {
		return nil, nil
	}
	gatherer := reg
	return gatherer.Gather()
}

func labelValue(labels []*dto.LabelPair, name string) string {
	for _, l := range labels {
		if l.GetName() == name {
			return l.GetValue()
		}
	}
	return ""
}

func TestHandler_metricsOutputIncludesRuntime(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{Enabled: true, Path: "/metrics"},
	}
	m := NewMetrics(cfg)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", Handler(m, "/metrics"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(string(body)), "go_goroutines")
}

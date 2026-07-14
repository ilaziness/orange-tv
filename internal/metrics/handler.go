// Package metrics provides Prometheus metrics collection.
package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns a Gin handler for serving Prometheus metrics.
func Handler(m *Metrics, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.registry == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics disabled"})
			return
		}
		handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

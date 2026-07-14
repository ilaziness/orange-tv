// Package metrics provides Prometheus metrics collection.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware for collecting HTTP metrics.
// Paths in skipPaths are excluded from request metrics (e.g. health probes).
func Middleware(m *Metrics, skipPaths ...string) gin.HandlerFunc {
	skipSet := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skipSet[p] = struct{}{}
	}

	return func(c *gin.Context) {
		if m.registry == nil {
			c.Next()
			return
		}

		if _, skip := skipSet[c.FullPath()]; skip {
			c.Next()
			return
		}

		start := time.Now()
		m.HTTPInFlight.Inc()

		c.Next()

		duration := time.Since(start).Seconds()
		m.HTTPInFlight.Dec()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

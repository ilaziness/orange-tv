package http

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// InjectTraceID injects the trace ID into the response header.
// This middleware should only be used when tracing is enabled.
func InjectTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span == nil {
			c.Next()
			return
		}
		spanContext := span.SpanContext()
		traceID := spanContext.TraceID().String()
		if traceID != "" {
			c.Header("X-Trace-ID", traceID)
		}
		c.Next()
	}
}

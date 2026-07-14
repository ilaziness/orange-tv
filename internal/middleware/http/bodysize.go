// Package http provides HTTP middleware implementations.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const defaultMaxBodySize = 10 << 20 // 10 MB

// BodySizeLimit returns a middleware that limits the maximum size of the request body.
// maxBytes specifies the maximum allowed size in bytes; if <= 0, defaults to 10 MB.
func BodySizeLimit(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBodySize
	}
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

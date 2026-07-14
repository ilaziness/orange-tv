// Package http provides HTTP middleware implementations.
package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Deprecated returns a middleware that marks an API route as deprecated by
// attaching the following response headers:
//
//   - Deprecation: true
//   - Sunset: <sunsetDate formatted as RFC3339>
//   - Link: <successorPath>; rel="successor-version"
//
// sunsetDate is the planned removal date. successorPath is the URL of the
// replacement API (e.g. "/api/v2/users").
func Deprecated(sunsetDate time.Time, successorPath string) gin.HandlerFunc {
	sunsetStr := sunsetDate.UTC().Format(time.RFC3339)
	linkHeader := fmt.Sprintf(`<%s>; rel="successor-version"`, successorPath)
	return func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Sunset", sunsetStr)
		c.Header("Link", linkHeader)
		c.Next()
	}
}

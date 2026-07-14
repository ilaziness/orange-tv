package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout returns a middleware that sets a timeout for the request.
// If the request takes longer than the specified duration, it returns 504 Gateway Timeout.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use a channel to signal when the handler is done
		done := make(chan struct{})

		// Run the next handlers in a goroutine
		go func() {
			defer close(done)
			c.Next()
		}()

		// Wait for either the handler to finish or timeout
		select {
		case <-done:
			// Handler completed within timeout
			return
		case <-time.After(timeout):
			// Timeout occurred
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"code":    10005,
				"message": "request timeout",
			})
			return
		}
	}
}

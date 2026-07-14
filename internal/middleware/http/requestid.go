package http

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestID returns a middleware that generates a unique request ID.
func RequestID() gin.HandlerFunc {
	return requestid.New()
}

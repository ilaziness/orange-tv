// Package http provides HTTP middleware implementations.
package http

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Compress returns a Gzip compression middleware using the default compression level.
// The live stream proxy path is excluded from compression because gzip-encoding
// binary segments (.ts/.flv) wastes CPU and can break Content-Length based playback.
func Compress() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedPathsRegexs([]string{`/api/client/v\d+/live/play/`}),
	)
}

// CompressWithLevel returns a Gzip compression middleware with the specified level.
// Valid levels: gzip.BestSpeed (1), gzip.DefaultCompression (-1), gzip.BestCompression (9).
func CompressWithLevel(level int) gin.HandlerFunc {
	return gzip.Gzip(level,
		gzip.WithExcludedPathsRegexs([]string{`/api/client/v\d+/live/play/`}),
	)
}

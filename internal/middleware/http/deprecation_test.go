package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeprecated_ResponseHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sunsetDate := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)

	r := gin.New()
	r.Use(Deprecated(sunsetDate, "/api/v2/users"))
	r.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("Deprecation"))
	assert.NotEmpty(t, w.Header().Get("Sunset"))
	assert.Contains(t, w.Header().Get("Link"), "/api/v2/users")
	assert.Contains(t, w.Header().Get("Link"), `rel="successor-version"`)
}

func TestDeprecated_SunsetDateFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sunsetDate := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	r := gin.New()
	r.Use(Deprecated(sunsetDate, "/api/v2/resource"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, "2027-01-01T00:00:00Z", w.Header().Get("Sunset"))
}

func TestNonDeprecatedRoute_NoHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/v2/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v2/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Deprecation"))
	assert.Empty(t, w.Header().Get("Sunset"))
	assert.Empty(t, w.Header().Get("Link"))
}

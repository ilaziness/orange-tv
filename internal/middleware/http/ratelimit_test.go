package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupRateLimitRouter(cfg RateLimitConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	r := gin.New()
	r.Use(RateLimit(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRateLimit_Disabled(t *testing.T) {
	cfg := RateLimitConfig{
		Enabled: false,
		Store:   NewMemoryRateLimitStore(10000),
	}
	r := setupRateLimitRouter(cfg)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimit_GlobalPassThrough(t *testing.T) {
	cfg := DefaultRateLimitConfig(NewMemoryRateLimitStore(10000))
	cfg.GlobalRPS = 1000
	cfg.IPRPS = 0
	cfg.UserRPS = 0
	r := setupRateLimitRouter(cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Rate limit headers are only set when the limit is exceeded (429).
	assert.Empty(t, w.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimit_IPThrottle(t *testing.T) {
	store := NewMemoryRateLimitStore(10000)
	cfg := RateLimitConfig{
		Enabled:   true,
		GlobalRPS: 0,
		IPRPS:     1, // allow only 1 req/s
		UserRPS:   0,
		Store:     store,
	}
	r := setupRateLimitRouter(cfg)

	// First request should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Rapid second request from same IP should be throttled
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Equal(t, "0", w2.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimit_GlobalThrottle(t *testing.T) {
	store := NewMemoryRateLimitStore(10000)
	cfg := RateLimitConfig{
		Enabled:   true,
		GlobalRPS: 1, // allow only 1 req/s globally
		IPRPS:     0,
		UserRPS:   0,
		Store:     store,
	}
	r := setupRateLimitRouter(cfg)

	// First request passes
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Immediate second request is throttled
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimit_SkipsSystemPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	store := NewMemoryRateLimitStore(10000)

	r := gin.New()
	r.Use(RateLimit(RateLimitConfig{
		Enabled:   true,
		GlobalRPS: 1,
		IPRPS:     0,
		UserRPS:   0,
		Store:     store,
		SkipPaths: []string{"/health"},
	}, logger))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Health probe should bypass global rate limit even when exhausted.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

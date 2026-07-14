// Package http provides HTTP middleware implementations.
package http

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	redisrate "github.com/go-redis/redis_rate/v10"
	lru "github.com/hashicorp/golang-lru/v2"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/response"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimitStore defines the backend storage interface for rate limiting.
type RateLimitStore interface {
	// Allow checks whether the given key is allowed under the specified limit (req/s).
	// Returns (allowed bool, remaining int, resetAfter time.Duration).
	Allow(ctx context.Context, key string, rps int) (bool, int, time.Duration, error)
}

// MemoryRateLimitStore implements RateLimitStore using in-memory token buckets with LRU eviction.
type MemoryRateLimitStore struct {
	mu    sync.RWMutex
	cache *lru.Cache[string, *rate.Limiter]
}

// NewMemoryRateLimitStore creates a new in-memory rate limit store with LRU cache.
// maxKeys is the maximum number of unique keys to keep in memory (e.g., 10000).
func NewMemoryRateLimitStore(maxKeys int) *MemoryRateLimitStore {
	if maxKeys <= 0 {
		maxKeys = 10000 // default
	}
	cache, err := lru.New[string, *rate.Limiter](maxKeys)
	if err != nil {
		panic(fmt.Sprintf("failed to create LRU cache: %v", err))
	}
	return &MemoryRateLimitStore{cache: cache}
}

func (s *MemoryRateLimitStore) getLimiter(key string, rps int) *rate.Limiter {
	s.mu.RLock()
	if l, ok := s.cache.Get(key); ok {
		s.mu.RUnlock()
		return l
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Double-check after acquiring write lock
	if l, ok := s.cache.Get(key); ok {
		return l
	}
	l := rate.NewLimiter(rate.Limit(rps), rps)
	s.cache.Add(key, l)
	return l
}

// Allow checks the in-memory rate limit for the given key.
func (s *MemoryRateLimitStore) Allow(ctx context.Context, key string, rps int) (bool, int, time.Duration, error) {
	l := s.getLimiter(key, rps)
	reservation := l.Reserve()
	if !reservation.OK() {
		return false, 0, reservation.Delay(), nil
	}
	if reservation.Delay() > 0 {
		reservation.Cancel()
		return false, 0, reservation.Delay(), nil
	}
	remaining := int(l.Tokens())
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, 0, nil
}

// RedisRateLimitStore implements RateLimitStore using Redis GCRA via redis_rate.
type RedisRateLimitStore struct {
	limiter *redisrate.Limiter
}

// NewRedisRateLimitStore creates a new Redis-backed rate limit store.
func NewRedisRateLimitStore(limiter *redisrate.Limiter) *RedisRateLimitStore {
	return &RedisRateLimitStore{limiter: limiter}
}

// Allow checks the Redis-backed rate limit for the given key.
func (s *RedisRateLimitStore) Allow(ctx context.Context, key string, rps int) (bool, int, time.Duration, error) {
	res, err := s.limiter.Allow(ctx, key, redisrate.PerSecond(rps))
	if err != nil {
		return true, 0, 0, err
	}
	return res.Allowed > 0, res.Remaining, res.ResetAfter, nil
}

// RateLimitConfig holds configuration for the rate limit middleware.
type RateLimitConfig struct {
	Enabled bool
	// GlobalRPS is the global rate limit in requests per second (0 = disabled).
	GlobalRPS int
	// IPRPS is the per-IP rate limit in requests per second (0 = disabled).
	IPRPS int
	// UserRPS is the per-user rate limit in requests per second (0 = disabled).
	// Requires JWT middleware to run before this middleware.
	UserRPS int
	Store   RateLimitStore
	// SkipPaths are route full paths excluded from rate limiting (e.g. health probes).
	SkipPaths []string
}

// DefaultRateLimitConfig returns a RateLimitConfig with sensible defaults.
func DefaultRateLimitConfig(store RateLimitStore) RateLimitConfig {
	return RateLimitConfig{
		Enabled:   true,
		GlobalRPS: 10000,
		IPRPS:     100,
		UserRPS:   50,
		Store:     store,
	}
}

// RateLimit returns a layered rate limit middleware.
// Layers are checked in order: global → IP → user (if JWT claims present).
// Any layer that triggers returns HTTP 429 with X-RateLimit-* headers.
func RateLimit(cfg RateLimitConfig, logger *zap.Logger) gin.HandlerFunc {
	skipSet := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipSet[p] = struct{}{}
	}

	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		if _, skip := skipSet[c.FullPath()]; skip {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		// Layer 1: Global rate limit
		if cfg.GlobalRPS > 0 {
			if ok, remaining, reset := checkLimit(ctx, cfg.Store, "global", cfg.GlobalRPS, logger); !ok {
				setRateLimitHeaders(c, cfg.GlobalRPS, remaining, reset)
				response.Error(c, errcode.TooManyRequests)
				c.Abort()
				return
			}
		}

		// Layer 2: IP rate limit
		if cfg.IPRPS > 0 {
			ip := c.ClientIP()
			key := fmt.Sprintf("ip:%s", ip)
			if ok, remaining, reset := checkLimit(ctx, cfg.Store, key, cfg.IPRPS, logger); !ok {
				setRateLimitHeaders(c, cfg.IPRPS, remaining, reset)
				response.Error(c, errcode.TooManyRequests)
				c.Abort()
				return
			}
		}

		// Layer 3: User rate limit (requires JWT middleware upstream)
		if cfg.UserRPS > 0 {
			if claims := GetClaims(c); claims != nil {
				key := fmt.Sprintf("user:%d", claims.UserID)
				if ok, remaining, reset := checkLimit(ctx, cfg.Store, key, cfg.UserRPS, logger); !ok {
					setRateLimitHeaders(c, cfg.UserRPS, remaining, reset)
					response.Error(c, errcode.TooManyRequests)
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// RouteRateLimit returns a per-route rate limit middleware with the given rps.
func RouteRateLimit(store RateLimitStore, rps int, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("route:%s:%s", c.Request.Method, c.FullPath())
		if ok, remaining, reset := checkLimit(ctx, store, key, rps, logger); !ok {
			setRateLimitHeaders(c, rps, remaining, reset)
			response.Error(c, errcode.TooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}

func checkLimit(ctx context.Context, store RateLimitStore, key string, rps int, logger *zap.Logger) (bool, int, time.Duration) {
	allowed, remaining, resetAfter, err := store.Allow(ctx, key, rps)
	if err != nil {
		// Log error but fail-open (allow request) to avoid service disruption
		// Consider making this configurable (fail-open vs fail-closed)
		logger.Warn("rate limit store error, allowing request",
			zap.String("key", key),
			zap.Error(err))
		return true, 0, 0
	}
	return allowed, remaining, resetAfter
}

func setRateLimitHeaders(c *gin.Context, limit, remaining int, resetAfter time.Duration) {
	resetTime := time.Now().Add(resetAfter).Unix()
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
	if remaining == 0 {
		c.Header("Retry-After", strconv.FormatInt(int64(resetAfter.Seconds()+1), 10))
	}
}

package server

import (
	"context"
	"fmt"
	"net/http"
	stdhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	redisrate "github.com/go-redis/redis_rate/v10"
	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/config"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/metrics"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	"github.com/ilaziness/orange-tv/internal/router"
	"github.com/ilaziness/orange-tv/internal/tracing"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HTTPServer struct {
	server   *stdhttp.Server
	router   *gin.Engine
	addr     string
	enabled  bool
	tls      bool
	certFile string
	keyFile  string
}

// NewHTTPServer creates the HTTP server. accessLogger is used for HTTP access
// logs (gin request logs) and should write to stdout only so that access logs
// are not persisted to the application log file. logger is used for
// non-access logging (recovery, rate limit warnings, etc.).
func NewHTTPServer(cfg *config.Config, logger *zap.Logger, accessLogger *zap.Logger, h *router.Handlers, m *metrics.Metrics, jwtMgr *auth.JWTManager) (*HTTPServer, error) {
	httpServer := &HTTPServer{
		enabled: cfg.HTTP.Enabled,
	}

	if !cfg.HTTP.Enabled {
		return httpServer, nil
	}

	ginRouter := gin.New()

	// Apply core middlewares
	ginRouter.Use(httpmiddleware.RequestID())
	ginRouter.Use(httpmiddleware.SecurityHeaders())
	// Access logs go to stdout only (accessLogger) to avoid cluttering log files.
	ginRouter.Use(httpmiddleware.Logger(accessLogger, router.SystemPaths...))
	ginRouter.Use(httpmiddleware.Recovery(logger))
	ginRouter.Use(httpmiddleware.CORS())
	ginRouter.Use(httpmiddleware.BodySizeLimit(0)) // default 10MB

	// Register metrics endpoint before gzip middleware (metrics should not be compressed)
	if cfg.Metrics.Enabled {
		ginRouter.GET(cfg.Metrics.Path, metrics.Handler(m, cfg.Metrics.Path))
	}

	ginRouter.Use(httpmiddleware.Compress())

	// Apply tracing middleware if enabled
	if cfg.Tracing.Enabled {
		ginRouter.Use(tracing.Middleware(cfg, router.SystemPaths...))
		ginRouter.Use(httpmiddleware.InjectTraceID())
	}

	// Apply metrics middleware if enabled
	if cfg.Metrics.Enabled {
		ginRouter.Use(metrics.Middleware(m, router.SystemPaths...))
	}

	// Apply JWT middleware if manager is provided
	if jwtMgr != nil {
		skipPaths := cfg.JWT.SkipPaths
		if len(skipPaths) == 0 {
			skipPaths = router.DefaultJWTSkipPaths()
		}
		ginRouter.Use(httpmiddleware.JWTAuth(jwtMgr, skipPaths...))
	}

	// Apply rate limit middleware if enabled (must be after JWT for per-user rate limiting)
	if cfg.RateLimit.Enabled {
		var store httpmiddleware.RateLimitStore
		if cfg.RateLimit.Store == "redis" && cfg.Redis.Enabled {
			redisClient := redis.NewClient(&redis.Options{
				Addr:            fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
				Password:        cfg.Redis.Password,
				DB:              cfg.Redis.DB,
				PoolSize:        cfg.Redis.PoolSize,
				MinIdleConns:    cfg.Redis.MinIdleConns,
				ConnMaxIdleTime: time.Duration(cfg.Redis.IdleTimeout) * time.Second,
			})
			limiter := redisrate.NewLimiter(redisClient)
			store = httpmiddleware.NewRedisRateLimitStore(limiter)
			logger.Info("rate limit: using Redis store")
		} else {
			if cfg.RateLimit.Store == "redis" {
				logger.Warn("rate limit: redis store requested but redis not enabled, falling back to memory")
			}
			store = httpmiddleware.NewMemoryRateLimitStore(10000)
		}
		rateCfg := httpmiddleware.RateLimitConfig{
			Enabled:   true,
			GlobalRPS: cfg.RateLimit.GlobalRPS,
			IPRPS:     cfg.RateLimit.IPRPS,
			UserRPS:   cfg.RateLimit.UserRPS,
			Store:     store,
			SkipPaths: router.SystemPaths,
		}
		ginRouter.Use(httpmiddleware.RateLimit(rateCfg, logger))
	}

	// Always-on login rate limit (IP-based, independent of global rate_limit config).
	loginStore := httpmiddleware.NewMemoryRateLimitStore(10000)
	ginRouter.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && (c.Request.URL.Path == router.PathAdminV1+"/auth/login" || c.Request.URL.Path == router.PathClientV1+"/auth/login") {
			ok, _, _, err := loginStore.Allow(c.Request.Context(), "login:"+c.ClientIP(), 5)
			if err == nil && !ok {
				response.Error(c, errcode.TooManyRequests)
				c.Abort()
				return
			}
		}
		c.Next()
	})

	// Register routes
	if err := router.RegisterRoutes(ginRouter, h); err != nil {
		return nil, fmt.Errorf("register routes: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	httpServer.router = ginRouter
	httpServer.addr = addr
	httpServer.server = &stdhttp.Server{
		Addr:              addr,
		Handler:           ginRouter,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if cfg.HTTP.TLS.Enabled {
		httpServer.tls = true
		httpServer.certFile = cfg.HTTP.TLS.CertFile
		httpServer.keyFile = cfg.HTTP.TLS.KeyFile
	}

	return httpServer, nil
}

func (s *HTTPServer) Start() error {
	if !s.enabled {
		return nil
	}
	if s.tls {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.server.ListenAndServe()
}

// Enabled returns whether the HTTP server is enabled.
func (s *HTTPServer) Enabled() bool {
	return s.enabled
}

// Serve starts the HTTP server (alias for Start, used by app lifecycle).
func (s *HTTPServer) Serve() error {
	return s.Start()
}

// Stop gracefully shuts down the HTTP server.
func (s *HTTPServer) Stop(ctx context.Context) error {
	if !s.enabled || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Shutdown gracefully shuts down the HTTP server (alias for Stop).
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.Stop(ctx)
}

func (s *HTTPServer) Addr() string {
	return s.addr
}

// Router returns the underlying gin engine (for testing).
func (s *HTTPServer) Router() *gin.Engine {
	return s.router
}

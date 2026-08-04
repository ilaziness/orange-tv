package app

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/auth"
	internalcache "github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/event"
	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/ilaziness/orange-tv/internal/metrics"
	"github.com/ilaziness/orange-tv/internal/tracing"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
	pkgevent "github.com/ilaziness/orange-tv/pkg/event"
	"go.uber.org/zap"
)

func (a *App) wireInfra() error {
	err := logger.Init(logger.Config{
		Level:      a.cfg.Log.Level,
		Output:     a.cfg.Log.Output,
		Filename:   a.cfg.Log.Filename,
		MaxSize:    a.cfg.Log.MaxSize,
		MaxBackups: a.cfg.Log.MaxBackups,
		MaxAge:     a.cfg.Log.MaxAge,
		Compress:   a.cfg.Log.Compress,
	})
	if err != nil {
		return err
	}
	logInst := logger.Log
	a.logger = logInst
	a.log = logInst.Logger

	db, err := newDatabase(a.cfg, a.log)
	if err != nil {
		return err
	}
	a.db = db
	a.addHook(Hook{
		Name:   "database",
		OnStop: func(ctx context.Context) error { return a.db.Close() },
	})

	a.jwtMgr = auth.NewJWTManagerFromConfig(a.cfg)
	if a.jwtMgr != nil && a.cfg.JWT.Secret == "" {
		a.log.Warn("jwt.secret is not configured, using default development secret")
	}

	cacheFactory := pkgcache.NewCacheFactory(pkgcache.CacheFactoryOptions{
		Enabled:    a.cfg.Cache.Enabled,
		Driver:     a.cfg.Cache.Driver,
		Memory:     a.cfg.Cache.Memory,
		RedisCache: a.cfg.Cache.Redis,
		RedisConn: pkgcache.RedisOptions{
			Host:               a.cfg.Redis.Host,
			Port:               a.cfg.Redis.Port,
			Password:           a.cfg.Redis.Password,
			DB:                 a.cfg.Redis.DB,
			PoolSize:           a.cfg.Redis.PoolSize,
			MinIdleConns:       a.cfg.Redis.MinIdleConns,
			IdleTimeout:        a.cfg.Redis.IdleTimeout,
			IdleCheckFrequency: a.cfg.Redis.IdleCheckFrequency,
		},
	})
	c, err := cacheFactory.Create()
	if err != nil {
		return err
	}
	a.cache = internalcache.NewManager(c)
	a.addHook(Hook{
		Name:   "cache",
		OnStop: func(ctx context.Context) error { return a.cache.Close() },
	})

	bus := pkgevent.NewEventBusWithLogger(a.log)
	pkgevent.SetDefault(bus)
	a.registerBuiltinEventListeners()
	a.addHook(Hook{
		Name: "event_bus",
		OnStop: func(ctx context.Context) error {
			err := pkgevent.Default().Close()
			pkgevent.SetDefault(nil)
			return err
		},
	})

	tracer, err := tracing.NewTracer(a.cfg)
	if err != nil {
		return err
	}
	a.tracer = tracer
	a.addHook(Hook{
		Name:   "tracer",
		OnStop: func(ctx context.Context) error { return a.tracer.Shutdown(ctx) },
	})

	a.metrics = metrics.NewMetrics(a.cfg)

	return nil
}

func (a *App) registerBuiltinEventListeners() {
	if err := pkgevent.Subscribe(event.EventAppStopped, func(ctx context.Context, ev *event.Event) error {
		if payload, ok := ev.Payload.(*event.AppStoppedPayload); ok {
			a.log.Info("Application stopped",
				zap.Duration("uptime", payload.Uptime),
				zap.Time("stop_time", payload.StopTime),
			)
		}
		return nil
	}); err != nil && a.log != nil {
		a.log.Warn("Failed to subscribe to app stopped event", zap.Error(err))
	}
}

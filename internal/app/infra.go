package app

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/event"
	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/ilaziness/orange-tv/internal/metrics"
	"github.com/ilaziness/orange-tv/internal/tracing"
	"go.uber.org/zap"
)

func (a *App) wireInfra() error {
	err := logger.Init(logger.Config{
		Level:      a.cfg.Log.Level,
		Format:     a.cfg.Log.Format,
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

	cacheFactory := cache.NewCacheFactory(a.cfg)
	c, err := cacheFactory.Create()
	if err != nil {
		return err
	}
	a.cache = c
	a.addHook(Hook{
		Name:   "cache",
		OnStop: func(ctx context.Context) error { return a.cache.Close() },
	})

	a.eventBus = event.NewEventBusWithLogger(a.log)
	a.registerBuiltinEventListeners()
	a.addHook(Hook{
		Name:   "event_bus",
		OnStop: func(ctx context.Context) error { return a.eventBus.Close() },
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
	if err := a.eventBus.Subscribe(event.EventAppStopped, func(ctx context.Context, ev *event.Event) error {
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

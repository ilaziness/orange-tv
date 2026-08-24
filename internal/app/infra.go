package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/auth"
	internalcache "github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/event"
	"github.com/ilaziness/orange-tv/internal/logger"
	"github.com/ilaziness/orange-tv/internal/metrics"
	"github.com/ilaziness/orange-tv/internal/tracing"
	"github.com/ilaziness/orange-tv/internal/utils"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
	pkgevent "github.com/ilaziness/orange-tv/pkg/event"
	pkglock "github.com/ilaziness/orange-tv/pkg/lock"
	"github.com/redis/go-redis/v9"
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

	// 为 utils.Go 的 panic recovery 设置 logger，确保 goroutine panic 能被记录。
	utils.SetLogger(a.log)

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
	a.pkgCache = c
	a.addHook(Hook{
		Name:   "cache",
		OnStop: func(ctx context.Context) error { return a.cache.Close() },
	})

	// 独立的 Redis 客户端，用于调度器分布式锁等非缓存用途。
	// 仅在 Redis 启用时创建，与 cache 内部 client 分离，职责清晰。
	if a.cfg.Redis.Enabled {
		redisClient, redisErr := newRedisClient(a.cfg)
		if redisErr != nil {
			return redisErr
		}
		a.redisClient = redisClient
		a.addHook(Hook{
			Name:   "redis_client",
			OnStop: func(ctx context.Context) error { return a.redisClient.Close() },
		})
	}

	// 通用分布式锁容器：Redis 启用时使用 Redis 实现，未启用时降级为进程内内存锁。
	// 业务方通过 internal/lock 包生成 key 后调用此 locker 做并发去重（如注册接口的邮箱锁）。
	lockerFactory := pkglock.NewLockerFactory(pkglock.LockerFactoryOptions{
		RedisClient: a.redisClient,
	})
	a.locker = lockerFactory.Create()

	bus := pkgevent.NewEventBusWithLogger(a.log)
	pkgevent.SetDefault(bus)
	a.registerBuiltinEventListeners()
	a.addHook(Hook{
		Name: "event_bus",
		OnStop: func(ctx context.Context) error {
			closeErr := pkgevent.Default().Close()
			pkgevent.SetDefault(nil)
			return closeErr
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

// newRedisClient 创建独立的 Redis 客户端，用于调度器分布式锁等非缓存用途。
// 与 cache 内部 client 分离，避免职责耦合。
func newRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		ConnMaxIdleTime: time.Duration(cfg.Redis.IdleTimeout) * time.Second,
	})

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to Redis for scheduler lock: %w", err)
	}

	return client, nil
}

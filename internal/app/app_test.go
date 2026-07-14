package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/health"
	"github.com/stretchr/testify/require"
)

func testSQLiteConfig(t *testing.T) *config.Config {
	return &config.Config{
		App: config.AppConfig{Name: "test", Version: "1.0", Env: "test"},
		Log: config.LogConfig{Level: "error", Format: "json", Output: "stdout"},
		Database: config.DatabaseConfig{
			Driver:   "sqlite",
			Database: filepath.Join(t.TempDir(), "test.db"),
		},
		Cache:   config.CacheConfig{Enabled: false},
		HTTP:    config.HTTPConfig{Enabled: false},
		Tracing: config.TracingConfig{Enabled: false},
		Metrics: config.MetricsConfig{Enabled: false},
	}
}

func TestNew_closesDBWhenCacheWireFails(t *testing.T) {
	cfg := testSQLiteConfig(t)
	cfg.Cache.Enabled = true
	cfg.Cache.Driver = "redis"
	cfg.Redis.Host = "127.0.0.1"
	cfg.Redis.Port = 59999

	_, err := New(cfg)
	require.Error(t, err)

	cfg2 := testSQLiteConfig(t)
	app, err := New(cfg2)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NoError(t, app.shutdown())
}

func TestNew_succeedsWithMinimalSQLite(t *testing.T) {
	cfg := testSQLiteConfig(t)

	app, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NoError(t, app.shutdown())
}

func TestShutdown_skipsAppStoppedWhenNotStarted(t *testing.T) {
	cfg := testSQLiteConfig(t)

	app, err := New(cfg)
	require.NoError(t, err)

	app.appStarted = false
	require.NoError(t, app.shutdown())
}

func TestMaxShutdownTimeout(t *testing.T) {
	a := &App{
		cfg: &config.Config{
			HTTP: config.HTTPConfig{Enabled: true, ShutdownTimeout: 60},
		},
	}

	require.Equal(t, 60*time.Second, a.maxShutdownTimeout())
}

func TestDBHealthChecker_usesTimeout(t *testing.T) {
	cfg := testSQLiteConfig(t)
	app, err := New(cfg)
	require.NoError(t, err)
	defer app.shutdown()

	checker := health.NewDatabaseChecker(app.db)
	require.NoError(t, checker.Check(context.Background()))
}

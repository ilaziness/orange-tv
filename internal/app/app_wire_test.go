package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWireHTTP_skipsHTTPServerWhenDisabled(t *testing.T) {
	cfg := testSQLiteConfig(t)
	cfg.HTTP.Enabled = false

	app, err := New(cfg)
	require.NoError(t, err)
	require.Nil(t, app.httpServer)

	require.NoError(t, app.shutdown())
}

func TestWireHTTP_buildsHTTPServerWhenEnabled(t *testing.T) {
	cfg := testSQLiteConfig(t)
	cfg.HTTP.Enabled = true
	cfg.HTTP.Host = "127.0.0.1"
	cfg.HTTP.Port = 18080
	cfg.Metrics.Enabled = false
	cfg.Tracing.Enabled = false
	cfg.JWT.Secret = ""

	app, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, app.httpServer)
	require.True(t, app.httpServer.Enabled())
	require.Equal(t, "127.0.0.1:18080", app.httpServer.Addr())

	require.NoError(t, app.shutdown())
}

package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShutdown_publishesAppStoppedWhenStarted(t *testing.T) {
	cfg := testSQLiteConfig(t)
	app, err := New(cfg)
	require.NoError(t, err)

	app.appStarted = true
	app.startTime = time.Now().Add(-time.Second)

	require.NoError(t, app.shutdown())
}

func TestAddrOf_formatsHostPort(t *testing.T) {
	require.Equal(t, "0.0.0.0:8080", addrOf("0.0.0.0", 8080))
}

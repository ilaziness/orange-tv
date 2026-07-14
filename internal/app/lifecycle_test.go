package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStartHooks_runsOnStartInOrder(t *testing.T) {
	app := &App{log: zap.NewNop()}
	var order []int
	app.addHook(Hook{
		OnStart: func(context.Context) error { order = append(order, 1); return nil },
	})
	app.addHook(Hook{
		OnStop: func(context.Context) error { return nil },
	})
	app.addHook(Hook{
		OnStart: func(context.Context) error { order = append(order, 2); return nil },
	})

	require.NoError(t, app.startHooks(context.Background()))
	require.Equal(t, []int{1, 2}, order)
}

func TestStartHooks_skipsStopOnlyHooks(t *testing.T) {
	app := &App{log: zap.NewNop()}
	called := false
	app.addHook(Hook{
		OnStop: func(context.Context) error { called = true; return nil },
	})

	require.NoError(t, app.startHooks(context.Background()))
	require.False(t, called)
}

func TestStartHooks_rollbackOnFailure(t *testing.T) {
	app := &App{log: zap.NewNop()}
	var stopOrder []int
	app.addHook(Hook{
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { stopOrder = append(stopOrder, 1); return nil },
	})
	app.addHook(Hook{
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { stopOrder = append(stopOrder, 2); return nil },
	})
	app.addHook(Hook{
		OnStart: func(context.Context) error { return errors.New("start failed") },
		OnStop:  func(context.Context) error { stopOrder = append(stopOrder, 3); return nil },
	})

	err := app.startHooks(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "start failed")
	require.Equal(t, []int{2, 1}, stopOrder)
}

func TestStartHooks_joinsRollbackErrors(t *testing.T) {
	app := &App{log: zap.NewNop()}
	app.addHook(Hook{
		Name:    "worker-a",
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { return errors.New("rollback-fail") },
	})
	app.addHook(Hook{
		Name:    "worker-b",
		OnStart: func(context.Context) error { return errors.New("start-fail") },
	})

	err := app.startHooks(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "start-fail")
	require.ErrorContains(t, err, "rollback-fail")
}

func TestStartHooks_failureThenShutdown_doesNotDoubleStop(t *testing.T) {
	app := &App{log: zap.NewNop()}
	stopCount := 0
	app.addHook(Hook{
		Name:    "worker",
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { stopCount++; return nil },
	})
	app.addHook(Hook{
		Name:    "failing-worker",
		OnStart: func(context.Context) error { return errors.New("start failed") },
		OnStop:  func(context.Context) error { stopCount++; return nil },
	})

	err := app.startHooks(context.Background())
	require.Error(t, err)
	require.NoError(t, app.stopHooks(context.Background()))
	require.Equal(t, 1, stopCount)
}

func TestStopHooks_skipsUnstartedPairedHooks(t *testing.T) {
	app := &App{log: zap.NewNop()}
	called := false
	app.addHook(Hook{
		Name:    "worker",
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { called = true; return nil },
	})

	require.NoError(t, app.stopHooks(context.Background()))
	require.False(t, called)
}

func TestStopHooks_stopsStartedPairedHooks(t *testing.T) {
	app := &App{log: zap.NewNop()}
	called := false
	app.addHook(Hook{
		Name:    "worker",
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { called = true; return nil },
	})

	require.NoError(t, app.startHooks(context.Background()))
	require.NoError(t, app.stopHooks(context.Background()))
	require.True(t, called)
}

func TestStopHooks_runsInReverseOrder(t *testing.T) {
	app := &App{log: zap.NewNop()}
	var order []int
	app.addHook(Hook{
		OnStop: func(context.Context) error { order = append(order, 1); return nil },
	})
	app.addHook(Hook{
		OnStop: func(context.Context) error { order = append(order, 2); return nil },
	})
	app.addHook(Hook{
		OnStop: func(context.Context) error { order = append(order, 3); return nil },
	})

	require.NoError(t, app.stopHooks(context.Background()))
	require.Equal(t, []int{3, 2, 1}, order)
}

func TestStopHooks_joinsErrors(t *testing.T) {
	app := &App{
		cfg: testSQLiteConfig(t),
		log: zap.NewNop(),
	}
	app.addHook(Hook{
		OnStop: func(context.Context) error { return errors.New("stop-one") },
	})
	app.addHook(Hook{
		OnStop: func(context.Context) error { return errors.New("stop-two") },
	})

	err := app.stopHooks(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "stop-one")
	require.ErrorContains(t, err, "stop-two")
}

func TestShutdown_stopHooksInReverseOrder(t *testing.T) {
	app := &App{
		cfg: testSQLiteConfig(t),
		log: zap.NewNop(),
	}
	var order []int
	app.addHook(Hook{
		OnStop: func(context.Context) error { order = append(order, 1); return nil },
	})
	app.addHook(Hook{
		OnStop: func(context.Context) error { order = append(order, 2); return nil },
	})
	app.addHook(Hook{
		OnStop: func(context.Context) error { order = append(order, 3); return nil },
	})

	require.NoError(t, app.shutdown())
	require.Equal(t, []int{3, 2, 1}, order)
}

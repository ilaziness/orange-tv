package app

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// Hook follows uber-go/fx lifecycle semantics (OnStart FIFO, OnStop LIFO).
type Hook struct {
	Name    string // optional, used in logs and errors
	OnStart func(ctx context.Context) error
	OnStop  func(ctx context.Context) error
}

type hookState struct {
	started bool
	stopped bool
}

func (h Hook) label(i int) string {
	if h.Name != "" {
		return h.Name
	}
	return fmt.Sprintf("hook %d", i)
}

func (a *App) addHook(h Hook) {
	a.hooks = append(a.hooks, h)
	a.hookStates = append(a.hookStates, hookState{})
}

func (a *App) startHooks(ctx context.Context) error {
	for i, hook := range a.hooks {
		if hook.OnStart == nil {
			continue
		}
		if hook.OnStop == nil && a.log != nil {
			a.log.Warn("Hook has OnStart but no OnStop; it cannot be rolled back on failure",
				zap.String("hook", hook.label(i)),
			)
		}
		if err := hook.OnStart(ctx); err != nil {
			rollbackErr := a.rollbackStartedHooks(ctx, i)
			return errors.Join(fmt.Errorf("%s OnStart: %w", hook.label(i), err), rollbackErr)
		}
		a.hookStates[i].started = true
	}
	return nil
}

func (a *App) rollbackStartedHooks(ctx context.Context, failedIndex int) error {
	var errs []error
	for j := failedIndex - 1; j >= 0; j-- {
		if err := a.stopHookAt(ctx, j, "startup rollback"); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (a *App) stopHooks(ctx context.Context) error {
	var errs []error
	for i := len(a.hooks) - 1; i >= 0; i-- {
		if err := a.stopHookAt(ctx, i, ""); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (a *App) stopHookAt(ctx context.Context, i int, phase string) error {
	if a.hookStates[i].stopped {
		return nil
	}

	hook := a.hooks[i]
	if hook.OnStop == nil {
		return nil
	}
	if hook.OnStart != nil && !a.hookStates[i].started {
		return nil
	}

	if err := hook.OnStop(ctx); err != nil {
		if a.log != nil {
			fields := []zap.Field{zap.String("hook", hook.label(i)), zap.Error(err)}
			if phase != "" {
				fields = append(fields, zap.String("phase", phase))
			}
			a.log.Error("Hook OnStop failed", fields...)
		}
		return fmt.Errorf("%s OnStop: %w", hook.label(i), err)
	}

	a.hookStates[i].stopped = true
	return nil
}

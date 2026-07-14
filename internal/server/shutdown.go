package server

import (
	"context"
	"fmt"
	"time"
)

// waitShutdown waits for done or returns when ctx is canceled or wait times out.
func waitShutdown(ctx context.Context, defaultTimeout time.Duration, done <-chan struct{}) error {
	timeout := defaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d < timeout {
			timeout = d
		}
	}
	if timeout < 0 {
		timeout = 0
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown canceled: %w", ctx.Err())
	case <-timer.C:
		return fmt.Errorf("shutdown timeout")
	}
}

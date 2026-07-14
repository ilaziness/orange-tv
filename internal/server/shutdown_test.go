package server

import (
	"context"
	"testing"
	"time"
)

func TestWaitShutdown_completesOnDone(t *testing.T) {
	done := make(chan struct{})
	close(done)

	err := waitShutdown(context.Background(), time.Second, done)
	if err != nil {
		t.Fatalf("waitShutdown() = %v, want nil", err)
	}
}

func TestWaitShutdown_respectsContextCancel(t *testing.T) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitShutdown(ctx, time.Second, done)
	if err == nil {
		t.Fatal("waitShutdown() = nil, want error")
	}
}

func TestWaitShutdown_timesOut(t *testing.T) {
	done := make(chan struct{})

	err := waitShutdown(context.Background(), 10*time.Millisecond, done)
	if err == nil {
		t.Fatal("waitShutdown() = nil, want timeout error")
	}
}

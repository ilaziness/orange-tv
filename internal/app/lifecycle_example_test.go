package app

import (
	"context"
	"fmt"
)

type exampleWorker struct{}

func (exampleWorker) Start(_ context.Context) error {
	fmt.Println("worker started")
	return nil
}

func (exampleWorker) Stop(_ context.Context) error {
	fmt.Println("worker stopped")
	return nil
}

// Example_addHook demonstrates registering startup and shutdown hooks during wiring.
func Example_addHook() {
	a := &App{}
	worker := exampleWorker{}

	// Paired init + cleanup (background worker, subscription, warm-up).
	a.addHook(Hook{
		OnStart: func(ctx context.Context) error { return worker.Start(ctx) },
		OnStop:  func(ctx context.Context) error { return worker.Stop(ctx) },
	})

	// Stop-only hook for resources created in New() that need no startup logic.
	a.addHook(Hook{
		OnStop: func(_ context.Context) error {
			fmt.Println("resource closed")
			return nil
		},
	})

	_ = a.startHooks(context.Background())
	_ = a.stopHooks(context.Background())
	// Output:
	// worker started
	// resource closed
	// worker stopped
}

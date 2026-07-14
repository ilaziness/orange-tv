// Package health provides dependency health checkers for readiness probes.
package health

import "context"

// Checker reports whether a named dependency is ready.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

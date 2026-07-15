package router

import (
	"fmt"

	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
)

// Handlers aggregates HTTP handlers for route registration.
type Handlers struct {
	Health             *httphandler.HealthHandler
	Stub               *httphandler.StubHandler
	InternalServiceKey string // optional; when set, internal API requires X-Internal-Service-Key
	// RequireAdminAuth enables RequireAuth on admin business routes.
	// Set true when JWT manager is configured; false keeps phase-1 stubs callable without tokens.
	RequireAdminAuth bool
}

// NewHandlers creates Handlers with required dependencies.
func NewHandlers(health *httphandler.HealthHandler) (*Handlers, error) {
	h := &Handlers{
		Health: health,
		Stub:   httphandler.NewStubHandler(),
	}
	if err := h.validate(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handlers) validate() error {
	if h == nil {
		return fmt.Errorf("router: handlers is nil")
	}
	if h.Health == nil {
		return fmt.Errorf("router: health handler is required")
	}
	if h.Stub == nil {
		return fmt.Errorf("router: stub handler is required")
	}
	return nil
}

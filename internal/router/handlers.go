package router

import (
	"fmt"

	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
)

// Handlers aggregates HTTP handlers for route registration.
type Handlers struct {
	Health             *httphandler.HealthHandler
	User               *httphandler.UserHandler
	InternalServiceKey string // optional; when set, internal API requires X-Internal-Service-Key
}

// NewHandlers creates Handlers with required dependencies.
func NewHandlers(health *httphandler.HealthHandler, user *httphandler.UserHandler) (*Handlers, error) {
	h := &Handlers{Health: health, User: user}
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
	if h.User == nil {
		return fmt.Errorf("router: user handler is required")
	}
	return nil
}

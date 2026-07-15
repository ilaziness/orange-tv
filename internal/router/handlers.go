package router

import (
	"fmt"

	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	adminhandler "github.com/ilaziness/orange-tv/internal/handler/http/admin"
	clienthandler "github.com/ilaziness/orange-tv/internal/handler/http/client"
	openhandler "github.com/ilaziness/orange-tv/internal/handler/http/open"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// Handlers aggregates HTTP handlers for route registration.
type Handlers struct {
	Health *httphandler.HealthHandler
	Stub   *httphandler.StubHandler

	// Admin surface
	AdminAuth     *adminhandler.AuthHandler
	AdminCategory *adminhandler.CategoryHandler
	AdminVideo    *adminhandler.VideoHandler
	AdminMetadata *adminhandler.MetadataHandler
	AdminPlay     *adminhandler.PlayHandler
	AdminLive     *adminhandler.LiveHandler
	AdminCollect  *adminhandler.CollectHandler
	AdminTheme    *adminhandler.ThemeHandler
	AdminSettings *adminhandler.SettingsHandler
	AdminLog      *adminhandler.LogHandler
	AuthService   adminsvc.AuthService

	// Client surface
	ClientCategory *clienthandler.CategoryHandler
	ClientVideo    *clienthandler.VideoHandler
	ClientLive     *clienthandler.LiveHandler
	ClientTheme    *clienthandler.ThemeHandler
	ClientSite     *clienthandler.SiteHandler

	// Open resource station
	OpenResource *openhandler.ResourceHandler

	InternalServiceKey string
	// RequireAdminAuth enables auth middleware on admin business routes.
	RequireAdminAuth bool
}

// NewHandlers creates Handlers with required shared dependencies.
// Business handlers are assigned by app wiring before RegisterRoutes.
func NewHandlers(health *httphandler.HealthHandler) (*Handlers, error) {
	h := &Handlers{
		Health: health,
		Stub:   httphandler.NewStubHandler(),
	}
	if err := h.validateBase(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handlers) validateBase() error {
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

// validateForRoutes ensures all handlers referenced by route registration are present.
func (h *Handlers) validateForRoutes() error {
	if err := h.validateBase(); err != nil {
		return err
	}
	required := []struct {
		name string
		ok   bool
	}{
		{"admin auth handler", h.AdminAuth != nil},
		{"admin category handler", h.AdminCategory != nil},
		{"admin video handler", h.AdminVideo != nil},
		{"admin metadata handler", h.AdminMetadata != nil},
		{"admin play handler", h.AdminPlay != nil},
		{"admin live handler", h.AdminLive != nil},
		{"admin collect handler", h.AdminCollect != nil},
		{"admin theme handler", h.AdminTheme != nil},
		{"admin settings handler", h.AdminSettings != nil},
		{"admin log handler", h.AdminLog != nil},
		{"client category handler", h.ClientCategory != nil},
		{"client video handler", h.ClientVideo != nil},
		{"client live handler", h.ClientLive != nil},
		{"client theme handler", h.ClientTheme != nil},
		{"client site handler", h.ClientSite != nil},
		{"open resource handler", h.OpenResource != nil},
	}
	for _, item := range required {
		if !item.ok {
			return fmt.Errorf("router: %s is required", item.name)
		}
	}
	if h.RequireAdminAuth && h.AuthService == nil {
		return fmt.Errorf("router: auth service is required when admin auth is enabled")
	}
	return nil
}

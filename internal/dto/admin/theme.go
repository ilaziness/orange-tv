package admin

// UpdateThemeRequest updates theme override fields.
type UpdateThemeRequest struct {
	Name      *string        `json:"name" validate:"omitempty,min=1,max=100"`
	Config    map[string]any `json:"config"`
	CustomCSS *string        `json:"custom_css" validate:"omitempty,max=100000"`
	CustomJS  *string        `json:"custom_js" validate:"omitempty,max=100000"`
}

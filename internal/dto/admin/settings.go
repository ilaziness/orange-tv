package admin

import (
	"encoding/json"

	dto "github.com/ilaziness/orange-tv/internal/dto"
)

// SiteSettings is an alias to the shared DTO for admin convenience.
type SiteSettings = dto.SiteSettings

// AdSettings is an alias to the shared DTO for admin convenience.
type AdSettings = dto.AdSettings

// FeatureSettings is an alias to the shared DTO for admin convenience.
type FeatureSettings = dto.FeatureSettings

// APISettings holds resource-station / API mode settings.
// ResourceAPIKey is never returned in plain text; ResourceAPIKeySet indicates whether configured.
type APISettings struct {
	SiteMode                string `json:"site_mode"`
	APIOutputFormat         string `json:"api_output_format"`
	EnableThirdPartyCollect bool   `json:"enable_third_party_collect"`
	ResourceAPIKeySet       bool   `json:"resource_api_key_set"`
	ResourceAPIKeyMasked    string `json:"resource_api_key_masked,omitempty"`
}

// UpdateSettingsRequest updates settings for a single group.
// Data is the group-specific key-value JSON payload.
type UpdateSettingsRequest struct {
	Group string          `json:"group" validate:"required,oneof=site api ad feature"`
	Data  json.RawMessage `json:"data" validate:"required"`
}

// UpdateSiteSettings updates public site fields (all optional).
type UpdateSiteSettings struct {
	Name        *string `json:"name" validate:"omitempty,max=100"`
	Logo        *string `json:"logo" validate:"omitempty,max=500"`
	Copyright   *string `json:"copyright" validate:"omitempty,max=255"`
	ICP         *string `json:"icp" validate:"omitempty,max=100"`
	SEOKeywords *string `json:"seo_keywords" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// UpdateAPISettings updates API / resource station fields (all optional).
type UpdateAPISettings struct {
	SiteMode                *string `json:"site_mode" validate:"omitempty,oneof=video_site resource_site"`
	APIOutputFormat         *string `json:"api_output_format" validate:"omitempty,oneof=default apple_cms"`
	EnableThirdPartyCollect *bool   `json:"enable_third_party_collect"`
	// ResourceAPIKey empty = leave unchanged; non-empty replaces key.
	ResourceAPIKey *string `json:"resource_api_key" validate:"omitempty,max=255"`
}

// UpdateAdSettings updates video ad fields (all optional).
type UpdateAdSettings struct {
	Enabled  *bool   `json:"enabled"`
	Type     *string `json:"type" validate:"omitempty,oneof=image video html"`
	URL      *string `json:"url" validate:"omitempty,max=500"`
	Link     *string `json:"link" validate:"omitempty,max=500"`
	Duration *int    `json:"duration" validate:"omitempty,min=1,max=300"`
	Skipable *bool   `json:"skipable"`
}

// UpdateFeatureSettings updates client feature toggles (all optional).
type UpdateFeatureSettings struct {
	LiveEnabled    *bool `json:"live_enabled"`
	CommentEnabled *bool `json:"comment_enabled"`
	CommentReview  *bool `json:"comment_review"`
	RatingEnabled  *bool `json:"rating_enabled"`
}

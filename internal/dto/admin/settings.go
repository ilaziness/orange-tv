package admin

import (
	"encoding/json"

	dto "github.com/ilaziness/orange-tv/internal/dto"
)

// SiteSettings is an alias to the shared DTO for admin convenience.
type SiteSettings = dto.SiteSettings

// FeatureSettings is an alias to the shared DTO for admin convenience.
type FeatureSettings = dto.FeatureSettings

// APISettings holds resource-station / API mode settings.
type APISettings struct {
	EnableThirdPartyCollect bool `json:"enable_third_party_collect"`
}

// GetSettingsQuery binds the group query parameter.
type GetSettingsQuery struct {
	Group string `form:"group" binding:"required,oneof=site api feature"`
}

// UpdateSettingsRequest updates settings for a single group.
// Data is the group-specific key-value JSON payload.
type UpdateSettingsRequest struct {
	Group string          `json:"group" binding:"required,oneof=site api feature"`
	Data  json.RawMessage `json:"data" binding:"required"`
}

// UpdateSiteSettings updates public site fields (all optional).
type UpdateSiteSettings struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Logo        *string `json:"logo" binding:"omitempty,max=500"`
	Copyright   *string `json:"copyright" binding:"omitempty,max=255"`
	ICP         *string `json:"icp" binding:"omitempty,max=100"`
	SEOKeywords *string `json:"seo_keywords" binding:"omitempty,max=255"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// UpdateAPISettings updates API / resource station fields (all optional).
type UpdateAPISettings struct {
	EnableThirdPartyCollect *bool `json:"enable_third_party_collect"`
}

// UpdateFeatureSettings updates client feature toggles (all optional).
type UpdateFeatureSettings struct {
	LiveEnabled    *bool `json:"live_enabled"`
	CommentEnabled *bool `json:"comment_enabled"`
	CommentReview  *bool `json:"comment_review"`
	RatingEnabled  *bool `json:"rating_enabled"`
}

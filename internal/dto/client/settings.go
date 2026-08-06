package client

import dto "github.com/ilaziness/orange-tv/internal/dto"

// SiteSettings is an alias to the shared DTO for client convenience.
type SiteSettings = dto.SiteSettings

// FeatureSettings is an alias to the shared DTO for client convenience.
type FeatureSettings = dto.FeatureSettings

// GetSettingsQuery binds the multi-value groups query parameter.
type GetSettingsQuery struct {
	Groups []string `form:"groups" binding:"required,min=1,dive,oneof=site feature"`
}

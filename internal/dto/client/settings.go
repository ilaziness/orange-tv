package client

import dto "github.com/ilaziness/orange-tv/internal/dto"

// SiteSettings is an alias to the shared DTO for client convenience.
type SiteSettings = dto.SiteSettings

// FeatureSettings is an alias to the shared DTO for client convenience.
type FeatureSettings = dto.FeatureSettings

// PublicSEOSettings is an alias to the shared DTO for client convenience.
type PublicSEOSettings = dto.PublicSEOSettings

// GetSettingsQuery binds the multi-value groups query parameter.
type GetSettingsQuery struct {
	// 配置分组（site=站点信息，feature=功能开关，seo=SEO公开字段），可传多个
	Groups []string `form:"groups" binding:"required,min=1,dive,oneof=site feature seo"`
}

package admin

import (
	dto "github.com/ilaziness/orange-tv/internal/dto"
)

// SiteSettings is an alias to the shared DTO for admin convenience.
type SiteSettings = dto.SiteSettings

// FeatureSettings is an alias to the shared DTO for admin convenience.
type FeatureSettings = dto.FeatureSettings

// APISettings holds resource-station / API mode settings.
type APISettings struct {
	// 是否启用第三方资源站采集
	EnableThirdPartyCollect bool `json:"enable_third_party_collect"`
}

// GetSettingsQuery binds the group query parameter.
type GetSettingsQuery struct {
	// 配置分组（site=站点信息，api=API/资源站，feature=功能开关）
	Group string `form:"group" binding:"required,oneof=site api feature"`
}

// UpdateSettingsRequest updates settings for a single group.
// Data is the group-specific key-value JSON payload. The service layer unmarshals it
// into the per-group struct (UpdateSiteSettings/UpdateAPISettings/UpdateFeatureSettings)
// and upserts each string/bool value into its own system_settings row; the raw payload
// is therefore never stored as a whole, so decoding it as `any` and re-marshaling is
// semantically identical to keeping the raw bytes.
type UpdateSettingsRequest struct {
	// 配置分组（必填：site=站点信息，api=API/资源站，feature=功能开关）
	Group string `json:"group" binding:"required,oneof=site api feature"`
	// 分组配置键值 JSON 数据（结构随分组变化：site=UpdateSiteSettings，api=UpdateAPISettings，feature=UpdateFeatureSettings）
	Data any `json:"data" binding:"required"`
}

// UpdateSiteSettings updates public site fields (all optional).
type UpdateSiteSettings struct {
	// 站点名称
	Name *string `json:"name" binding:"omitempty,max=100"`
	// 站点 Logo 地址
	Logo *string `json:"logo" binding:"omitempty,max=500"`
	// 版权信息
	Copyright *string `json:"copyright" binding:"omitempty,max=255"`
	// 备案号
	ICP *string `json:"icp" binding:"omitempty,max=100"`
	// SEO 关键词
	SEOKeywords *string `json:"seo_keywords" binding:"omitempty,max=255"`
	// 站点描述
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// UpdateAPISettings updates API / resource station fields (all optional).
type UpdateAPISettings struct {
	// 是否启用第三方资源站采集
	EnableThirdPartyCollect *bool `json:"enable_third_party_collect"`
}

// UpdateFeatureSettings updates client feature toggles (all optional).
type UpdateFeatureSettings struct {
	// 是否启用直播功能
	LiveEnabled *bool `json:"live_enabled"`
	// 是否启用评论功能
	CommentEnabled *bool `json:"comment_enabled"`
	// 评论是否需要审核
	CommentReview *bool `json:"comment_review"`
	// 是否启用评分功能
	RatingEnabled *bool `json:"rating_enabled"`
}

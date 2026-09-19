package admin

import (
	dto "github.com/ilaziness/orange-tv/internal/dto"
)

// SiteSettings is an alias to the shared DTO for admin convenience.
type SiteSettings = dto.SiteSettings

// FeatureMatrix is an alias to the shared DTO for admin convenience.
type FeatureMatrix = dto.FeatureMatrix

// PlatformFlags is an alias to the shared DTO for admin convenience.
type PlatformFlags = dto.PlatformFlags

// APISettings holds resource-station / API mode settings.
type APISettings struct {
	// 是否启用第三方资源站采集
	EnableThirdPartyCollect bool `json:"enable_third_party_collect"`
}

// PaymentSettings is an alias for admin convenience.
type PaymentSettings = dto.PaymentSettings

// GetSettingsQuery binds the group query parameter.
type GetSettingsQuery struct {
	// 配置分组（site=站点信息，api=API/资源站，feature=功能开关，seo=SEO设置，storage=云存储，payment=支付商）
	Group string `form:"group" binding:"required,oneof=site api feature seo storage payment"`
}

// UpdateSettingsRequest updates settings for a single group.
// Data is the group-specific key-value JSON payload. The service layer unmarshals it
// into the per-group struct (UpdateSiteSettings/UpdateAPISettings/UpdateFeatureSettings/UpdateSEOSettings/UpdateStorageSettings)
// and upserts each string/bool value into its own system_settings row; the raw payload
// is therefore never stored as a whole, so decoding it as `any` and re-marshaling is
// semantically identical to keeping the raw bytes.
type UpdateSettingsRequest struct {
	// 配置分组（必填：site=站点信息，api=API/资源站，feature=功能开关，seo=SEO设置，storage=云存储，payment=支付商）
	Group string `json:"group" binding:"required,oneof=site api feature seo storage payment"`
	// 分组配置键值 JSON 数据（结构随分组变化）
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
	// 站点统计代码（百度统计、Google Analytics 等）
	AnalyticsCode *string `json:"analytics_code" binding:"omitempty,max=2048"`
}

// UpdateAPISettings updates API / resource station fields (all optional).
type UpdateAPISettings struct {
	// 是否启用第三方资源站采集
	EnableThirdPartyCollect *bool `json:"enable_third_party_collect"`
}

// UpdateFeatureSettings updates client feature toggles (all optional).
// When a field is present, all four platform booleans are written together.
type UpdateFeatureSettings struct {
	// 是否启用电视直播功能（按端）
	LiveTVEnabled *dto.PlatformFlags `json:"livetv_enabled"`
	// 是否启用评论功能（按端）
	CommentEnabled *dto.PlatformFlags `json:"comment_enabled"`
	// 评论是否需要审核（按端）
	CommentReview *dto.PlatformFlags `json:"comment_review"`
	// 是否启用评分功能（按端）
	RatingEnabled *dto.PlatformFlags `json:"rating_enabled"`
}

// UpdateSEOSettings updates SEO / social-sharing fields (all optional).
type UpdateSEOSettings struct {
	// 公开站点根地址（无尾斜杠）
	PublicBaseURL *string `json:"public_base_url" binding:"omitempty,max=500"`
	// 默认 Open Graph 图片地址
	DefaultOGImage *string `json:"default_og_image" binding:"omitempty,max=500"`
	// 是否输出 sitemap
	SitemapEnabled *bool `json:"sitemap_enabled"`
	// 是否输出 llms.txt
	LLMsEnabled *bool `json:"llms_enabled"`
	// llms.txt 站点简介
	LLMsIntro *string `json:"llms_intro" binding:"omitempty,max=2000"`
	// 是否允许 AI 检索类爬虫
	AllowAISearch *bool `json:"allow_ai_search"`
	// 是否允许 AI 训练类爬虫
	AllowAITraining *bool `json:"allow_ai_training"`
	// Google Search Console 站点验证码
	GoogleSiteVerification *string `json:"google_site_verification" binding:"omitempty,max=255"`
	// 百度站长验证码
	BaiduSiteVerification *string `json:"baidu_site_verification" binding:"omitempty,max=255"`
	// Bing Webmaster 验证码
	BingSiteVerification *string `json:"bing_site_verification" binding:"omitempty,max=255"`
}

// UpdatePaymentAlipayConfig updates Alipay merchant JSON (all optional).
// Empty secret/cert fields keep the stored values.
type UpdatePaymentAlipayConfig struct {
	Enabled          *bool   `json:"enabled"`
	Sandbox          *bool   `json:"sandbox"`
	AppID            *string `json:"app_id" binding:"omitempty,max=64"`
	SignMode         *string `json:"sign_mode" binding:"omitempty,oneof=key cert"`
	PrivateKey       *string `json:"private_key"`
	AlipayPublicKey  *string `json:"alipay_public_key"`
	AppCert          *string `json:"app_cert"`
	AlipayPublicCert *string `json:"alipay_public_cert"`
	AlipayRootCert   *string `json:"alipay_root_cert"`
	NotifyURL        *string `json:"notify_url" binding:"omitempty,max=500"`
	ReturnURL        *string `json:"return_url" binding:"omitempty,max=500"`
	PCWebEnabled     *bool   `json:"pc_web_enabled"`
	AppEnabled       *bool   `json:"app_enabled"`
}

// UpdatePaymentWechatConfig updates WeChat Pay merchant JSON (all optional).
type UpdatePaymentWechatConfig struct {
	Enabled      *bool   `json:"enabled"`
	MchID        *string `json:"mch_id" binding:"omitempty,max=32"`
	MchSerialNo  *string `json:"mch_serial_no" binding:"omitempty,max=64"`
	APIv3Key     *string `json:"api_v3_key"`
	PrivateKey   *string `json:"private_key"`
	AppIDWeb     *string `json:"app_id_web" binding:"omitempty,max=64"`
	AppIDApp     *string `json:"app_id_app" binding:"omitempty,max=64"`
	NotifyURL    *string `json:"notify_url" binding:"omitempty,max=500"`
	PCWebEnabled *bool   `json:"pc_web_enabled"`
	AppEnabled   *bool   `json:"app_enabled"`
}

// UpdatePaymentSettings updates one or both payment merchants.
type UpdatePaymentSettings struct {
	Alipay *UpdatePaymentAlipayConfig `json:"alipay"`
	Wechat *UpdatePaymentWechatConfig `json:"wechat"`
}

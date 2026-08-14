package dto

// SiteSettings holds public site branding fields.
type SiteSettings struct {
	// 站点名称
	Name string `json:"name"`
	// 站点 Logo 地址
	Logo string `json:"logo"`
	// 版权信息
	Copyright string `json:"copyright"`
	// 备案号
	ICP string `json:"icp"`
	// SEO 关键词
	SEOKeywords string `json:"seo_keywords"`
	// 站点描述
	Description string `json:"description"`
}

// FeatureSettings holds client feature toggle settings.
type FeatureSettings struct {
	// 是否启用直播功能
	LiveEnabled bool `json:"live_enabled"`
	// 是否启用评论功能
	CommentEnabled bool `json:"comment_enabled"`
	// 评论是否需要审核
	CommentReview bool `json:"comment_review"`
	// 是否启用评分功能
	RatingEnabled bool `json:"rating_enabled"`
}

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
	// 站点统计代码（百度统计、Google Analytics 等）
	AnalyticsCode string `json:"analytics_code"`
}

// FeatureSettings holds client feature toggle settings.
type FeatureSettings struct {
	// 是否启用电视直播功能
	LiveTVEnabled bool `json:"livetv_enabled"`
	// 是否启用评论功能
	CommentEnabled bool `json:"comment_enabled"`
	// 评论是否需要审核
	CommentReview bool `json:"comment_review"`
	// 是否启用评分功能
	RatingEnabled bool `json:"rating_enabled"`
}

// SEOSettings holds admin SEO / social-sharing configuration.
type SEOSettings struct {
	// 公开站点根地址（无尾斜杠），用于 canonical / sitemap / OG
	PublicBaseURL string `json:"public_base_url"`
	// 默认 Open Graph 图片地址
	DefaultOGImage string `json:"default_og_image"`
	// 是否输出 sitemap
	SitemapEnabled bool `json:"sitemap_enabled"`
	// 是否输出 llms.txt
	LLMsEnabled bool `json:"llms_enabled"`
	// llms.txt 站点简介
	LLMsIntro string `json:"llms_intro"`
	// 是否允许 AI 检索类爬虫
	AllowAISearch bool `json:"allow_ai_search"`
	// 是否允许 AI 训练类爬虫
	AllowAITraining bool `json:"allow_ai_training"`
	// Google Search Console 站点验证码
	GoogleSiteVerification string `json:"google_site_verification"`
	// 百度站长验证码
	BaiduSiteVerification string `json:"baidu_site_verification"`
	// Bing Webmaster 验证码
	BingSiteVerification string `json:"bing_site_verification"`
}

// PublicSEOSettings is the client-visible subset of SEO settings.
type PublicSEOSettings struct {
	// 公开站点根地址（无尾斜杠）
	PublicBaseURL string `json:"public_base_url"`
	// 默认 Open Graph 图片地址
	DefaultOGImage string `json:"default_og_image"`
	// Google Search Console 站点验证码
	GoogleSiteVerification string `json:"google_site_verification"`
	// 百度站长验证码
	BaiduSiteVerification string `json:"baidu_site_verification"`
	// Bing Webmaster 验证码
	BingSiteVerification string `json:"bing_site_verification"`
}

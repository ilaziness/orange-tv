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

// PlatformFlags holds per-platform (web/desktop/app/tv) enable flags for one feature.
type PlatformFlags struct {
	// 网页端
	Web bool `json:"web"`
	// 桌面端
	Desktop bool `json:"desktop"`
	// 移动 App 端
	App bool `json:"app"`
	// 电视端
	TV bool `json:"tv"`
}

// FeatureMatrix holds admin feature toggles as a per-platform matrix.
type FeatureMatrix struct {
	// 是否启用电视直播功能（按端）
	LiveTVEnabled PlatformFlags `json:"livetv_enabled"`
	// 是否启用评论功能（按端）
	CommentEnabled PlatformFlags `json:"comment_enabled"`
	// 评论是否需要审核（按端）
	CommentReview PlatformFlags `json:"comment_review"`
	// 是否启用评分功能（按端）
	RatingEnabled PlatformFlags `json:"rating_enabled"`
}

// FeatureSettings holds client feature toggles flattened for the current client type.
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

// StorageProviderConfig is one cloud vendor credential set (GET is masked).
type StorageProviderConfig struct {
	// 存储桶 / 空间名
	Bucket string `json:"bucket"`
	// 区域（如 oss-cn-hangzhou、ap-guangzhou、z0）
	Region string `json:"region"`
	// 可选自定义 API Endpoint
	Endpoint string `json:"endpoint"`
	// AccessKey / SecretId；GET 永远为空，用 access_configured 判断是否已配置
	AccessKey string `json:"access_key"`
	// SecretKey；GET 永远为空，用 secret_configured 判断是否已配置
	SecretKey string `json:"secret_key"`
	// 是否已配置 AccessKey（仅响应）
	AccessConfigured bool `json:"access_configured"`
	// 是否已配置 Secret（仅响应）
	SecretConfigured bool `json:"secret_configured"`
	// 加速域名（https，无尾斜杠）
	CDNDomain string `json:"cdn_domain"`
}

// StorageSettings is the admin storage group response.
type StorageSettings struct {
	// 当前启用厂商：none / aliyun / tencent / qiniu
	Provider string `json:"provider"`
	// 阿里云 OSS
	Aliyun StorageProviderConfig `json:"aliyun"`
	// 腾讯云 COS
	Tencent StorageProviderConfig `json:"tencent"`
	// 七牛云 Kodo
	Qiniu StorageProviderConfig `json:"qiniu"`
}

// MediaAsset is a media library item returned by upload/list APIs.
type MediaAsset struct {
	// 媒体 ID
	ID uint64 `json:"id"`
	// 媒体类型（当前仅 image）
	MediaType string `json:"media_type"`
	// 上传方 admin/user
	OwnerKind string `json:"owner_kind"`
	// 上传方 ID
	OwnerID uint32 `json:"owner_id"`
	// 云厂商
	Provider string `json:"provider"`
	// 加速域名公网地址
	URL string `json:"url"`
	// MIME
	Mime string `json:"mime"`
	// 大小（字节）
	Size uint64 `json:"size"`
	// 原始文件名
	OriginalName string `json:"original_name"`
	// 宽
	Width uint32 `json:"width"`
	// 高
	Height uint32 `json:"height"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

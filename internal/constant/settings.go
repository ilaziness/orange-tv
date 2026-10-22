package constant

// Setting group constants.
const (
	SettingGroupSite    = "site"    // 站点设置
	SettingGroupAPI     = "api"     // API/资源站设置
	SettingGroupFeature = "feature" // 功能设置
	SettingGroupLiveTV  = "livetv"  // 电视直播设置
	SettingGroupSEO     = "seo"     // SEO / 社交分享设置
	SettingGroupStorage = "storage" // 云存储设置
	SettingGroupPayment = "payment" // 支付商设置
	SettingGroupLive    = "live"    // 用户直播推流配置（admin-only）
)

// AllSettingGroups returns all valid setting group names.
func AllSettingGroups() []string {
	return []string{SettingGroupSite, SettingGroupAPI, SettingGroupFeature, SettingGroupLiveTV, SettingGroupSEO, SettingGroupStorage, SettingGroupPayment, SettingGroupLive}
}

// System setting keys.
const (
	SettingEnableThirdPartyCollect = "enable_third_party_collect"
	SettingSiteName                = "site_name"
	SettingSiteLogo                = "site_logo"
	SettingSiteCopyright           = "site_copyright"
	SettingSiteICP                 = "site_icp"
	SettingSiteSEOKeywords         = "site_seo_keywords"
	SettingSiteDescription         = "site_description"
	SettingSiteAnalyticsCode       = "site_analytics_code"

	// Feature settings keys.
	SettingFeatureLiveTVEnabled  = "livetv_enabled"  // 电视直播开关
	SettingFeatureLiveEnabled    = "live_enabled"    // 用户直播开关
	SettingFeatureCommentEnabled = "comment_enabled" // 视频评论开关
	SettingFeatureCommentReview  = "comment_review"  // 评论是否需要审核
	SettingFeatureRatingEnabled  = "rating_enabled"  // 视频评分开关

	// User live streaming (admin-only group live) keys.
	SettingLivePushServer      = "live_push_server"       // 自建 OBS 推流服务器
	SettingLivePlayHLSTemplate = "live_play_hls_template" // 自建 HLS 播放模板
	SettingLivePlayFLVTemplate = "live_play_flv_template" // 自建 FLV 播放模板
	SettingLiveProvider        = "live_provider"          // self / aliyun / qiniu
	SettingLiveAuthEnabled     = "live_auth_enabled"      // 仅自建：是否开启推流/播放鉴权
	SettingLivePushTTLSeconds  = "live_push_ttl_seconds"  // 推流签名有效秒数
	SettingLivePlayTTLSeconds  = "live_play_ttl_seconds"  // 播放签名有效秒数
	SettingLiveSignSecret      = "live_sign_secret"       // 自建 HMAC 签名密钥
	SettingLiveHookSecret      = "live_hook_secret"       // SRS 回调查询参数密钥
	SettingLiveSRSAPIBase      = "live_srs_api_base"      // SRS HTTP API，如 http://127.0.0.1:1985
	SettingLiveAliyun          = "live_aliyun"            // 阿里云直播 JSON
	SettingLiveQiniu           = "live_qiniu"             // 七牛直播 JSON

	// LiveTV settings keys.
	SettingLiveTVSyncSourceURL = "livetv_sync_source_url" // 直播源同步地址

	// SEO settings keys.
	SettingSEOPublicBaseURL          = "seo_public_base_url"
	SettingSEODefaultOGImage         = "seo_default_og_image"
	SettingSEOSitemapEnabled         = "seo_sitemap_enabled"
	SettingSEOLLMsEnabled            = "seo_llms_enabled"
	SettingSEOLLMsIntro              = "seo_llms_intro"
	SettingSEOAllowAISearch          = "seo_allow_ai_search"
	SettingSEOAllowAITraining        = "seo_allow_ai_training"
	SettingSEOGoogleSiteVerification = "seo_google_site_verification"
	SettingSEOBaiduSiteVerification  = "seo_baidu_site_verification"
	SettingSEOBingSiteVerification   = "seo_bing_site_verification"

	// Storage settings keys.
	SettingStorageProvider = "storage_provider" // 当前启用厂商 none/aliyun/tencent/qiniu
	SettingStorageAliyun   = "storage_aliyun"   // 阿里云配置 JSON
	SettingStorageTencent  = "storage_tencent"  // 腾讯云配置 JSON
	SettingStorageQiniu    = "storage_qiniu"    // 七牛云配置 JSON

	// Payment settings keys.
	SettingPaymentAlipay = "payment_alipay" // 支付宝配置 JSON
	SettingPaymentWechat = "payment_wechat" // 微信支付配置 JSON
)

// Storage provider values.
const (
	StorageProviderNone    = "none"
	StorageProviderAliyun  = "aliyun"
	StorageProviderTencent = "tencent"
	StorageProviderQiniu   = "qiniu"
)

// Media type / owner constants.
const (
	MediaTypeImage = "image"

	MediaOwnerAdmin = "admin"
	MediaOwnerUser  = "user"
)

// GroupKeys maps each setting group to its constituent key list.
var GroupKeys = map[string][]string{
	SettingGroupSite: {
		SettingSiteName,
		SettingSiteLogo,
		SettingSiteCopyright,
		SettingSiteICP,
		SettingSiteSEOKeywords,
		SettingSiteDescription,
		SettingSiteAnalyticsCode,
	},
	SettingGroupAPI: {
		SettingEnableThirdPartyCollect,
	},
	SettingGroupFeature: {
		SettingFeatureLiveTVEnabled,
		SettingFeatureLiveEnabled,
		SettingFeatureCommentEnabled,
		SettingFeatureCommentReview,
		SettingFeatureRatingEnabled,
	},
	SettingGroupLive: {
		SettingLivePushServer,
		SettingLivePlayHLSTemplate,
		SettingLivePlayFLVTemplate,
		SettingLiveProvider,
		SettingLiveAuthEnabled,
		SettingLivePushTTLSeconds,
		SettingLivePlayTTLSeconds,
		SettingLiveSignSecret,
		SettingLiveHookSecret,
		SettingLiveSRSAPIBase,
		SettingLiveAliyun,
		SettingLiveQiniu,
	},
	SettingGroupLiveTV: {
		SettingLiveTVSyncSourceURL,
	},
	SettingGroupSEO: {
		SettingSEOPublicBaseURL,
		SettingSEODefaultOGImage,
		SettingSEOSitemapEnabled,
		SettingSEOLLMsEnabled,
		SettingSEOLLMsIntro,
		SettingSEOAllowAISearch,
		SettingSEOAllowAITraining,
		SettingSEOGoogleSiteVerification,
		SettingSEOBaiduSiteVerification,
		SettingSEOBingSiteVerification,
	},
	SettingGroupStorage: {
		SettingStorageProvider,
		SettingStorageAliyun,
		SettingStorageTencent,
		SettingStorageQiniu,
	},
	SettingGroupPayment: {
		SettingPaymentAlipay,
		SettingPaymentWechat,
	},
}

// KeyToGroup maps each setting key to its group.
var KeyToGroup = map[string]string{
	SettingSiteName:                  SettingGroupSite,
	SettingSiteLogo:                  SettingGroupSite,
	SettingSiteCopyright:             SettingGroupSite,
	SettingSiteICP:                   SettingGroupSite,
	SettingSiteSEOKeywords:           SettingGroupSite,
	SettingSiteDescription:           SettingGroupSite,
	SettingSiteAnalyticsCode:         SettingGroupSite,
	SettingEnableThirdPartyCollect:   SettingGroupAPI,
	SettingFeatureLiveTVEnabled:      SettingGroupFeature,
	SettingFeatureLiveEnabled:        SettingGroupFeature,
	SettingFeatureCommentEnabled:     SettingGroupFeature,
	SettingFeatureCommentReview:      SettingGroupFeature,
	SettingFeatureRatingEnabled:      SettingGroupFeature,
	SettingLivePushServer:            SettingGroupLive,
	SettingLivePlayHLSTemplate:       SettingGroupLive,
	SettingLivePlayFLVTemplate:       SettingGroupLive,
	SettingLiveProvider:              SettingGroupLive,
	SettingLiveAuthEnabled:           SettingGroupLive,
	SettingLivePushTTLSeconds:        SettingGroupLive,
	SettingLivePlayTTLSeconds:        SettingGroupLive,
	SettingLiveSignSecret:            SettingGroupLive,
	SettingLiveHookSecret:            SettingGroupLive,
	SettingLiveSRSAPIBase:            SettingGroupLive,
	SettingLiveAliyun:                SettingGroupLive,
	SettingLiveQiniu:                 SettingGroupLive,
	SettingLiveTVSyncSourceURL:       SettingGroupLiveTV,
	SettingSEOPublicBaseURL:          SettingGroupSEO,
	SettingSEODefaultOGImage:         SettingGroupSEO,
	SettingSEOSitemapEnabled:         SettingGroupSEO,
	SettingSEOLLMsEnabled:            SettingGroupSEO,
	SettingSEOLLMsIntro:              SettingGroupSEO,
	SettingSEOAllowAISearch:          SettingGroupSEO,
	SettingSEOAllowAITraining:        SettingGroupSEO,
	SettingSEOGoogleSiteVerification: SettingGroupSEO,
	SettingSEOBaiduSiteVerification:  SettingGroupSEO,
	SettingSEOBingSiteVerification:   SettingGroupSEO,
	SettingStorageProvider:           SettingGroupStorage,
	SettingStorageAliyun:             SettingGroupStorage,
	SettingStorageTencent:            SettingGroupStorage,
	SettingStorageQiniu:              SettingGroupStorage,
	SettingPaymentAlipay:             SettingGroupPayment,
	SettingPaymentWechat:             SettingGroupPayment,
}

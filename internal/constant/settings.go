package constant

// Setting group constants.
const (
	SettingGroupSite    = "site"    // 站点设置
	SettingGroupAPI     = "api"     // API/资源站设置
	SettingGroupFeature = "feature" // 功能设置
	SettingGroupLiveTV  = "livetv"  // 电视直播设置
	SettingGroupSEO     = "seo"     // SEO / 社交分享设置
)

// AllSettingGroups returns all valid setting group names.
func AllSettingGroups() []string {
	return []string{SettingGroupSite, SettingGroupAPI, SettingGroupFeature, SettingGroupLiveTV, SettingGroupSEO}
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
	SettingFeatureCommentEnabled = "comment_enabled" // 视频评论开关
	SettingFeatureCommentReview  = "comment_review"  // 评论是否需要审核
	SettingFeatureRatingEnabled  = "rating_enabled"  // 视频评分开关

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
		SettingFeatureCommentEnabled,
		SettingFeatureCommentReview,
		SettingFeatureRatingEnabled,
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
	SettingFeatureCommentEnabled:     SettingGroupFeature,
	SettingFeatureCommentReview:      SettingGroupFeature,
	SettingFeatureRatingEnabled:      SettingGroupFeature,
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
}

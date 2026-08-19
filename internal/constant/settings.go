package constant

// Setting group constants.
const (
	SettingGroupSite    = "site"    // 站点设置
	SettingGroupAPI     = "api"     // API/资源站设置
	SettingGroupFeature = "feature" // 功能设置
	SettingGroupLive    = "live"    // 直播设置
)

// AllSettingGroups returns all valid setting group names.
func AllSettingGroups() []string {
	return []string{SettingGroupSite, SettingGroupAPI, SettingGroupFeature, SettingGroupLive}
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
	SettingFeatureLiveEnabled    = "live_enabled"    // 电视直播开关
	SettingFeatureCommentEnabled = "comment_enabled" // 视频评论开关
	SettingFeatureCommentReview  = "comment_review"  // 评论是否需要审核
	SettingFeatureRatingEnabled  = "rating_enabled"  // 视频评分开关

	// Live settings keys.
	SettingLiveSyncSourceURL = "live_sync_source_url" // 直播源同步地址
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
		SettingFeatureLiveEnabled,
		SettingFeatureCommentEnabled,
		SettingFeatureCommentReview,
		SettingFeatureRatingEnabled,
	},
	SettingGroupLive: {
		SettingLiveSyncSourceURL,
	},
}

// KeyToGroup maps each setting key to its group.
var KeyToGroup = map[string]string{
	SettingSiteName:                SettingGroupSite,
	SettingSiteLogo:                SettingGroupSite,
	SettingSiteCopyright:           SettingGroupSite,
	SettingSiteICP:                 SettingGroupSite,
	SettingSiteSEOKeywords:         SettingGroupSite,
	SettingSiteDescription:         SettingGroupSite,
	SettingSiteAnalyticsCode:       SettingGroupSite,
	SettingEnableThirdPartyCollect: SettingGroupAPI,
	SettingFeatureLiveEnabled:      SettingGroupFeature,
	SettingFeatureCommentEnabled:   SettingGroupFeature,
	SettingFeatureCommentReview:    SettingGroupFeature,
	SettingFeatureRatingEnabled:    SettingGroupFeature,
	SettingLiveSyncSourceURL:       SettingGroupLive,
}

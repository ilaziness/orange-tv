package constant

// Setting group constants.
const (
	SettingGroupSite = "site" // 站点设置
	SettingGroupAPI  = "api"  // API/资源站设置
	SettingGroupAd   = "ad"   // 视频广告设置
)

// SettingGroupsDesc maps group constant to human-readable description.
var SettingGroupsDesc = map[string]string{
	SettingGroupSite: "站点设置",
	SettingGroupAPI:  "API/资源站设置",
	SettingGroupAd:   "视频广告设置",
}

// AllSettingGroups returns all valid setting group names.
func AllSettingGroups() []string {
	return []string{SettingGroupSite, SettingGroupAPI, SettingGroupAd}
}

// IsValidSettingGroup reports whether the given group is a known setting group.
func IsValidSettingGroup(group string) bool {
	_, ok := SettingGroupsDesc[group]
	return ok
}

// ClientSettingGroupWhitelist defines which setting groups are visible to the client API.
var ClientSettingGroupWhitelist = map[string]bool{
	SettingGroupSite: true,
	SettingGroupAd:   true,
}

// IsClientAllowedGroup reports whether the given group is in the client whitelist.
func IsClientAllowedGroup(group string) bool {
	return ClientSettingGroupWhitelist[group]
}

// ClientAllowedGroups returns whitelisted groups in a deterministic order.
func ClientAllowedGroups() []string {
	return []string{SettingGroupSite, SettingGroupAd}
}

// System setting keys.
const (
	SettingSiteMode                = "site_mode"
	SettingAPIOutputFormat         = "api_output_format"
	SettingEnableThirdPartyCollect = "enable_third_party_collect"
	SettingSiteName                = "site_name"
	SettingSiteLogo                = "site_logo"
	SettingSiteCopyright           = "site_copyright"
	SettingSiteICP                 = "site_icp"
	SettingSiteSEOKeywords         = "site_seo_keywords"
	SettingSiteDescription         = "site_description"
	SettingResourceAPIKey          = "resource_api_key"
	SettingVideoAdEnabled          = "video_ad_enabled"
	SettingVideoAdType             = "video_ad_type"
	SettingVideoAdUrl              = "video_ad_url"
	SettingVideoAdLink             = "video_ad_link"
	SettingVideoAdDuration         = "video_ad_duration"
	SettingVideoAdSkipable         = "video_ad_skipable"
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
	},
	SettingGroupAPI: {
		SettingSiteMode,
		SettingAPIOutputFormat,
		SettingEnableThirdPartyCollect,
		SettingResourceAPIKey,
	},
	SettingGroupAd: {
		SettingVideoAdEnabled,
		SettingVideoAdType,
		SettingVideoAdUrl,
		SettingVideoAdLink,
		SettingVideoAdDuration,
		SettingVideoAdSkipable,
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
	SettingSiteMode:                SettingGroupAPI,
	SettingAPIOutputFormat:         SettingGroupAPI,
	SettingEnableThirdPartyCollect: SettingGroupAPI,
	SettingResourceAPIKey:          SettingGroupAPI,
	SettingVideoAdEnabled:          SettingGroupAd,
	SettingVideoAdType:             SettingGroupAd,
	SettingVideoAdUrl:              SettingGroupAd,
	SettingVideoAdLink:             SettingGroupAd,
	SettingVideoAdDuration:         SettingGroupAd,
	SettingVideoAdSkipable:         SettingGroupAd,
}

// Site mode values.
const (
	SiteModeVideoSite    = "video_site"
	SiteModeResourceSite = "resource_site"
)

// API output format values for resource open API.
const (
	APIOutputDefault  = "default"   // 系统默认/自有 JSON 格式
	APIOutputAppleCMS = "apple_cms" // 苹果 CMS 兼容
)

// Video ad type values.
const (
	VideoAdTypeImage = "image"
	VideoAdTypeVideo = "video"
	VideoAdTypeHTML  = "html"
)

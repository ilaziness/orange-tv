package admin

// SettingsResponse is the structured system settings payload for admin.
type SettingsResponse struct {
	Site SiteSettings `json:"site"`
	API  APISettings  `json:"api"`
}

// SiteSettings holds public site branding fields.
type SiteSettings struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Copyright   string `json:"copyright"`
	ICP         string `json:"icp"`
	SEOKeywords string `json:"seo_keywords"`
	Description string `json:"description"`
}

// APISettings holds resource-station / API mode settings.
// ResourceAPIKey is never returned in plain text; ResourceAPIKeySet indicates whether configured.
type APISettings struct {
	SiteMode                string `json:"site_mode"`
	APIOutputFormat         string `json:"api_output_format"`
	EnableThirdPartyCollect bool   `json:"enable_third_party_collect"`
	ResourceAPIKeySet       bool   `json:"resource_api_key_set"`
	ResourceAPIKeyMasked    string `json:"resource_api_key_masked,omitempty"`
}

// UpdateSettingsRequest updates site + API settings.
// Empty ResourceAPIKey means "do not change".
type UpdateSettingsRequest struct {
	Site *UpdateSiteSettings `json:"site"`
	API  *UpdateAPISettings  `json:"api"`
}

// UpdateSiteSettings updates public site fields (all optional).
type UpdateSiteSettings struct {
	Name        *string `json:"name" validate:"omitempty,max=100"`
	Logo        *string `json:"logo" validate:"omitempty,max=500"`
	Copyright   *string `json:"copyright" validate:"omitempty,max=255"`
	ICP         *string `json:"icp" validate:"omitempty,max=100"`
	SEOKeywords *string `json:"seo_keywords" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// UpdateAPISettings updates API / resource station fields (all optional).
type UpdateAPISettings struct {
	SiteMode                *string `json:"site_mode" validate:"omitempty,oneof=video_site resource_site"`
	APIOutputFormat         *string `json:"api_output_format" validate:"omitempty,oneof=default apple_cms"`
	EnableThirdPartyCollect *bool   `json:"enable_third_party_collect"`
	// ResourceAPIKey empty = leave unchanged; non-empty replaces key.
	ResourceAPIKey *string `json:"resource_api_key" validate:"omitempty,max=255"`
}

// PublicSiteResponse is safe public site info for the client.
type PublicSiteResponse struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Copyright   string `json:"copyright"`
	ICP         string `json:"icp"`
	SEOKeywords string `json:"seo_keywords"`
	Description string `json:"description"`
}

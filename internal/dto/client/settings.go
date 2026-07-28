package client

// SiteSettings holds public site branding fields visible to the client.
type SiteSettings struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Copyright   string `json:"copyright"`
	ICP         string `json:"icp"`
	SEOKeywords string `json:"seo_keywords"`
	Description string `json:"description"`
}

// AdSettings holds video loading ad configuration visible to the client.
type AdSettings struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Link     string `json:"link"`
	Duration int    `json:"duration"`
	Skipable bool   `json:"skipable"`
}

// SettingsResponse is the grouped settings payload for the client.
// Only whitelisted groups are populated; non-whitelisted groups are always empty.
type SettingsResponse struct {
	Group string       `json:"group,omitempty"`
	Site  SiteSettings `json:"site"`
	Ad    AdSettings   `json:"ad"`
}

// PublicSiteResponse is safe public site info for the client.
type PublicSiteResponse struct {
	Name        string     `json:"name"`
	Logo        string     `json:"logo"`
	Copyright   string     `json:"copyright"`
	ICP         string     `json:"icp"`
	SEOKeywords string     `json:"seo_keywords"`
	Description string     `json:"description"`
	Ad          AdSettings `json:"ad"`
}

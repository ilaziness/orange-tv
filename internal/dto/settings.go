package dto

// SiteSettings holds public site branding fields.
type SiteSettings struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Copyright   string `json:"copyright"`
	ICP         string `json:"icp"`
	SEOKeywords string `json:"seo_keywords"`
	Description string `json:"description"`
}

// AdSettings holds video loading ad configuration.
type AdSettings struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Link     string `json:"link"`
	Duration int    `json:"duration"`
	Skipable bool   `json:"skipable"`
}

// FeatureSettings holds client feature toggle settings.
type FeatureSettings struct {
	LiveEnabled    bool `json:"live_enabled"`
	CommentEnabled bool `json:"comment_enabled"`
	CommentReview  bool `json:"comment_review"`
}

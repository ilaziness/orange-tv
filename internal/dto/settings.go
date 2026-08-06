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

// FeatureSettings holds client feature toggle settings.
type FeatureSettings struct {
	LiveEnabled    bool `json:"live_enabled"`
	CommentEnabled bool `json:"comment_enabled"`
	CommentReview  bool `json:"comment_review"`
	RatingEnabled  bool `json:"rating_enabled"`
}

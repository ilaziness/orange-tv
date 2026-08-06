package client

// ListAdQuery binds the scene query parameter.
type ListAdQuery struct {
	Scene string `form:"scene" binding:"required,oneof=video_loading general"`
}

// AdItem is the client advertisement item.
type AdItem struct {
	ID          uint32  `json:"id"`
	AdKey       string  `json:"ad_key"`
	Type        string  `json:"type"`
	ContentURL  string  `json:"content_url"`
	ContentCode *string `json:"content_code"`
	LinkURL     string  `json:"link_url"`
	Duration    uint32  `json:"duration"`
}

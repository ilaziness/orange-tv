package admin

// AdItem is the advertisement list item.
type AdItem struct {
	ID          uint32  `json:"id"`
	AdKey       string  `json:"ad_key"`
	Title       string  `json:"title"`
	Scene       string  `json:"scene"`
	Type        string  `json:"type"`
	ContentURL  string  `json:"content_url"`
	ContentCode *string `json:"content_code"`
	LinkURL     string  `json:"link_url"`
	Duration    uint32  `json:"duration"`
	Sort        uint32  `json:"sort"`
	Status      uint8   `json:"status"`
}

// CreateAdRequest creates an advertisement.
type CreateAdRequest struct {
	AdKey       string  `json:"ad_key" binding:"required,min=1,max=50"`
	Title       string  `json:"title" binding:"required,min=1,max=128"`
	Scene       string  `json:"scene" binding:"required,oneof=video_loading general"`
	Type        string  `json:"type" binding:"required,oneof=image video html code"`
	ContentURL  string  `json:"content_url" binding:"omitempty,max=500"`
	ContentCode *string `json:"content_code" binding:"omitempty,max=10000"`
	LinkURL     string  `json:"link_url" binding:"omitempty,max=500"`
	Duration    uint32  `json:"duration" binding:"omitempty,min=1,max=300"`
	Sort        uint32  `json:"sort" binding:"omitempty,min=0"`
	Status      *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateAdRequest updates an advertisement.
type UpdateAdRequest struct {
	AdKey       *string `json:"ad_key" binding:"omitempty,min=1,max=50"`
	Title       *string `json:"title" binding:"omitempty,min=1,max=128"`
	Scene       *string `json:"scene" binding:"omitempty,oneof=video_loading general"`
	Type        *string `json:"type" binding:"omitempty,oneof=image video html code"`
	ContentURL  *string `json:"content_url" binding:"omitempty,max=500"`
	ContentCode *string `json:"content_code" binding:"omitempty,max=10000"`
	LinkURL     *string `json:"link_url" binding:"omitempty,max=500"`
	Duration    *uint32 `json:"duration" binding:"omitempty,min=1,max=300"`
	Sort        *uint32 `json:"sort" binding:"omitempty,min=0"`
	Status      *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

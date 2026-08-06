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
	AdKey       string  `json:"ad_key" validate:"required,min=1,max=50"`
	Title       string  `json:"title" validate:"required,min=1,max=128"`
	Scene       string  `json:"scene" validate:"required,oneof=video_loading general"`
	Type        string  `json:"type" validate:"required,oneof=image video html code"`
	ContentURL  string  `json:"content_url" validate:"omitempty,max=500"`
	ContentCode *string `json:"content_code" validate:"omitempty,max=10000"`
	LinkURL     string  `json:"link_url" validate:"omitempty,max=500"`
	Duration    uint32  `json:"duration" validate:"omitempty,min=1,max=300"`
	Sort        uint32  `json:"sort" validate:"omitempty,min=0"`
	Status      *uint8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateAdRequest updates an advertisement.
type UpdateAdRequest struct {
	AdKey       *string `json:"ad_key" validate:"omitempty,min=1,max=50"`
	Title       *string `json:"title" validate:"omitempty,min=1,max=128"`
	Scene       *string `json:"scene" validate:"omitempty,oneof=video_loading general"`
	Type        *string `json:"type" validate:"omitempty,oneof=image video html code"`
	ContentURL  *string `json:"content_url" validate:"omitempty,max=500"`
	ContentCode *string `json:"content_code" validate:"omitempty,max=10000"`
	LinkURL     *string `json:"link_url" validate:"omitempty,max=500"`
	Duration    *uint32 `json:"duration" validate:"omitempty,min=1,max=300"`
	Sort        *uint32 `json:"sort" validate:"omitempty,min=0"`
	Status      *uint8  `json:"status" validate:"omitempty,oneof=0 1"`
}

package admin

// AdItem is the advertisement list item.
type AdItem struct {
	// 广告ID
	ID uint32 `json:"id"`
	// 广告标识 key
	AdKey string `json:"ad_key"`
	// 广告标题
	Title string `json:"title"`
	// 广告场景（video_loading=片头加载广告，general=通用广告）
	Scene string `json:"scene"`
	// 广告类型（image=图片，video=视频，html=HTML，code=自定义代码）
	Type string `json:"type"`
	// 内容资源地址
	ContentURL string `json:"content_url"`
	// 自定义代码内容（type=code 时使用）
	ContentCode *string `json:"content_code"`
	// 点击跳转链接
	LinkURL string `json:"link_url"`
	// 广告时长（秒）
	Duration uint32 `json:"duration"`
	// 排序权重，值越小越靠前
	Sort uint32 `json:"sort"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
}

// CreateAdRequest creates an advertisement.
type CreateAdRequest struct {
	// 广告标识 key（必填）
	AdKey string `json:"ad_key" binding:"required,min=1,max=50"`
	// 广告标题（必填）
	Title string `json:"title" binding:"required,min=1,max=128"`
	// 广告场景（必填：video_loading=片头加载广告，general=通用广告）
	Scene string `json:"scene" binding:"required,oneof=video_loading general"`
	// 广告类型（必填：image=图片，video=视频，html=HTML，code=自定义代码）
	Type string `json:"type" binding:"required,oneof=image video html code"`
	// 内容资源地址
	ContentURL string `json:"content_url" binding:"omitempty,max=500"`
	// 自定义代码内容（type=code 时使用）
	ContentCode *string `json:"content_code" binding:"omitempty,max=10000"`
	// 点击跳转链接
	LinkURL string `json:"link_url" binding:"omitempty,max=500"`
	// 广告时长（秒，1-300）
	Duration uint32 `json:"duration" binding:"omitempty,min=1,max=300"`
	// 排序权重，值越小越靠前
	Sort uint32 `json:"sort" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateAdRequest updates an advertisement.
type UpdateAdRequest struct {
	// 广告标识 key
	AdKey *string `json:"ad_key" binding:"omitempty,min=1,max=50"`
	// 广告标题
	Title *string `json:"title" binding:"omitempty,min=1,max=128"`
	// 广告场景（video_loading=片头加载广告，general=通用广告）
	Scene *string `json:"scene" binding:"omitempty,oneof=video_loading general"`
	// 广告类型（image=图片，video=视频，html=HTML，code=自定义代码）
	Type *string `json:"type" binding:"omitempty,oneof=image video html code"`
	// 内容资源地址
	ContentURL *string `json:"content_url" binding:"omitempty,max=500"`
	// 自定义代码内容（type=code 时使用）
	ContentCode *string `json:"content_code" binding:"omitempty,max=10000"`
	// 点击跳转链接
	LinkURL *string `json:"link_url" binding:"omitempty,max=500"`
	// 广告时长（秒，1-300）
	Duration *uint32 `json:"duration" binding:"omitempty,min=1,max=300"`
	// 排序权重，值越小越靠前
	Sort *uint32 `json:"sort" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

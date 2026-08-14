package client

// ListAdQuery binds the scene query parameter.
type ListAdQuery struct {
	// 广告场景（video_loading=片头加载广告，general=通用广告）
	Scene string `form:"scene" binding:"required,oneof=video_loading general"`
}

// AdItem is the client advertisement item.
type AdItem struct {
	// 广告ID
	ID uint32 `json:"id"`
	// 广告标识 key
	AdKey string `json:"ad_key"`
	// 广告类型（image=图片，video=视频，html=HTML，code=自定义代码）
	Type string `json:"type"`
	// 内容资源地址
	ContentURL string `json:"content_url"`
	// 自定义代码内容（type=code 时返回）
	ContentCode *string `json:"content_code"`
	// 点击跳转链接
	LinkURL string `json:"link_url"`
	// 广告时长（秒）
	Duration uint32 `json:"duration"`
}

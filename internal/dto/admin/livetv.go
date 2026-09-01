package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveTVListRequest filters admin livetv channel list.
type LiveTVListRequest struct {
	dto.PaginationRequest
	// 频道分类筛选
	Category string `form:"category"`
	// 状态筛选（0=禁用，1=启用）
	Status *uint8 `form:"status"`
	// 关键词搜索（频道名称）
	Keyword string `form:"keyword"`
}

// CreateLiveTVRequest creates a live channel.
type CreateLiveTVRequest struct {
	// 频道名称（必填）
	Name string `json:"name" binding:"required,min=1,max=100"`
	// 频道分类
	Category string `json:"category" binding:"omitempty,max=50"`
	// 播放流地址（必填）
	StreamURL string `json:"stream_url" binding:"required,min=1,max=1000"`
	// 频道 Logo 地址
	Logo string `json:"logo" binding:"omitempty,max=500"`
	// 频道描述
	Description string `json:"description" binding:"omitempty,max=2000"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateLiveTVRequest updates a live channel.
type UpdateLiveTVRequest struct {
	// 频道名称
	Name *string `json:"name" binding:"omitempty,min=1,max=100"`
	// 频道分类
	Category *string `json:"category" binding:"omitempty,max=50"`
	// 播放流地址
	StreamURL *string `json:"stream_url" binding:"omitempty,min=1,max=1000"`
	// 频道 Logo 地址
	Logo *string `json:"logo" binding:"omitempty,max=500"`
	// 频道描述
	Description *string `json:"description" binding:"omitempty,max=2000"`
	// 排序权重，值越小越靠前
	SortOrder *uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// LiveTVChannelItem is a live channel payload for admin (includes status, sort_order, stream_url).
type LiveTVChannelItem struct {
	// 频道ID
	ID uint32 `json:"id"`
	// 频道名称
	Name string `json:"name"`
	// 频道分类
	Category string `json:"category"`
	// 播放流地址
	StreamURL string `json:"stream_url"`
	// 频道 Logo 地址
	Logo string `json:"logo"`
	// 频道描述
	Description string `json:"description"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
}

// LiveTVSyncResult is the result of a live source synchronization.
type LiveTVSyncResult struct {
	// 总计频道数
	Total int `json:"total"`
	// 新增频道数
	Created int `json:"created"`
	// 更新频道数
	Updated int `json:"updated"`
	// 删除频道数
	Deleted int `json:"deleted"`
}

// LiveTVSyncSourceResponse returns the last saved live source URL.
type LiveTVSyncSourceResponse struct {
	// 直播源地址
	SourceURL string `json:"source_url"`
}

// LiveTVSyncRequest is the request body for live source synchronization.
type LiveTVSyncRequest struct {
	// 直播源地址（必填，合法 URL）
	SourceURL string `json:"source_url" binding:"required,url,min=1,max=2000"`
}

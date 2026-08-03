package client

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveListRequest filters public live channel list.
// 一次性返回所有在线频道，Category 仅用于前端按分类筛选时复用同一接口。
type LiveListRequest struct {
	dto.PaginationRequest
	Category string `form:"category"`
}

// LiveChannelItem is a live channel payload for client (no status, no stream_url).
type LiveChannelItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	SortOrder   uint32 `json:"sort_order"`
}

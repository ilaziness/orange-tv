package client

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveListRequest filters public live channel list.
// 一次性返回所有在线频道，Category 仅用于前端按分类筛选时复用同一接口。
type LiveListRequest struct {
	dto.PaginationRequest
	Category string `form:"category"`
}

// LivePlayRequest binds the live play URI id and optional segment URL query param.
type LivePlayRequest struct {
	ID uint32 `uri:"id" binding:"required,gt=0"`
	U  string `form:"u"`
}

// LiveChannelItem is a live channel payload for client (no status, no stream_url).
type LiveChannelItem struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	SortOrder   uint32 `json:"sort_order"`
	Format      string `json:"format"`
}

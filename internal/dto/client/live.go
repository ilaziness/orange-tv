package client

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveListRequest filters public live channel list.
// 一次性返回所有在线频道，Category 仅用于前端按分类筛选时复用同一接口。
type LiveListRequest struct {
	dto.PaginationRequest
	// 分类筛选（可选）
	Category string `form:"category"`
}

// LivePlayRequest binds the live play URI id and optional segment URL query param.
type LivePlayRequest struct {
	// 频道ID（路径参数）
	ID uint32 `uri:"id" binding:"required,gt=0"`
	// 分片 URL，用于代理具体分片（可选）
	U string `form:"u"`
}

// LiveChannelItem is a live channel payload for client (no status).
// StreamURL 仅 app/tv/desktop 端返回，web 端为空且 omitempty 隐藏。
type LiveChannelItem struct {
	// 频道ID
	ID uint32 `json:"id"`
	// 频道名称
	Name string `json:"name"`
	// 频道分类
	Category string `json:"category"`
	// 频道 Logo 地址
	Logo string `json:"logo"`
	// 频道描述
	Description string `json:"description"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order"`
	// 播放流格式（如 hls、flv 等）
	Format string `json:"format"`
	// 播放流地址，仅 app/tv/desktop 端返回，web 端为空并省略
	StreamURL string `json:"stream_url,omitempty"`
}

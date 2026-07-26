package client

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveListRequest filters public live channel list.
// 一次性返回所有在线频道，Category 仅用于前端按分类筛选时复用同一接口。
type LiveListRequest struct {
	dto.PaginationRequest
	Category string `form:"category"`
}

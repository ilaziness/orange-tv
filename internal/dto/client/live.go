package client

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveListRequest filters public live channel list.
type LiveListRequest struct {
	dto.PaginationRequest
	Category string `form:"category"`
}

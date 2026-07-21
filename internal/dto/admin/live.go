package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// LiveListRequest filters admin live channel list.
type LiveListRequest struct {
	dto.PaginationRequest
	Category string `form:"category"`
	Status   *uint8 `form:"status"`
	Keyword  string `form:"keyword"`
}

// CreateLiveRequest creates a live channel.
type CreateLiveRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Category    string `json:"category" validate:"omitempty,max=50"`
	StreamURL   string `json:"stream_url" validate:"required,min=1,max=1000"`
	Logo        string `json:"logo" validate:"omitempty,max=500"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	SortOrder   uint32 `json:"sort_order" validate:"omitempty,min=0"`
	Status      *uint8 `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateLiveRequest updates a live channel.
type UpdateLiveRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Category    *string `json:"category" validate:"omitempty,max=50"`
	StreamURL   *string `json:"stream_url" validate:"omitempty,min=1,max=1000"`
	Logo        *string `json:"logo" validate:"omitempty,max=500"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	SortOrder   *uint32 `json:"sort_order" validate:"omitempty,min=0"`
	Status      *uint8  `json:"status" validate:"omitempty,oneof=0 1"`
}

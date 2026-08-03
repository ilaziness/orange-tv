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

// LiveChannelItem is a live channel payload for admin (includes status, sort_order, stream_url).
type LiveChannelItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	StreamURL   string `json:"stream_url"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	SortOrder   uint32 `json:"sort_order"`
	Status      uint8  `json:"status"`
}

// LiveSyncResult is the result of a live source synchronization.
type LiveSyncResult struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Deleted int `json:"deleted"`
}

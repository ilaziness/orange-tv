package client

import "github.com/ilaziness/orange-tv/internal/dto"

// VideoListRequest filters public video list.
type VideoListRequest struct {
	dto.PaginationRequest
	CategoryID int64  `form:"category_id"`
	Year       int32  `form:"year"`
	Region     string `form:"region"`
	Language   string `form:"language"`
	Sort       string `form:"sort"`
}

// SearchRequest is the public search query with optional filters.
type SearchRequest struct {
	dto.PaginationRequest
	Keyword    string `form:"keyword" validate:"required,min=1,max=100"`
	CategoryID int64  `form:"category_id"`
	Year       int32  `form:"year"`
	Region     string `form:"region"`
	Language   string `form:"language"`
	Sort       string `form:"sort"`
}

// RelatedRequest loads related videos for a detail page.
type RelatedRequest struct {
	Limit int `form:"limit" validate:"omitempty,min=1,max=50"`
}

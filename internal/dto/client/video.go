package client

import "github.com/ilaziness/orange-tv/internal/dto"

// VideoListRequest filters public video list.
type VideoListRequest struct {
	dto.PaginationRequest
	CategoryID       uint64 `form:"category_id"`
	ParentCategoryID uint64 `form:"parent_category_id"`
	YearStart        uint32 `form:"year_start"`
	YearEnd          uint32 `form:"year_end"`
	Region           string `form:"region"`
	Sort             string `form:"sort"`
}

// SearchRequest is the public search query with optional filters.
type SearchRequest struct {
	dto.PaginationRequest
	Keyword          string `form:"keyword" validate:"required,min=1,max=10,search"`
	CategoryID       uint64 `form:"category_id"`
	ParentCategoryID uint64 `form:"parent_category_id"`
	YearStart        uint32 `form:"year_start"`
	YearEnd          uint32 `form:"year_end"`
	Region           string `form:"region"`
	Sort             string `form:"sort"`
}

// RelatedRequest loads related videos for a detail page.
type RelatedRequest struct {
	Limit int `form:"limit" validate:"omitempty,min=1,max=50"`
}

// EpisodeURI captures route params for single-episode play URL.
type EpisodeURI struct {
	ID        int64 `uri:"id" binding:"required,gt=0"`
	SourceID  int64 `uri:"source_id" binding:"required,gt=0"`
	EpisodeID int64 `uri:"episode_id" binding:"required,gt=0"`
}

// ClientVideoDetailResponse is the client video detail payload (no play URLs).
type ClientVideoDetailResponse struct {
	ID           uint64                       `json:"id"`
	Title        string                       `json:"title"`
	Subtitle     string                       `json:"subtitle"`
	Description  string                       `json:"description"`
	CategoryID   uint64                       `json:"category_id"`
	SerialStatus uint8                        `json:"serial_status"`
	Cover        string                       `json:"cover"`
	Poster       string                       `json:"poster"`
	Year         uint32                       `json:"year"`
	Region       string                       `json:"region"`
	Language     string                       `json:"language"`
	Duration     uint32                       `json:"duration"`
	ReleaseDate  string                       `json:"release_date,omitempty"`
	Rating       float64                      `json:"rating"`
	RatingCount  uint32                       `json:"rating_count"`
	ViewCount    uint32                       `json:"view_count"`
	Directors    []dto.NamedItem              `json:"directors"`
	Actors       []dto.NamedItem              `json:"actors"`
	Tags         []dto.NamedItem              `json:"tags"`
	Sources      []dto.VideoDetailSourceGroup `json:"sources"`
}

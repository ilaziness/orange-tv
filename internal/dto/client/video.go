package client

import "github.com/ilaziness/orange-tv/internal/dto"

// VideoListRequest filters public video list.
type VideoListRequest struct {
	dto.PaginationRequest
	CategoryID       uint32 `form:"category_id"`
	ParentCategoryID uint32 `form:"parent_category_id"`
	YearStart        uint32 `form:"year_start"`
	YearEnd          uint32 `form:"year_end"`
	Region           string `form:"region"`
	Sort             string `form:"sort"`
}

// SearchRequest is the public search query with optional filters.
type SearchRequest struct {
	dto.PaginationRequest
	Keyword          string `form:"keyword" binding:"required,min=1,max=10,search"`
	CategoryID       uint32 `form:"category_id"`
	ParentCategoryID uint32 `form:"parent_category_id"`
	YearStart        uint32 `form:"year_start"`
	YearEnd          uint32 `form:"year_end"`
	Region           string `form:"region"`
	Sort             string `form:"sort"`
}

// RelatedRequest loads related videos for a detail page.
type RelatedRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=50"`
}

// EpisodeURI captures route params for single-episode play URL.
type EpisodeURI struct {
	ID        uint32 `uri:"id" binding:"required,gt=0"`
	SourceID  uint32 `uri:"source_id" binding:"required,gt=0"`
	EpisodeID uint32 `uri:"episode_id" binding:"required,gt=0"`
}

// VideoListItem is a compact video card payload for client (no publish_status, no timestamps).
type VideoListItem struct {
	ID           uint32          `json:"id"`
	Title        string          `json:"title"`
	Subtitle     string          `json:"subtitle"`
	Cover        string          `json:"cover"`
	Poster       string          `json:"poster"`
	Year         uint32          `json:"year"`
	Region       string          `json:"region"`
	Language     string          `json:"language"`
	Rating       float64         `json:"rating"`
	CategoryID   uint32          `json:"category_id"`
	SerialStatus uint8           `json:"serial_status"`
	Duration     uint32          `json:"duration"`
	ViewCount    uint32          `json:"view_count"`
	Tags         []dto.NamedItem `json:"tags,omitempty"`
}

// VideoDetailEpisode is an episode summary without play URL (client detail API).
type VideoDetailEpisode struct {
	ID      uint32 `json:"id"`
	Episode uint32 `json:"episode"`
	Title   string `json:"title"`
}

// VideoDetailSourceGroup groups episode summaries by play source (client detail API, no URL).
type VideoDetailSourceGroup struct {
	ID       uint32               `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoDetailEpisode `json:"episodes"`
}

// VideoDetailResponse is the client video detail payload (no play URLs).
type VideoDetailResponse struct {
	ID           uint32                   `json:"id"`
	Title        string                   `json:"title"`
	Subtitle     string                   `json:"subtitle"`
	Description  string                   `json:"description"`
	CategoryID   uint32                   `json:"category_id"`
	SerialStatus uint8                    `json:"serial_status"`
	Cover        string                   `json:"cover"`
	Poster       string                   `json:"poster"`
	Year         uint32                   `json:"year"`
	Region       string                   `json:"region"`
	Language     string                   `json:"language"`
	Duration     uint32                   `json:"duration"`
	ReleaseDate  string                   `json:"release_date,omitempty"`
	Rating       float64                  `json:"rating"`
	RatingCount  uint32                   `json:"rating_count"`
	ViewCount    uint32                   `json:"view_count"`
	Directors    []dto.NamedItem          `json:"directors"`
	Actors       []dto.NamedItem          `json:"actors"`
	Tags         []dto.NamedItem          `json:"tags"`
	Sources      []VideoDetailSourceGroup `json:"sources"`
}

// PlayEpisodeResponse is the single-episode play URL response.
type PlayEpisodeResponse struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
}

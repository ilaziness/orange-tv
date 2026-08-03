package open

import (
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
)

// VideoListRequest is the query parameter for the open video list.
type VideoListRequest struct {
	shareddto.PaginationRequest
	// DataRange filters by created_at (today/last1d/last3d/last1w/last1m/all). Empty = all.
	DataRange string `form:"data_range" json:"data_range" binding:"omitempty,oneof=today last1d last3d last1w last1m all"`
	// Source filters videos by play source name (exact match). Empty = no filter.
	Source string `form:"source" json:"source" binding:"omitempty,max=10"`
}

// VideoDetailRequest is the query parameter for the open video detail (multiple ids).
type VideoDetailRequest struct {
	IDs []uint32 `form:"id" binding:"required,max=50,dive,gt=0"`
}

// VideoListItem is one compact video item in the open video list.
type VideoListItem struct {
	ID         uint32 `json:"id"`
	Title      string `json:"title"`
	CategoryID uint32 `json:"category_id"`
	CreatedAt  string `json:"created_at"`
}

// VideoSource is a play source group used in the open detail response.
type VideoSource struct {
	ID       uint32               `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// VideoSourceEpisode is one playable episode in the open detail response.
type VideoSourceEpisode struct {
	Episode uint32 `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
}

// VideoDetailItem is the full video detail payload in the open API.
type VideoDetailItem struct {
	ID          uint32        `json:"id"`
	Title       string        `json:"title"`
	Subtitle    string        `json:"subtitle"`
	Cover       string        `json:"cover"`
	CategoryID  uint32        `json:"category_id"`
	Year        uint32        `json:"year"`
	Rating      float64       `json:"rating"`
	ReleaseDate string        `json:"release_date"`
	Region      string        `json:"region"`
	Language    string        `json:"language"`
	Description string        `json:"description"`
	Directors   []string      `json:"directors"`
	Actors      []string      `json:"actors"`
	Sources     []VideoSource `json:"sources"`
	CreatedAt   string        `json:"created_at"`
}

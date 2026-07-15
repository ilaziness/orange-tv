package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CreatePlaySourceRequest creates a global play source.
type CreatePlaySourceRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	SortOrder int32  `json:"sort_order" validate:"omitempty,min=0"`
	Status    *int8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdatePlaySourceRequest updates a global play source.
type UpdatePlaySourceRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	SortOrder *int32  `json:"sort_order" validate:"omitempty,min=0"`
	Status    *int8   `json:"status" validate:"omitempty,oneof=0 1"`
}

// PlaySourceResponse is a play source payload.
type PlaySourceResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int32  `json:"sort_order"`
	Status    int8   `json:"status"`
}

// PlayEpisodeListRequest filters episode list.
type PlayEpisodeListRequest struct {
	dto.PaginationRequest
	VideoID  int64 `form:"video_id" validate:"required,min=1"`
	SourceID int64 `form:"source_id" validate:"required,min=1"`
}

// CreatePlayEpisodeRequest creates a play episode.
type CreatePlayEpisodeRequest struct {
	SourceID      int64  `json:"source_id" validate:"required,min=1"`
	VideoID       int64  `json:"video_id" validate:"required,min=1"`
	EpisodeNumber int32  `json:"episode_number" validate:"required,min=1"`
	Title         string `json:"title" validate:"omitempty,max=255"`
	PlayURL       string `json:"play_url" validate:"required,min=1,max=1000"`
	Quality       string `json:"quality" validate:"omitempty,max=50"`
	Format        string `json:"format" validate:"required,oneof=hls mp4 dash flv"`
	SortOrder     int32  `json:"sort_order" validate:"omitempty,min=0"`
	Status        *int8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdatePlayEpisodeRequest updates a play episode.
type UpdatePlayEpisodeRequest struct {
	SourceID      *int64  `json:"source_id" validate:"omitempty,min=1"`
	VideoID       *int64  `json:"video_id" validate:"omitempty,min=1"`
	EpisodeNumber *int32  `json:"episode_number" validate:"omitempty,min=1"`
	Title         *string `json:"title" validate:"omitempty,max=255"`
	PlayURL       *string `json:"play_url" validate:"omitempty,min=1,max=1000"`
	Quality       *string `json:"quality" validate:"omitempty,max=50"`
	Format        *string `json:"format" validate:"omitempty,oneof=hls mp4 dash flv"`
	SortOrder     *int32  `json:"sort_order" validate:"omitempty,min=0"`
	Status        *int8   `json:"status" validate:"omitempty,oneof=0 1"`
}

// PlayEpisodeResponse is a play episode payload.
type PlayEpisodeResponse struct {
	ID            int64  `json:"id"`
	SourceID      int64  `json:"source_id"`
	VideoID       int64  `json:"video_id"`
	EpisodeNumber int32  `json:"episode_number"`
	Title         string `json:"title"`
	PlayURL       string `json:"play_url"`
	Quality       string `json:"quality"`
	Format        string `json:"format"`
	SortOrder     int32  `json:"sort_order"`
	Status        int8   `json:"status"`
}

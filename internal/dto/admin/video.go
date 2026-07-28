package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// VideoListRequest filters admin video list.
type VideoListRequest struct {
	dto.PaginationRequest
	Keyword       string `form:"keyword"`
	CategoryID    uint64 `form:"category_id"`
	PublishStatus *uint8 `form:"publish_status"`
	Year          uint32 `form:"year"`
	Region        string `form:"region"`
	Language      string `form:"language"`
	Sort          string `form:"sort"`
	DirectorID    uint64 `form:"director_id"`
	ActorID       uint64 `form:"actor_id"`
	TagID         uint64 `form:"tag_id"`
}

// CreateVideoRequest creates a video with associations.
type CreateVideoRequest struct {
	Title         string            `json:"title" validate:"required,min=1,max=255"`
	Subtitle      string            `json:"subtitle" validate:"omitempty,max=255"`
	Description   string            `json:"description" validate:"omitempty,max=10000"`
	CategoryID    uint64            `json:"category_id" validate:"required,min=1"`
	PublishStatus *uint8            `json:"publish_status" validate:"omitempty,oneof=0 1"`
	SerialStatus  *uint8            `json:"serial_status" validate:"omitempty,oneof=1 2 3"`
	CoverImage    string            `json:"cover_image" validate:"omitempty,max=500"`
	PosterImage   string            `json:"poster_image" validate:"omitempty,max=500"`
	Year          uint32            `json:"year" validate:"omitempty,min=0,max=9999"`
	Region        string            `json:"region" validate:"omitempty,max=50"`
	Rating        float64           `json:"rating" validate:"omitempty,min=0,max=10"`
	Duration      uint32            `json:"duration" validate:"omitempty,min=0"`
	Language      string            `json:"language" validate:"omitempty,max=50"`
	ReleaseDate   string            `json:"release_date" validate:"omitempty"`
	DirectorIDs   []uint64          `json:"director_ids" validate:"omitempty,dive,min=1"`
	Actors        []VideoActorInput `json:"actors" validate:"omitempty,dive"`
	TagIDs        []uint64          `json:"tag_ids" validate:"omitempty,dive,min=1"`
}

// UpdateVideoRequest updates a video and optional associations.
type UpdateVideoRequest struct {
	Title         *string            `json:"title" validate:"omitempty,min=1,max=255"`
	Subtitle      *string            `json:"subtitle" validate:"omitempty,max=255"`
	Description   *string            `json:"description" validate:"omitempty,max=10000"`
	CategoryID    *uint64            `json:"category_id" validate:"omitempty,min=1"`
	PublishStatus *uint8             `json:"publish_status" validate:"omitempty,oneof=0 1"`
	SerialStatus  *uint8             `json:"serial_status" validate:"omitempty,oneof=1 2 3"`
	CoverImage    *string            `json:"cover_image" validate:"omitempty,max=500"`
	PosterImage   *string            `json:"poster_image" validate:"omitempty,max=500"`
	Year          *uint32            `json:"year" validate:"omitempty,min=0,max=9999"`
	Region        *string            `json:"region" validate:"omitempty,max=50"`
	Rating        *float64           `json:"rating" validate:"omitempty,min=0,max=10"`
	Duration      *uint32            `json:"duration" validate:"omitempty,min=0"`
	Language      *string            `json:"language" validate:"omitempty,max=50"`
	ReleaseDate   *string            `json:"release_date" validate:"omitempty"`
	DirectorIDs   *[]uint64          `json:"director_ids" validate:"omitempty,dive,min=1"`
	Actors        *[]VideoActorInput `json:"actors" validate:"omitempty,dive"`
	TagIDs        *[]uint64          `json:"tag_ids" validate:"omitempty,dive,min=1"`
}

// VideoActorInput binds actor to video.
type VideoActorInput struct {
	ActorID uint64 `json:"actor_id" validate:"required,min=1"`
}

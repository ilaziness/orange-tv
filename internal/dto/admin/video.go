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
	Duration      uint32            `json:"duration" validate:"omitempty,min=0"`
	Language      string            `json:"language" validate:"omitempty,max=50"`
	ReleaseDate   string            `json:"release_date" validate:"omitempty,max=64"`
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
	Duration      *uint32            `json:"duration" validate:"omitempty,min=0"`
	Language      *string            `json:"language" validate:"omitempty,max=50"`
	ReleaseDate   *string            `json:"release_date" validate:"omitempty,max=64"`
	DirectorIDs   *[]uint64          `json:"director_ids" validate:"omitempty,dive,min=1"`
	Actors        *[]VideoActorInput `json:"actors" validate:"omitempty,dive"`
	TagIDs        *[]uint64          `json:"tag_ids" validate:"omitempty,dive,min=1"`
}

// VideoActorInput binds actor to video.
type VideoActorInput struct {
	ActorID uint64 `json:"actor_id" validate:"required,min=1"`
}

// VideoListItem is a compact video card payload for admin (includes publish_status, timestamps).
type VideoListItem struct {
	ID            uint64          `json:"id"`
	Title         string          `json:"title"`
	Subtitle      string          `json:"subtitle"`
	Cover         string          `json:"cover"`
	Poster        string          `json:"poster"`
	Year          uint32          `json:"year"`
	Region        string          `json:"region"`
	Language      string          `json:"language"`
	Rating        float64         `json:"rating"`
	CategoryID    uint64          `json:"category_id"`
	CategoryName  string          `json:"category_name,omitempty"`
	PublishStatus uint8           `json:"publish_status"`
	SerialStatus  uint8           `json:"serial_status"`
	Duration      uint32          `json:"duration"`
	ViewCount     uint32          `json:"view_count"`
	Tags          []dto.NamedItem `json:"tags,omitempty"`
	CreatedAt     string          `json:"created_at,omitempty"`
	UpdatedAt     string          `json:"updated_at,omitempty"`
}

// VideoSourceEpisode is one playable episode under a source for admin (includes URL).
type VideoSourceEpisode struct {
	ID      uint64 `json:"id"`
	Episode uint32 `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
	Status  uint8  `json:"status"`
}

// VideoSourceGroup groups episodes by play source for admin (includes URL).
type VideoSourceGroup struct {
	ID       uint64               `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// VideoDetailResponse is a full video detail payload for admin (includes publish_status & play URLs).
type VideoDetailResponse struct {
	ID            uint64             `json:"id"`
	Title         string             `json:"title"`
	Subtitle      string             `json:"subtitle"`
	Description   string             `json:"description"`
	CategoryID    uint64             `json:"category_id"`
	PublishStatus uint8              `json:"publish_status"`
	SerialStatus  uint8              `json:"serial_status"`
	Cover         string             `json:"cover"`
	Poster        string             `json:"poster"`
	Year          uint32             `json:"year"`
	Region        string             `json:"region"`
	Language      string             `json:"language"`
	Duration      uint32             `json:"duration"`
	ReleaseDate   string             `json:"release_date,omitempty"`
	Rating        float64            `json:"rating"`
	ViewCount     uint32             `json:"view_count"`
	Directors     []dto.NamedItem    `json:"directors"`
	Actors        []dto.NamedItem    `json:"actors"`
	Tags          []dto.NamedItem    `json:"tags"`
	Sources       []VideoSourceGroup `json:"sources"`
}

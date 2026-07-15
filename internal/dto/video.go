package dto

// AdminVideoListRequest filters admin video list.
type AdminVideoListRequest struct {
	PaginationRequest
	Keyword       string `form:"keyword"`
	CategoryID    int64  `form:"category_id"`
	PublishStatus *int8  `form:"publish_status"`
	Year          int32  `form:"year"`
	Region        string `form:"region"`
	Language      string `form:"language"`
	Sort          string `form:"sort"`
}

// ClientVideoListRequest filters public video list.
type ClientVideoListRequest struct {
	PaginationRequest
	CategoryID int64  `form:"category_id"`
	Year       int32  `form:"year"`
	Region     string `form:"region"`
	Language   string `form:"language"`
	Sort       string `form:"sort"`
}

// SearchRequest is the public search query.
type SearchRequest struct {
	PaginationRequest
	Keyword string `form:"keyword" validate:"required,min=1,max=100"`
}

// CreateVideoRequest creates a video with associations.
type CreateVideoRequest struct {
	Title         string              `json:"title" validate:"required,min=1,max=255"`
	Subtitle      string              `json:"subtitle" validate:"omitempty,max=255"`
	Description   string              `json:"description" validate:"omitempty,max=10000"`
	CategoryID    int64               `json:"category_id" validate:"required,min=1"`
	PublishStatus *int8               `json:"publish_status" validate:"omitempty,oneof=0 1"`
	SerialStatus  *int8               `json:"serial_status" validate:"omitempty,oneof=1 2 3"`
	CoverImage    string              `json:"cover_image" validate:"omitempty,max=500"`
	PosterImage   string              `json:"poster_image" validate:"omitempty,max=500"`
	Year          int32               `json:"year" validate:"omitempty,min=0,max=9999"`
	Region        string              `json:"region" validate:"omitempty,max=50"`
	Rating        float64             `json:"rating" validate:"omitempty,min=0,max=10"`
	Duration      int32               `json:"duration" validate:"omitempty,min=0"`
	Language      string              `json:"language" validate:"omitempty,max=50"`
	ReleaseDate   string              `json:"release_date" validate:"omitempty"`
	DirectorIDs   []int64             `json:"director_ids" validate:"omitempty,dive,min=1"`
	Actors        []VideoActorInput   `json:"actors" validate:"omitempty,dive"`
	TagIDs        []int64             `json:"tag_ids" validate:"omitempty,dive,min=1"`
}

// UpdateVideoRequest updates a video and optional associations.
type UpdateVideoRequest struct {
	Title         *string            `json:"title" validate:"omitempty,min=1,max=255"`
	Subtitle      *string            `json:"subtitle" validate:"omitempty,max=255"`
	Description   *string            `json:"description" validate:"omitempty,max=10000"`
	CategoryID    *int64             `json:"category_id" validate:"omitempty,min=1"`
	PublishStatus *int8              `json:"publish_status" validate:"omitempty,oneof=0 1"`
	SerialStatus  *int8              `json:"serial_status" validate:"omitempty,oneof=1 2 3"`
	CoverImage    *string            `json:"cover_image" validate:"omitempty,max=500"`
	PosterImage   *string            `json:"poster_image" validate:"omitempty,max=500"`
	Year          *int32             `json:"year" validate:"omitempty,min=0,max=9999"`
	Region        *string            `json:"region" validate:"omitempty,max=50"`
	Rating        *float64           `json:"rating" validate:"omitempty,min=0,max=10"`
	Duration      *int32             `json:"duration" validate:"omitempty,min=0"`
	Language      *string            `json:"language" validate:"omitempty,max=50"`
	ReleaseDate   *string            `json:"release_date" validate:"omitempty"`
	DirectorIDs   *[]int64           `json:"director_ids" validate:"omitempty,dive,min=1"`
	Actors        *[]VideoActorInput `json:"actors" validate:"omitempty,dive"`
	TagIDs        *[]int64           `json:"tag_ids" validate:"omitempty,dive,min=1"`
}

// VideoActorInput binds actor and optional role.
type VideoActorInput struct {
	ActorID int64  `json:"actor_id" validate:"required,min=1"`
	Role    string `json:"role" validate:"omitempty,max=100"`
}

// NamedItem is a simple id/name pair.
type NamedItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ActorItem is actor info with role.
type ActorItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// VideoListItem is a compact video card payload.
type VideoListItem struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Subtitle      string  `json:"subtitle"`
	Cover         string  `json:"cover"`
	Poster        string  `json:"poster"`
	Year          int32   `json:"year"`
	Region        string  `json:"region"`
	Language      string  `json:"language"`
	Rating        float64 `json:"rating"`
	CategoryID    int64   `json:"category_id"`
	PublishStatus int8    `json:"publish_status,omitempty"`
	SerialStatus  int8    `json:"serial_status"`
	Duration      int32   `json:"duration"`
	ViewCount     int32   `json:"view_count"`
}

// VideoSourceEpisode is one playable episode under a source.
type VideoSourceEpisode struct {
	Episode int32  `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
}

// VideoSourceGroup groups episodes by play source.
type VideoSourceGroup struct {
	ID       int64                `json:"id"`
	Name     string               `json:"name"`
	Episodes []VideoSourceEpisode `json:"episodes"`
}

// VideoDetailResponse is a full video detail payload.
type VideoDetailResponse struct {
	ID            int64              `json:"id"`
	Title         string             `json:"title"`
	Subtitle      string             `json:"subtitle"`
	Description   string             `json:"description"`
	CategoryID    int64              `json:"category_id"`
	PublishStatus int8               `json:"publish_status,omitempty"`
	SerialStatus  int8               `json:"serial_status"`
	Cover         string             `json:"cover"`
	Poster        string             `json:"poster"`
	Year          int32              `json:"year"`
	Region        string             `json:"region"`
	Language      string             `json:"language"`
	Duration      int32              `json:"duration"`
	ReleaseDate   string             `json:"release_date,omitempty"`
	Rating        float64            `json:"rating"`
	ViewCount     int32              `json:"view_count"`
	Directors     []NamedItem        `json:"directors"`
	Actors        []ActorItem        `json:"actors"`
	Tags          []NamedItem        `json:"tags"`
	Sources       []VideoSourceGroup `json:"sources"`
}

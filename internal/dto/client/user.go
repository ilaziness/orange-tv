package client

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== User auth (C5) =====

// RegisterRequest is the user registration payload.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Email    string `json:"email" validate:"omitempty,email,max=128"`
}

// LoginRequest is the user login payload.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// LoginResponse is returned after successful user login.
type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	User        *Profile `json:"user"`
}

// Profile is the authenticated user public profile.
type Profile struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Status   uint8  `json:"status"`
}

// ===== Favorites (C6) =====

// FavoriteListRequest filters user favorites.
type FavoriteListRequest struct {
	dto.PaginationRequest
}

// FavoriteItem is the favorite list item.
type FavoriteItem struct {
	VideoID      uint64  `json:"video_id"`
	Title        string  `json:"title"`
	Cover        string  `json:"cover"`
	Year         uint32  `json:"year"`
	Rating       float64 `json:"rating"`
	CategoryName string  `json:"category_name"`
	CreatedAt    string  `json:"created_at"`
}

// FavoriteCheckResult is the favorite check result.
type FavoriteCheckResult struct {
	Favorited bool `json:"favorited"`
}

// ===== Play history (C6) =====

// HistoryListRequest filters user play history.
type HistoryListRequest struct {
	dto.PaginationRequest
}

// HistoryItem is the play history list item.
type HistoryItem struct {
	VideoID      uint64 `json:"video_id"`
	Title        string `json:"title"`
	Cover        string `json:"cover"`
	Year         string `json:"year"`
	CategoryName string `json:"category_name"`
	PlaySourceID uint64 `json:"play_source_id"`
	EpisodeID    uint64 `json:"episode_id"`
	Progress     uint32 `json:"progress"`
	Duration     uint32 `json:"duration"`
	LastPlayedAt string `json:"last_played_at"`
}

// UpsertHistoryRequest upserts play progress.
type UpsertHistoryRequest struct {
	VideoID      uint64 `json:"video_id" validate:"required,min=1"`
	PlaySourceID uint64 `json:"play_source_id" validate:"omitempty,min=1"`
	EpisodeID    uint64 `json:"episode_id" validate:"omitempty,min=1"`
	Progress     uint32 `json:"progress" validate:"omitempty,min=0"`
	Duration     uint32 `json:"duration" validate:"omitempty,min=0"`
}

// ===== Comments (C6) =====

// CommentListRequest filters video comments.
type CommentListRequest struct {
	dto.PaginationRequest
}

// CommentItem is the comment list item.
type CommentItem struct {
	ID           uint64 `json:"id"`
	VideoID      uint64 `json:"video_id"`
	UserID       uint64 `json:"user_id"`
	Username     string `json:"username"`
	Avatar       string `json:"avatar"`
	ParentID     uint64 `json:"parent_id"`
	Content      string `json:"content"`
	LikeCount    uint32 `json:"like_count"`
	DislikeCount uint32 `json:"dislike_count"`
	CreatedAt    string `json:"created_at"`
}

// CreateCommentRequest creates a comment.
type CreateCommentRequest struct {
	VideoID  uint64 `json:"video_id" validate:"required,min=1"`
	ParentID uint64 `json:"parent_id" validate:"omitempty,min=1"`
	Content  string `json:"content" validate:"required,min=1,max=500"`
}

// ===== Banner (C1) =====

// BannerItem is the client banner item.
type BannerItem struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	Cover   string `json:"cover"`
	Link    string `json:"link"`
	VideoID uint64 `json:"video_id"`
}

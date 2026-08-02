package client

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== User auth (C5) =====

// RegisterRequest is the user registration payload.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=2,max=15,alphanum"`
	Password string `json:"password" validate:"required,min=5,max=30"`
	Email    string `json:"email" validate:"omitempty,email,max=128"`
}

// LoginRequest is the user login payload.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=2,max=15,alphanum"`
	Password string `json:"password" validate:"required,min=5,max=30"`
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
	StrID    string `json:"str_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Status   uint8  `json:"status"`
}

// UpdateProfileRequest updates the current user's profile.
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" validate:"omitempty,min=3,max=15"`
	Email    string `json:"email" validate:"omitempty,email,max=20"`
	Avatar   string `json:"avatar" validate:"omitempty,url,max=120"`
}

// ChangePasswordRequest changes the current user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=5,max=30"`
	NewPassword     string `json:"new_password" validate:"required,min=5,max=30"`
}

// LoginHistoryListRequest filters user login history.
type LoginHistoryListRequest struct {
	dto.PaginationRequest
}

// LoginHistoryItem is a single user login log entry.
type LoginHistoryItem struct {
	ID        uint64 `json:"id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Status    uint8  `json:"status"`
	CreatedAt string `json:"created_at"`
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
	ID           uint64         `json:"id"`
	VideoID      uint64         `json:"video_id"`
	UserID       uint64         `json:"user_id"`
	Username     string         `json:"username"`
	Avatar       string         `json:"avatar"`
	ParentID     uint64         `json:"parent_id"`
	Content      string         `json:"content"`
	LikeCount    uint32         `json:"like_count"`
	DislikeCount uint32         `json:"dislike_count"`
	MyVote       int8           `json:"my_vote"` // 1=顶 -1=踩 0=未投票
	ReplyCount   int            `json:"reply_count"`
	Replies      []*CommentItem `json:"replies"`
	CreatedAt    string         `json:"created_at"`
}

// CreateCommentRequest creates a comment.
type CreateCommentRequest struct {
	VideoID  uint64 `json:"video_id" validate:"required,min=1"`
	ParentID uint64 `json:"parent_id" validate:"omitempty,min=1"`
	Content  string `json:"content" validate:"required,min=1,max=200"`
}

// VoteCommentRequest votes on a comment.
type VoteCommentRequest struct {
	Action string `json:"action" validate:"required,oneof=like dislike cancel"`
}

// VoteCommentResult is returned after a comment vote.
type VoteCommentResult struct {
	LikeCount    uint32 `json:"like_count"`
	DislikeCount uint32 `json:"dislike_count"`
	MyVote       int8   `json:"my_vote"`
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

// ===== Ratings (C6) =====

// RateVideoRequest is the user video rating payload.
type RateVideoRequest struct {
	Score float64 `json:"score" validate:"required"`
}

// RatingResult is the rating response containing the user's score and video stats.
type RatingResult struct {
	MyScore     float64 `json:"my_score"`     // 当前用户评分，0 表示未评/未登录
	Rating      float64 `json:"rating"`       // 视频平均分
	RatingCount uint32  `json:"rating_count"` // 评分人数
}

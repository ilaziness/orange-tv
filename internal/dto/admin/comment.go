package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CommentListRequest filters admin comment list.
type CommentListRequest struct {
	dto.PaginationRequest
	Keyword string  `form:"keyword"`
	Status  *uint8  `form:"status" validate:"omitempty,oneof=0 1"`
	VideoID *uint32 `form:"video_id" validate:"omitempty,min=1"`
}

// UpdateCommentStatusRequest updates comment audit status.
type UpdateCommentStatusRequest struct {
	Status uint8 `json:"status" validate:"required,oneof=0 1"`
}

// CommentListItem is one admin comment list row.
type CommentListItem struct {
	ID           uint32 `json:"id"`
	VideoID      uint32 `json:"video_id"`
	VideoTitle   string `json:"video_title"`
	Content      string `json:"content"`
	UserID       uint32 `json:"user_id"`
	Username     string `json:"username"`
	Status       uint8  `json:"status"`
	LikeCount    uint32 `json:"like_count"`
	DislikeCount uint32 `json:"dislike_count"`
	ParentID     uint32 `json:"parent_id"`
	CreatedAt    string `json:"created_at"`
}

// CommentParentItem is one ancestor in a parent comment chain.
type CommentParentItem struct {
	ID        uint32 `json:"id"`
	UserID    uint32 `json:"user_id"`
	Username  string `json:"username"`
	ParentID  uint32 `json:"parent_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

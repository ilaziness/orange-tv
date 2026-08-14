package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CommentListRequest filters admin comment list.
type CommentListRequest struct {
	dto.PaginationRequest
	// 关键词搜索（评论内容/用户名）
	Keyword string `form:"keyword"`
	// 审核状态筛选（0=待审核/未通过，1=已通过）
	Status *uint8 `form:"status" binding:"omitempty,oneof=0 1"`
	// 视频ID筛选
	VideoID *uint32 `form:"video_id" binding:"omitempty,min=1"`
}

// UpdateCommentStatusRequest updates comment audit status.
type UpdateCommentStatusRequest struct {
	// 审核状态（必填：0=未通过，1=通过）
	Status uint8 `json:"status" binding:"required,oneof=0 1"`
}

// CommentListItem is one admin comment list row.
type CommentListItem struct {
	// 评论ID
	ID uint32 `json:"id"`
	// 视频ID
	VideoID uint32 `json:"video_id"`
	// 视频标题
	VideoTitle string `json:"video_title"`
	// 评论内容
	Content string `json:"content"`
	// 用户ID
	UserID uint32 `json:"user_id"`
	// 用户名
	Username string `json:"username"`
	// 审核状态（0=未通过，1=通过）
	Status uint8 `json:"status"`
	// 点赞数
	LikeCount uint32 `json:"like_count"`
	// 踩数
	DislikeCount uint32 `json:"dislike_count"`
	// 父评论ID，0 表示顶级评论
	ParentID uint32 `json:"parent_id"`
	// 评论时间
	CreatedAt string `json:"created_at"`
}

// CommentParentItem is one ancestor in a parent comment chain.
type CommentParentItem struct {
	// 评论ID
	ID uint32 `json:"id"`
	// 用户ID
	UserID uint32 `json:"user_id"`
	// 用户名
	Username string `json:"username"`
	// 父评论ID
	ParentID uint32 `json:"parent_id"`
	// 评论内容
	Content string `json:"content"`
	// 评论时间
	CreatedAt string `json:"created_at"`
}

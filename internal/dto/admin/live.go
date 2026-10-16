package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== Live categories =====

// LiveCategoryItem is an admin live category row.
type LiveCategoryItem struct {
	ID         uint32 `json:"id"`
	ParentID   uint32 `json:"parent_id"`
	ParentName string `json:"parent_name"`
	Name       string `json:"name"`
	Sort       uint32 `json:"sort"`
	Status     uint8  `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// LiveCategoryListRequest filters admin live categories.
type LiveCategoryListRequest struct {
	dto.PaginationRequest
	// 状态（0=禁用，1=启用）
	Status *uint8 `form:"status" binding:"omitempty,oneof=0 1"`
	// 父分类ID；传 0 只查一级。不传则不按层级过滤
	ParentID *uint32 `form:"parent_id" binding:"omitempty,min=0"`
	Keyword  string  `form:"keyword" binding:"omitempty,max=50"`
}

// CreateLiveCategoryRequest creates a live category.
// ParentID is fixed at create time and cannot be changed later.
type CreateLiveCategoryRequest struct {
	// 父分类ID，0=一级；大于0时必须是启用的一级分类
	ParentID uint32 `json:"parent_id" binding:"omitempty,min=0"`
	Name     string `json:"name" binding:"required,min=1,max=50"`
	Sort     uint32 `json:"sort" binding:"omitempty,min=0"`
	Status   *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateLiveCategoryRequest updates a live category (parent_id is immutable).
type UpdateLiveCategoryRequest struct {
	Name   *string `json:"name" binding:"omitempty,min=1,max=50"`
	Sort   *uint32 `json:"sort" binding:"omitempty,min=0"`
	Status *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// ===== Streamer applies =====

// LiveApplyListRequest filters admin streamer applies.
type LiveApplyListRequest struct {
	dto.PaginationRequest
	// 状态（0=待审，1=通过，2=拒绝）
	Status  *uint8  `form:"status" binding:"omitempty,oneof=0 1 2"`
	UserID  *uint32 `form:"user_id" binding:"omitempty,min=1"`
	Keyword string  `form:"keyword" binding:"omitempty,max=100"`
}

// LiveApplyItem is an admin streamer apply row.
type LiveApplyItem struct {
	ID            uint32 `json:"id"`
	UserID        uint32 `json:"user_id"`
	UserStrID     string `json:"user_str_id"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	Avatar        string `json:"avatar"`
	Reason        string `json:"reason"`
	Contact       string `json:"contact"`
	Status        uint8  `json:"status"`
	ReviewAdminID uint32 `json:"review_admin_id"`
	ReviewRemark  string `json:"review_remark"`
	ReviewedAt    string `json:"reviewed_at,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// ReviewLiveApplyRequest is the body for approve/reject.
type ReviewLiveApplyRequest struct {
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// ===== Streamers =====

// LiveStreamerListRequest filters admin streamers.
type LiveStreamerListRequest struct {
	dto.PaginationRequest
	Keyword string `form:"keyword" binding:"omitempty,max=100"`
	// 房间状态（0=空闲，1=直播中，2=禁播）
	RoomStatus *uint8 `form:"room_status" binding:"omitempty,oneof=0 1 2"`
}

// LiveStreamerItem is an admin streamer list row (users.is_streamer=1 + room).
type LiveStreamerItem struct {
	UserID          uint32 `json:"user_id"`
	UserStrID       string `json:"user_str_id"`
	Nickname        string `json:"nickname"`
	Email           string `json:"email"`
	Avatar          string `json:"avatar"`
	StreamerAt      string `json:"streamer_at,omitempty"`
	RoomID          uint32 `json:"room_id"`
	RoomTitle       string `json:"room_title"`
	RoomStatus      uint8  `json:"room_status"`
	CategoryID      uint32 `json:"category_id"`
	CategoryName    string `json:"category_name"`
	SubCategoryID   uint32 `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
}

// ===== Live rooms =====

// LiveRoomListRequest filters admin live rooms.
type LiveRoomListRequest struct {
	dto.PaginationRequest
	Status     *uint8  `form:"status" binding:"omitempty,oneof=0 1 2"`
	CategoryID *uint32 `form:"category_id" binding:"omitempty,min=1"`
	UserID     *uint32 `form:"user_id" binding:"omitempty,min=1"`
	Keyword    string  `form:"keyword" binding:"omitempty,max=100"`
}

// LiveRoomItem is an admin live room row.
type LiveRoomItem struct {
	ID               uint32 `json:"id"`
	UserID           uint32 `json:"user_id"`
	UserStrID        string `json:"user_str_id"`
	Nickname         string `json:"nickname"`
	Avatar           string `json:"avatar"`
	CategoryID       uint32 `json:"category_id"`
	CategoryName     string `json:"category_name"`
	SubCategoryID    uint32 `json:"sub_category_id"`
	SubCategoryName  string `json:"sub_category_name"`
	Title            string `json:"title"`
	Cover            string `json:"cover"`
	Notice           string `json:"notice"`
	StreamKey        string `json:"stream_key"`
	Status           uint8  `json:"status"`
	CurrentSessionID uint32 `json:"current_session_id,omitempty"`
	StartedAt        string `json:"started_at,omitempty"`
	EndedAt          string `json:"ended_at,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// ===== Live sessions =====

// LiveSessionListRequest filters admin live sessions.
type LiveSessionListRequest struct {
	dto.PaginationRequest
	RoomID *uint32 `form:"room_id" binding:"omitempty,min=1"`
	UserID *uint32 `form:"user_id" binding:"omitempty,min=1"`
}

// LiveSessionItem is an admin live session row.
type LiveSessionItem struct {
	ID               uint32 `json:"id"`
	RoomID           uint32 `json:"room_id"`
	UserID           uint32 `json:"user_id"`
	Title            string `json:"title"`
	CategoryID       uint32 `json:"category_id"`
	CategoryName     string `json:"category_name"`
	SubCategoryID    uint32 `json:"sub_category_id"`
	SubCategoryName  string `json:"sub_category_name"`
	StreamerNickname string `json:"streamer_nickname"`
	StreamerAvatar   string `json:"streamer_avatar"`
	StartedAt        string `json:"started_at"`
	EndedAt          string `json:"ended_at,omitempty"`
	EndReason        string `json:"end_reason"`
}

// UserIDURI binds :userId path param.
type UserIDURI struct {
	UserID uint32 `uri:"userId" binding:"required,gt=0"`
}

package client

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== Applies =====

// CreateLiveApplyRequest submits a streamer application.
type CreateLiveApplyRequest struct {
	Reason  string `json:"reason" binding:"required,min=1,max=500"`
	Contact string `json:"contact" binding:"required,min=1,max=100"`
}

// LiveApplyItem is the current user's latest apply status.
type LiveApplyItem struct {
	ID           uint32 `json:"id"`
	Reason       string `json:"reason"`
	Contact      string `json:"contact"`
	Status       uint8  `json:"status"`
	ReviewRemark string `json:"review_remark"`
	ReviewedAt   string `json:"reviewed_at,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// ===== Categories =====

// LiveCategoryItem is a public enabled live category node (two-level tree).
type LiveCategoryItem struct {
	ID       uint32             `json:"id"`
	ParentID uint32             `json:"parent_id"`
	Name     string             `json:"name"`
	Sort     uint32             `json:"sort"`
	Children []LiveCategoryItem `json:"children,omitempty"`
}

// ===== Rooms =====

// LiveRoomListRequest filters public living rooms.
// CategoryID may be L1 or L2; the server resolves filter by parent_id.
type LiveRoomListRequest struct {
	dto.PaginationRequest
	CategoryID *uint32 `form:"category_id" binding:"omitempty,min=1"`
}

// LiveRoomListItem is a public living room card.
type LiveRoomListItem struct {
	ID               uint32 `json:"id"`
	Title            string `json:"title"`
	Cover            string `json:"cover"`
	CategoryID       uint32 `json:"category_id"`
	CategoryName     string `json:"category_name"`
	SubCategoryID    uint32 `json:"sub_category_id"`
	SubCategoryName  string `json:"sub_category_name"`
	StreamerUserID   uint32 `json:"streamer_user_id"`
	StreamerStrID    string `json:"streamer_str_id"`
	StreamerNickname string `json:"streamer_nickname"`
	StreamerAvatar   string `json:"streamer_avatar"`
	StartedAt        string `json:"started_at,omitempty"`
}

// LiveRoomDetail is the public room detail (never includes stream_key / push_server).
type LiveRoomDetail struct {
	ID               uint32 `json:"id"`
	Title            string `json:"title"`
	Cover            string `json:"cover"`
	Notice           string `json:"notice"`
	CategoryID       uint32 `json:"category_id"`
	CategoryName     string `json:"category_name"`
	SubCategoryID    uint32 `json:"sub_category_id"`
	SubCategoryName  string `json:"sub_category_name"`
	Status           uint8  `json:"status"`
	StreamerUserID   uint32 `json:"streamer_user_id"`
	StreamerStrID    string `json:"streamer_str_id"`
	StreamerNickname string `json:"streamer_nickname"`
	StreamerAvatar   string `json:"streamer_avatar"`
	PlayHLSURL       string `json:"play_hls_url"`
	PlayFLVURL       string `json:"play_flv_url"`
	// 播放地址过期时间（Unix 秒）。未签名时为 0
	PlayURLExpiresAt int64  `json:"play_url_expires_at,omitempty"`
	StartedAt        string `json:"started_at,omitempty"`
}

// MyLiveRoom is the streamer's own room (may include push credentials).
type MyLiveRoom struct {
	ID              uint32 `json:"id"`
	CategoryID      uint32 `json:"category_id"`
	CategoryName    string `json:"category_name"`
	SubCategoryID   uint32 `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
	Title           string `json:"title"`
	Cover           string `json:"cover"`
	Notice          string `json:"notice"`
	Status          uint8  `json:"status"`
	StreamKey       string `json:"stream_key"`
	PushServer      string `json:"push_server"`
	// 推流签名过期时间（Unix 秒）。未签名时为 0
	PushExpiresAt    int64  `json:"push_expires_at,omitempty"`
	CurrentSessionID uint32 `json:"current_session_id,omitempty"`
	StartedAt        string `json:"started_at,omitempty"`
	EndedAt          string `json:"ended_at,omitempty"`
}

// UpdateMyLiveRoomRequest updates the streamer's room profile.
// When category fields are present, submit category_id + sub_category_id together (use 0 when no L2).
type UpdateMyLiveRoomRequest struct {
	Title         *string `json:"title" binding:"omitempty,min=1,max=100"`
	Cover         *string `json:"cover" binding:"omitempty,max=500"`
	Notice        *string `json:"notice" binding:"omitempty,max=500"`
	CategoryID    *uint32 `json:"category_id" binding:"omitempty,min=1"`
	SubCategoryID *uint32 `json:"sub_category_id" binding:"omitempty,min=0"`
}

// StartLiveRequest starts a broadcast (optional category override; submit L1+L2 together).
type StartLiveRequest struct {
	CategoryID    *uint32 `json:"category_id" binding:"omitempty,min=1"`
	SubCategoryID *uint32 `json:"sub_category_id" binding:"omitempty,min=0"`
	Title         *string `json:"title" binding:"omitempty,min=1,max=100"`
}

// StartLiveResponse returns push credentials after start (or idempotent living).
type StartLiveResponse struct {
	RoomID     uint32 `json:"room_id"`
	SessionID  uint32 `json:"session_id"`
	Status     uint8  `json:"status"`
	PushServer string `json:"push_server"`
	StreamKey  string `json:"stream_key"`
	// 推流签名过期时间（Unix 秒）。未签名时为 0
	PushExpiresAt int64  `json:"push_expires_at,omitempty"`
	StartedAt     string `json:"started_at"`
}

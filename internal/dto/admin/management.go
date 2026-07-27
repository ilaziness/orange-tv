package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== Dashboard / Stats (A1) =====

// DashboardResponse is the admin dashboard overview.
type DashboardResponse struct {
	TotalVideos     int64 `json:"total_videos"`
	TodayVideos     int64 `json:"today_videos"`
	OnlineVideos    int64 `json:"online_videos"`
	OfflineVideos   int64 `json:"offline_videos"`
	TotalCategories int64 `json:"total_categories"`
	TotalAdmins     int64 `json:"total_admins"`
	TotalUsers      int64 `json:"total_users"`
	OnlineCount     int64 `json:"online_count"`
	TodayPV         int64 `json:"today_pv"`
	TodayUV         int64 `json:"today_uv"`
	CollectRunning  int64 `json:"collect_running"`
}

// ===== Batch video ops (A2) =====

// BatchVideoRequest is the batch video operation payload.
type BatchVideoRequest struct {
	IDs    []uint64 `json:"ids" validate:"required,min=1,max=500"`
	Status *uint8   `json:"status" validate:"omitempty,oneof=0 1"`
}

// BatchVideoResponse is the batch operation result.
type BatchVideoResponse struct {
	Affected int `json:"affected"`
}

// ===== Admin CRUD (A3) =====

// AdminListRequest filters admin list.
type AdminListRequest struct {
	dto.PaginationRequest
	Keyword string  `form:"keyword"`
	Status  *uint8  `form:"status"`
	GroupID *uint64 `form:"group_id"`
}

// AdminItem is the admin list item.
type AdminItem struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
	GroupID     uint64 `json:"group_id"`
	GroupName   string `json:"group_name"`
	Status      uint8  `json:"status"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

// CreateAdminRequest creates a new admin.
type CreateAdminRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Email    string `json:"email" validate:"omitempty,email,max=128"`
	Avatar   string `json:"avatar" validate:"omitempty,max=500"`
	GroupID  uint64 `json:"group_id" validate:"required,min=1"`
	Status   *uint8 `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateAdminRequest updates an admin.
type UpdateAdminRequest struct {
	Username string  `json:"username" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email" validate:"omitempty,max=128"`
	Avatar   string  `json:"avatar" validate:"omitempty,max=500"`
	GroupID  *uint64 `json:"group_id" validate:"omitempty,min=1"`
	Status   *uint8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// ResetAdminPasswordRequest resets admin password.
type ResetAdminPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// ===== User group CRUD (A4) =====

// UserGroupListRequest filters user group list.
type UserGroupListRequest struct {
	dto.PaginationRequest
	Keyword string `form:"keyword"`
}

// UserGroupItem is the user group list item.
type UserGroupItem struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Permissions *string `json:"permissions"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

// CreateUserGroupRequest creates a user group.
type CreateUserGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=64"`
	Permissions string `json:"permissions" validate:"omitempty,max=2000"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

// UpdateUserGroupRequest updates a user group.
type UpdateUserGroupRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=64"`
	Permissions string `json:"permissions" validate:"omitempty,max=2000"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

// ===== Regular user CRUD (A5) =====

// UserListRequest filters regular user list.
type UserListRequest struct {
	dto.PaginationRequest
	Keyword string `form:"keyword"`
	Status  *uint8 `form:"status"`
}

// UserItem is the user list item.
type UserItem struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
	Status      uint8  `json:"status"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

// CreateUserRequest creates a new regular user.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Email    string `json:"email" validate:"omitempty,email,max=128"`
	Avatar   string `json:"avatar" validate:"omitempty,max=500"`
	Status   *uint8 `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateUserRequest updates a regular user.
type UpdateUserRequest struct {
	Username string  `json:"username" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email" validate:"omitempty,max=128"`
	Avatar   string  `json:"avatar" validate:"omitempty,max=500"`
	Status   *uint8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// ResetUserPasswordRequest resets user password.
type ResetUserPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// ===== Banner CRUD (C1) =====

// BannerItem is the banner list item.
type BannerItem struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	Cover   string `json:"cover"`
	Link    string `json:"link"`
	VideoID uint64 `json:"video_id"`
	Sort    uint32 `json:"sort"`
	Status  uint8  `json:"status"`
}

// CreateBannerRequest creates a banner.
type CreateBannerRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=128"`
	Cover   string `json:"cover" validate:"required,max=500"`
	Link    string `json:"link" validate:"omitempty,max=500"`
	VideoID uint64 `json:"video_id" validate:"omitempty,min=1"`
	Sort    uint32 `json:"sort" validate:"omitempty,min=0"`
	Status  *uint8 `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateBannerRequest updates a banner.
type UpdateBannerRequest struct {
	Title   string  `json:"title" validate:"omitempty,min=1,max=128"`
	Cover   string  `json:"cover" validate:"omitempty,max=500"`
	Link    string  `json:"link" validate:"omitempty,max=500"`
	VideoID *uint64 `json:"video_id" validate:"omitempty,min=1"`
	Sort    *uint32 `json:"sort" validate:"omitempty,min=0"`
	Status  *uint8  `json:"status" validate:"omitempty,oneof=0 1"`
}

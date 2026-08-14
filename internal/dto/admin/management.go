package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== Dashboard / Stats (A1) =====

// DashboardResponse is the admin dashboard overview.
type DashboardResponse struct {
	// 视频总数
	TotalVideos int64 `json:"total_videos"`
	// 今日新增视频数
	TodayVideos int64 `json:"today_videos"`
	// 已上架视频数
	OnlineVideos int64 `json:"online_videos"`
	// 未上架视频数
	OfflineVideos int64 `json:"offline_videos"`
	// 分类总数
	TotalCategories int64 `json:"total_categories"`
	// 管理员总数
	TotalAdmins int64 `json:"total_admins"`
	// 用户总数
	TotalUsers int64 `json:"total_users"`
	// 当前在线用户数
	OnlineCount int64 `json:"online_count"`
	// 今日浏览量（PV）
	TodayPV int64 `json:"today_pv"`
	// 今日访客数（UV）
	TodayUV int64 `json:"today_uv"`
	// 正在执行的采集任务数
	CollectRunning int64 `json:"collect_running"`
}

// ===== Batch video ops (A2) =====

// BatchVideoRequest is the batch video operation payload.
type BatchVideoRequest struct {
	// 视频ID列表（1-500 个，必填）
	IDs []uint32 `json:"ids" binding:"required,min=1,max=500"`
	// 操作后的状态（0=下架，1=上架）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// BatchVideoResponse is the batch operation result.
type BatchVideoResponse struct {
	// 受影响的行数
	Affected int `json:"affected"`
}

// ===== Admin CRUD (A3) =====

// AdminListRequest filters admin list.
type AdminListRequest struct {
	dto.PaginationRequest
	// 关键词搜索（用户名/昵称）
	Keyword string `form:"keyword"`
	// 状态筛选（0=禁用，1=启用）
	Status *uint8 `form:"status"`
	// 用户组ID筛选
	GroupID *uint32 `form:"group_id"`
}

// AdminItem is the admin list item.
type AdminItem struct {
	// 管理员ID
	ID uint32 `json:"id"`
	// 用户名
	Username string `json:"username"`
	// 邮箱
	Email string `json:"email"`
	// 头像地址
	Avatar string `json:"avatar"`
	// 用户组ID
	GroupID uint32 `json:"group_id"`
	// 用户组名称
	GroupName string `json:"group_name"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
	// 最近登录时间
	LastLoginAt string `json:"last_login_at"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

// CreateAdminRequest creates a new admin.
type CreateAdminRequest struct {
	// 用户名（必填，3-50 位）
	Username string `json:"username" binding:"required,min=3,max=50"`
	// 密码（必填，6-72 位）
	Password string `json:"password" binding:"required,min=6,max=72"`
	// 邮箱（可选）
	Email string `json:"email" binding:"omitempty,email,max=128"`
	// 头像地址（可选）
	Avatar string `json:"avatar" binding:"omitempty,max=500"`
	// 用户组ID（必填）
	GroupID uint32 `json:"group_id" binding:"required,min=1"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateAdminRequest updates an admin.
type UpdateAdminRequest struct {
	// 用户名（3-50 位）
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	// 邮箱
	Email *string `json:"email" binding:"omitempty,max=128"`
	// 头像地址
	Avatar string `json:"avatar" binding:"omitempty,max=500"`
	// 用户组ID
	GroupID *uint32 `json:"group_id" binding:"omitempty,min=1"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// ResetAdminPasswordRequest resets admin password.
type ResetAdminPasswordRequest struct {
	// 新密码（必填，6-72 位）
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// ===== User group CRUD (A4) =====

// UserGroupListRequest filters user group list.
type UserGroupListRequest struct {
	dto.PaginationRequest
	// 关键词搜索（名称）
	Keyword string `form:"keyword"`
}

// UserGroupItem is the user group list item.
type UserGroupItem struct {
	// 用户组ID
	ID uint32 `json:"id"`
	// 用户组名称
	Name string `json:"name"`
	// 权限配置（JSON 字符串）
	Permissions *string `json:"permissions"`
	// 描述
	Description string `json:"description"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

// CreateUserGroupRequest creates a user group.
type CreateUserGroupRequest struct {
	// 用户组名称（必填，1-64 字）
	Name string `json:"name" binding:"required,min=1,max=64"`
	// 权限配置（JSON 字符串）
	Permissions string `json:"permissions" binding:"omitempty,max=2000"`
	// 描述
	Description string `json:"description" binding:"omitempty,max=255"`
}

// UpdateUserGroupRequest updates a user group.
type UpdateUserGroupRequest struct {
	// 用户组名称
	Name string `json:"name" binding:"omitempty,min=1,max=64"`
	// 权限配置（JSON 字符串）
	Permissions string `json:"permissions" binding:"omitempty,max=2000"`
	// 描述
	Description string `json:"description" binding:"omitempty,max=255"`
}

// ===== Regular user CRUD (A5) =====

// UserListRequest filters regular user list.
type UserListRequest struct {
	dto.PaginationRequest
	// 关键词搜索（用户名/昵称）
	Keyword string `form:"keyword"`
	// 状态筛选（0=禁用，1=启用）
	Status *uint8 `form:"status"`
}

// UserItem is the user list item.
type UserItem struct {
	// 用户ID
	ID uint32 `json:"id"`
	// 字符串形式用户ID
	StrID string `json:"str_id"`
	// 用户名
	Username string `json:"username"`
	// 昵称
	Nickname string `json:"nickname"`
	// 邮箱
	Email string `json:"email"`
	// 头像地址
	Avatar string `json:"avatar"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
	// 最近登录时间
	LastLoginAt string `json:"last_login_at"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}

// CreateUserRequest creates a new regular user.
type CreateUserRequest struct {
	// 用户名（必填，3-50 位）
	Username string `json:"username" binding:"required,min=3,max=50"`
	// 密码（必填，6-72 位）
	Password string `json:"password" binding:"required,min=6,max=72"`
	// 昵称（可选）
	Nickname string `json:"nickname" binding:"omitempty,min=3,max=15"`
	// 邮箱（可选）
	Email string `json:"email" binding:"omitempty,email,max=128"`
	// 头像地址（可选）
	Avatar string `json:"avatar" binding:"omitempty,max=500"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateUserRequest updates a regular user.
type UpdateUserRequest struct {
	// 用户名（3-50 位）
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	// 昵称（3-15 位）
	Nickname string `json:"nickname" binding:"omitempty,min=3,max=15"`
	// 邮箱
	Email *string `json:"email" binding:"omitempty,max=128"`
	// 头像地址
	Avatar string `json:"avatar" binding:"omitempty,max=500"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// ResetUserPasswordRequest resets user password.
type ResetUserPasswordRequest struct {
	// 新密码（必填，6-72 位）
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// ===== Banner CRUD (C1) =====

// BannerItem is the banner list item.
type BannerItem struct {
	// 横幅ID
	ID uint32 `json:"id"`
	// 横幅标题
	Title string `json:"title"`
	// 封面地址
	Cover string `json:"cover"`
	// 跳转链接
	Link string `json:"link"`
	// 关联视频ID
	VideoID uint32 `json:"video_id"`
	// 排序权重，值越小越靠前
	Sort uint32 `json:"sort"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
}

// CreateBannerRequest creates a banner.
type CreateBannerRequest struct {
	// 横幅标题（必填）
	Title string `json:"title" binding:"required,min=1,max=128"`
	// 封面地址（必填）
	Cover string `json:"cover" binding:"required,max=500"`
	// 跳转链接
	Link string `json:"link" binding:"omitempty,max=500"`
	// 关联视频ID
	VideoID uint32 `json:"video_id" binding:"omitempty,min=1"`
	// 排序权重，值越小越靠前
	Sort uint32 `json:"sort" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateBannerRequest updates a banner.
type UpdateBannerRequest struct {
	// 横幅标题
	Title string `json:"title" binding:"omitempty,min=1,max=128"`
	// 封面地址
	Cover string `json:"cover" binding:"omitempty,max=500"`
	// 跳转链接
	Link string `json:"link" binding:"omitempty,max=500"`
	// 关联视频ID
	VideoID *uint32 `json:"video_id" binding:"omitempty,min=1"`
	// 排序权重，值越小越靠前
	Sort *uint32 `json:"sort" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

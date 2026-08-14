package admin

// LoginRequest is the admin login payload.
type LoginRequest struct {
	// 管理员用户名（必填，3-50 位）
	Username string `json:"username" binding:"required,min=3,max=50"`
	// 密码（必填，6-72 位）
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginResponse is returned after successful admin login.
type LoginResponse struct {
	// 访问令牌
	AccessToken string `json:"access_token"`
	// 令牌类型（如 Bearer）
	TokenType string `json:"token_type"`
	// 令牌有效期（秒）
	ExpiresIn int `json:"expires_in"`
	// 管理员公开资料
	Admin *Profile `json:"admin"`
}

// Profile is the authenticated admin public profile.
type Profile struct {
	// 管理员ID
	ID uint32 `json:"id"`
	// 用户名
	Username string `json:"username"`
	// 昵称
	Nickname string `json:"nickname"`
	// 邮箱
	Email string `json:"email"`
	// 头像地址
	Avatar string `json:"avatar"`
	// 角色
	Role string `json:"role"`
	// 账号状态（0=禁用，1=正常）
	Status uint8 `json:"status"`
}

// UpdateProfileRequest is the payload for self-service profile update.
type UpdateProfileRequest struct {
	// 昵称（可选）
	Nickname string `json:"nickname" binding:"max=50"`
	// 邮箱（可选）
	Email string `json:"email" binding:"omitempty,max=100,email"`
	// 头像地址（可选）
	Avatar string `json:"avatar" binding:"max=500"`
}

// ChangePasswordRequest is the payload for self-service password change.
type ChangePasswordRequest struct {
	// 旧密码（必填）
	OldPassword string `json:"old_password" binding:"required,min=6,max=72"`
	// 新密码（必填，6-72 位）
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

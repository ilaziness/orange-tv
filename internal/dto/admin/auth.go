package admin

// LoginRequest is the admin login payload.
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginResponse is returned after successful admin login.
type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	Admin       *Profile `json:"admin"`
}

// Profile is the authenticated admin public profile.
type Profile struct {
	ID       uint32 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Status   uint8  `json:"status"`
}

// UpdateProfileRequest is the payload for self-service profile update.
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Email    string `json:"email" binding:"omitempty,max=100,email"`
	Avatar   string `json:"avatar" binding:"max=500"`
}

// ChangePasswordRequest is the payload for self-service password change.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=72"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

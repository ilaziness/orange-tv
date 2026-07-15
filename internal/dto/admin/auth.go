package admin

// LoginRequest is the admin login payload.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=72"`
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
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Status   int8   `json:"status"`
}

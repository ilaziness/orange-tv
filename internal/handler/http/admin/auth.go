package admin

import (
	"github.com/gin-gonic/gin"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// AuthHandler handles admin authentication endpoints.
type AuthHandler struct {
	svc adminsvc.AuthService
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(svc adminsvc.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login handles POST /api/admin/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req admindto.LoginRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	meta := &adminsvc.LoginMeta{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	resp, err := h.svc.Login(c.Request.Context(), &req, meta)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// Logout handles POST /api/admin/v1/auth/logout (stateless).
func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, nil)
}

// Profile handles GET /api/admin/v1/auth/profile.
func (h *AuthHandler) Profile(c *gin.Context) {
	claims := httpmiddleware.GetClaims(c)
	if claims == nil {
		response.Error(c, errcode.AuthFailed)
		return
	}
	profile, err := h.svc.Profile(c.Request.Context(), claims.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, profile)
}

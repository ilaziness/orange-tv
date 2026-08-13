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
// @Summary 管理员登录
// @Description 管理员账号密码登录，返回 JWT
// @Tags 管理端｜管理员认证
// @Accept json
// @Produce json
// @Param body body admindto.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=admindto.LoginResponse}
// @Router /api/admin/v1/auth/login [post]
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
// @Summary 管理员登出
// @Description 无状态登出，客户端清除 token 即可
// @Tags 管理端｜管理员认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/admin/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, nil)
}

// Profile handles GET /api/admin/v1/auth/profile.
// @Summary 获取管理员个人资料
// @Description 获取当前登录管理员的个人资料
// @Tags 管理端｜管理员认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=admindto.Profile}
// @Router /api/admin/v1/auth/profile [get]
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

// UpdateProfile handles PUT /api/admin/v1/auth/profile.
//
// @Summary 更新管理员个人资料
// @Description 当前登录管理员更新自己的昵称、邮箱、头像
// @Tags 管理端｜管理员认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.UpdateProfileRequest true "更新资料请求"
// @Success 200 {object} response.Response{data=admindto.Profile}
// @Router /api/admin/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	claims := httpmiddleware.GetClaims(c)
	if claims == nil {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req admindto.UpdateProfileRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	profile, err := h.svc.UpdateProfile(c.Request.Context(), claims.UserID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, profile)
}

// ChangePassword handles PUT /api/admin/v1/auth/profile/password.
//
// @Summary 修改管理员密码
// @Description 当前登录管理员修改自己的密码
// @Tags 管理端｜管理员认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/auth/profile/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims := httpmiddleware.GetClaims(c)
	if claims == nil {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req admindto.ChangePasswordRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), claims.UserID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

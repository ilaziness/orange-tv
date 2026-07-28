package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/audit"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// SettingsHandler handles admin system settings.
type SettingsHandler struct {
	svc   adminsvc.SettingsService
	audit *audit.Recorder
}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler(svc adminsvc.SettingsService, recorder *audit.Recorder) *SettingsHandler {
	return &SettingsHandler{svc: svc, audit: recorder}
}

// GetSettings godoc
// @Summary 获取系统设置
// @Description 按分组获取系统设置（site/api/ad）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Param group query string true "设置分组 (site/api/ad)"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/settings [get]
func (h *SettingsHandler) Get(c *gin.Context) {
	group := c.Query("group")
	if group == "" {
		response.Error(c, errcode.WithMessage(errcode.ParamError, "group 参数不能为空"))
		return
	}
	resp, err := h.svc.Get(c.Request.Context(), group)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// UpdateSettings godoc
// @Summary 更新系统设置
// @Description 按分组更新系统设置，data 为对应分组的 key-value JSON
// @Tags 系统设置
// @Accept json
// @Produce json
// @Param body body admindto.UpdateSettingsRequest true "更新设置请求"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/settings [put]
func (h *SettingsHandler) Update(c *gin.Context) {
	var req admindto.UpdateSettingsRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.Update(c.Request.Context(), req.Group, req.Data)
	if err != nil {
		response.Error(c, err)
		return
	}
	if h.audit != nil {
		adminID := int64(0)
		if claims := httpmiddleware.GetClaims(c); claims != nil {
			adminID = claims.UserID
		}
		h.audit.AdminAction(c.Request.Context(), adminID, "settings", "update", "更新系统设置", c.ClientIP())
	}
	response.Success(c, resp)
}

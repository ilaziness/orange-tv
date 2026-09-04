package admin

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/audit"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
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

// GetSettings
// @Summary 获取系统设置
// @Description 按分组获取系统设置。data 结构随 group 变化：site=SiteSettings，api=APISettings，feature=FeatureSettings，seo=SEOSettings
// @Tags 管理端｜系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.GetSettingsQuery true "查询参数"
// @Success 200 {object} response.Response{data=map[string]any}
// @Router /api/admin/v1/settings [get]
func (h *SettingsHandler) Get(c *gin.Context) {
	var q admindto.GetSettingsQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}
	resp, err := h.svc.Get(c.Request.Context(), q.Group)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// UpdateSettings
// @Summary 更新系统设置
// @Description 按分组更新系统设置，data 为对应分组的 key-value JSON（site/api/feature/seo），返回更新后的设置结构
// @Tags 管理端｜系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.UpdateSettingsRequest true "更新设置请求"
// @Success 200 {object} response.Response{data=map[string]any}
// @Router /api/admin/v1/settings [put]
func (h *SettingsHandler) Update(c *gin.Context) {
	var req admindto.UpdateSettingsRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	data, err := json.Marshal(req.Data)
	if err != nil {
		response.Error(c, err)
		return
	}
	resp, err := h.svc.Update(c.Request.Context(), req.Group, json.RawMessage(data))
	if err != nil {
		response.Error(c, err)
		return
	}
	if h.audit != nil {
		adminID := uint32(0)
		if claims := httpmiddleware.GetClaims(c); claims != nil {
			adminID = claims.UserID
		}
		h.audit.AdminAction(c.Request.Context(), adminID, "settings", "update", "更新系统设置", c.ClientIP())
	}
	response.Success(c, resp)
}

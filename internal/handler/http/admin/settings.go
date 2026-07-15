package admin

import (
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

func (h *SettingsHandler) Get(c *gin.Context) {
	resp, err := h.svc.Get(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var req admindto.UpdateSettingsRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.Update(c.Request.Context(), &req)
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

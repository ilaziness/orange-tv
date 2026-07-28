package client

import (
	"github.com/gin-gonic/gin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// SiteHandler exposes public site info and client settings.
type SiteHandler struct {
	settings clientsvc.ClientSettingsService
}

// NewSiteHandler creates a SiteHandler.
func NewSiteHandler(settings clientsvc.ClientSettingsService) *SiteHandler {
	return &SiteHandler{settings: settings}
}

// GetSettings handles GET /api/client/v1/settings?group=xxx
// Returns settings for the specified group (whitelist filtered).
func (h *SiteHandler) GetSettings(c *gin.Context) {
	group := c.Query("group")
	if group == "" {
		response.Error(c, errcode.WithMessage(errcode.ParamError, "group 参数不能为空"))
		return
	}
	resp, err := h.settings.GetByGroup(c.Request.Context(), group)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

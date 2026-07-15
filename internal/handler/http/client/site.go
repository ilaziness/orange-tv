package client

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// SiteHandler exposes public site info.
type SiteHandler struct {
	settings adminsvc.SettingsService
}

// NewSiteHandler creates a SiteHandler.
func NewSiteHandler(settings adminsvc.SettingsService) *SiteHandler {
	return &SiteHandler{settings: settings}
}

func (h *SiteHandler) Public(c *gin.Context) {
	resp, err := h.settings.GetPublic(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

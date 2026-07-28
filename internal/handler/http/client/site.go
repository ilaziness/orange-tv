package client

import (
	"github.com/gin-gonic/gin"
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

// Public handles GET /api/client/v1/site
func (h *SiteHandler) Public(c *gin.Context) {
	resp, err := h.settings.GetPublic(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// GetSettings handles GET /api/client/v1/settings
// If group query param is provided, returns settings for that group (whitelist filtered).
// If no group param, returns all whitelisted groups.
func (h *SiteHandler) GetSettings(c *gin.Context) {
	group := c.Query("group")
	if group == "" {
		resp, err := h.settings.GetAll(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}
		response.Success(c, resp)
		return
	}
	resp, err := h.settings.GetByGroup(c.Request.Context(), group)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

package client

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// ThemeHandler handles public theme endpoints.
type ThemeHandler struct {
	svc clientsvc.ThemeService
}

// NewThemeHandler creates a ThemeHandler.
func NewThemeHandler(svc clientsvc.ThemeService) *ThemeHandler {
	return &ThemeHandler{svc: svc}
}

func (h *ThemeHandler) Current(c *gin.Context) {
	item, err := h.svc.Current(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

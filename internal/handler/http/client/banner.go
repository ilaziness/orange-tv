package client

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// BannerHandler exposes public banner list for client.
type BannerHandler struct {
	svc clientsvc.BannerService
}

// NewBannerHandler creates a BannerHandler.
func NewBannerHandler(svc clientsvc.BannerService) *BannerHandler {
	return &BannerHandler{svc: svc}
}

func (h *BannerHandler) List(c *gin.Context) {
	list, err := h.svc.ListBanners(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

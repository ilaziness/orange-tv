package client

import (
	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
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

// List
// @Summary Banner列表
// @Description 获取公开启用的Banner列表
// @Tags 用户端｜Banner
// @Produce json
// @Success 200 {object} response.Response{data=[]clientdto.BannerItem}
// @Router /api/client/v1/banners [get]
func (h *BannerHandler) List(c *gin.Context) {
	var items []clientdto.BannerItem
	var err error
	items, err = h.svc.ListBanners(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

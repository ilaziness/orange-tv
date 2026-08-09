package client

import (
	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// AdHandler exposes public advertisement list for client.
type AdHandler struct {
	svc clientsvc.AdService
}

// NewAdHandler creates a client AdHandler.
func NewAdHandler(svc clientsvc.AdService) *AdHandler {
	return &AdHandler{svc: svc}
}

// List godoc
// @Summary 广告列表
// @Description 按场景获取启用的广告列表
// @Tags 用户端｜广告
// @Produce json
// @Param scene query string true "广告场景：video_loading 播放前广告 / general 一般广告" Enums(video_loading, general)
// @Success 200 {object} response.Response
// @Router /api/client/v1/promotions [get]
func (h *AdHandler) List(c *gin.Context) {
	var q clientdto.ListAdQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}
	list, err := h.svc.ListByScene(c.Request.Context(), q.Scene)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

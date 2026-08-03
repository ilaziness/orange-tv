package client

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// LiveHandler handles public live channel endpoints.
type LiveHandler struct {
	svc      clientsvc.LiveService
	proxySvc clientsvc.LiveProxyService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc clientsvc.LiveService, proxySvc clientsvc.LiveProxyService) *LiveHandler {
	return &LiveHandler{svc: svc, proxySvc: proxySvc}
}

func (h *LiveHandler) List(c *gin.Context) {
	var req clientdto.LiveListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// Play proxies a live stream by delegating to LiveProxyService.
func (h *LiveHandler) Play(c *gin.Context) {
	id := c.Param("id")
	channelID, err := parseID(id)
	if err != nil {
		response.Error(c, errcode.WithMessage(errcode.ParamError, "无效的频道 id"))
		return
	}

	segURL := strings.TrimSpace(c.Query("u"))
	if err := h.proxySvc.Proxy(c, channelID, segURL); err != nil {
		response.Error(c, err)
	}
}

// parseID 解析频道 id。
func parseID(s string) (uint32, error) {
	var id uint32
	if _, err := fmt.Sscanf(s, "%d", &id); err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("id must be positive")
	}
	return id, nil
}

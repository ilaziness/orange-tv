package client

import (
	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// LiveHandler handles public live channel endpoints.
type LiveHandler struct {
	svc clientsvc.LiveService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc clientsvc.LiveService) *LiveHandler {
	return &LiveHandler{svc: svc}
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

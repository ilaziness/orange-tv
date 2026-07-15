package client

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// VideoHandler handles client video endpoints.
type VideoHandler struct {
	svc clientsvc.VideoService
}

// NewVideoHandler creates a VideoHandler.
func NewVideoHandler(svc clientsvc.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

func (h *VideoHandler) List(c *gin.Context) {
	var req clientdto.VideoListRequest
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

func (h *VideoHandler) Search(c *gin.Context) {
	var req clientdto.SearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.Search(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

func (h *VideoHandler) Get(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	item, err := h.svc.Get(c.Request.Context(), uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

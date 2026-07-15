package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// VideoHandler handles admin video endpoints.
type VideoHandler struct {
	svc adminsvc.VideoService
}

// NewVideoHandler creates a VideoHandler.
func NewVideoHandler(svc adminsvc.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

func (h *VideoHandler) List(c *gin.Context) {
	var req admindto.VideoListRequest
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

func (h *VideoHandler) Create(c *gin.Context) {
	var req admindto.CreateVideoRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *VideoHandler) Update(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateVideoRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.Update(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *VideoHandler) Delete(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

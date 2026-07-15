package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// CollectHandler handles admin collect endpoints.
type CollectHandler struct {
	svc adminsvc.CollectService
}

// NewCollectHandler creates a CollectHandler.
func NewCollectHandler(svc adminsvc.CollectService) *CollectHandler {
	return &CollectHandler{svc: svc}
}

func (h *CollectHandler) ListSources(c *gin.Context) {
	var req admindto.CollectSourceListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListSources(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

func (h *CollectHandler) CreateSource(c *gin.Context) {
	var req admindto.CreateCollectSourceRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateSource(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *CollectHandler) UpdateSource(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateCollectSourceRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateSource(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *CollectHandler) DeleteSource(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteSource(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *CollectHandler) ListCategories(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	items, err := h.svc.ListCategories(c.Request.Context(), uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

func (h *CollectHandler) SetCategories(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.SetCollectCategoriesRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	items, err := h.svc.SetCategories(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

func (h *CollectHandler) Start(c *gin.Context) {
	var uri admindto.SourceIDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.Start(c.Request.Context(), uri.SourceID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"started": true})
}

func (h *CollectHandler) Stop(c *gin.Context) {
	var uri admindto.SourceIDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.Stop(c.Request.Context(), uri.SourceID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"stopped": true})
}

func (h *CollectHandler) ListLogs(c *gin.Context) {
	var req admindto.CollectLogListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// MetadataHandler handles directors/actors/tags admin endpoints.
type MetadataHandler struct {
	svc adminsvc.MetadataService
}

// NewMetadataHandler creates a MetadataHandler.
func NewMetadataHandler(svc adminsvc.MetadataService) *MetadataHandler {
	return &MetadataHandler{svc: svc}
}

func (h *MetadataHandler) ListDirectors(c *gin.Context) {
	var req admindto.NameSearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListDirectors(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

func (h *MetadataHandler) CreateDirector(c *gin.Context) {
	var req admindto.CreateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateDirector(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *MetadataHandler) UpdateDirector(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateDirector(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *MetadataHandler) DeleteDirector(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteDirector(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *MetadataHandler) ListActors(c *gin.Context) {
	var req admindto.NameSearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListActors(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

func (h *MetadataHandler) CreateActor(c *gin.Context) {
	var req admindto.CreateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateActor(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *MetadataHandler) UpdateActor(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateActor(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *MetadataHandler) DeleteActor(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteActor(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *MetadataHandler) ListTags(c *gin.Context) {
	var req admindto.NameSearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListTags(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

func (h *MetadataHandler) CreateTag(c *gin.Context) {
	var req admindto.CreateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateTag(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *MetadataHandler) UpdateTag(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateTag(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *MetadataHandler) DeleteTag(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteTag(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

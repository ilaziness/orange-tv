package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// CategoryHandler handles admin category endpoints.
type CategoryHandler struct {
	svc adminsvc.CategoryService
}

// NewCategoryHandler creates a CategoryHandler.
func NewCategoryHandler(svc adminsvc.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c *gin.Context) {
	items, err := h.svc.ListTree(c.Request.Context(), false)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req admindto.CreateCategoryRequest
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

func (h *CategoryHandler) Update(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateCategoryRequest
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

func (h *CategoryHandler) Delete(c *gin.Context) {
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

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

// List
// @Summary 分类树
// @Description 获取分类树形列表
// @Tags 管理端｜分类管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]admindto.CategoryResponse}
// @Router /api/admin/v1/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	items, err := h.svc.ListTree(c.Request.Context(), false)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

// Create
// @Summary 新建分类
// @Description 创建一个新的分类
// @Tags 管理端｜分类管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateCategoryRequest true "分类参数"
// @Success 200 {object} response.Response{data=admindto.CategoryResponse}
// @Router /api/admin/v1/categories [post]
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

// Update
// @Summary 更新分类
// @Description 更新指定分类信息
// @Tags 管理端｜分类管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Param body body admindto.UpdateCategoryRequest true "分类参数"
// @Success 200 {object} response.Response{data=admindto.CategoryResponse}
// @Router /api/admin/v1/categories/{id} [put]
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

// Delete
// @Summary 删除分类
// @Description 删除指定分类
// @Tags 管理端｜分类管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/categories/{id} [delete]
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

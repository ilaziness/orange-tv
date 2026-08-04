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

// List godoc
// @Summary 影视列表
// @Description 分页获取影视列表
// @Tags 管理端｜视频管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/videos [get]
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

// Get godoc
// @Summary 影视详情
// @Description 获取指定影视详情
// @Tags 管理端｜视频管理
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/videos/{id} [get]
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

// Create godoc
// @Summary 新建影视
// @Description 创建一个新的影视
// @Tags 管理端｜视频管理
// @Accept json
// @Produce json
// @Param body body admindto.CreateVideoRequest true "影视参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/videos [post]
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

// Update godoc
// @Summary 更新影视
// @Description 更新指定影视信息
// @Tags 管理端｜视频管理
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Param body body admindto.UpdateVideoRequest true "影视参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/videos/{id} [put]
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

// Delete godoc
// @Summary 删除影视
// @Description 删除指定影视
// @Tags 管理端｜视频管理
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/videos/{id} [delete]
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

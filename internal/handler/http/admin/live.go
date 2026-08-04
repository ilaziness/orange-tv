package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// LiveHandler handles admin live channel endpoints.
type LiveHandler struct {
	svc adminsvc.LiveService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc adminsvc.LiveService) *LiveHandler {
	return &LiveHandler{svc: svc}
}

// List godoc
// @Summary 直播频道列表
// @Description 分页获取直播频道列表
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live [get]
func (h *LiveHandler) List(c *gin.Context) {
	var req admindto.LiveListRequest
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

// Create godoc
// @Summary 新建直播频道
// @Description 创建一个新的直播频道
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Param body body admindto.CreateLiveRequest true "频道参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live [post]
func (h *LiveHandler) Create(c *gin.Context) {
	var req admindto.CreateLiveRequest
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
// @Summary 更新直播频道
// @Description 更新指定直播频道信息
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Param id path int true "频道ID"
// @Param body body admindto.UpdateLiveRequest true "频道参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/{id} [put]
func (h *LiveHandler) Update(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateLiveRequest
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
// @Summary 删除直播频道
// @Description 删除指定直播频道
// @Tags 管理端｜直播管理
// @Produce json
// @Param id path int true "频道ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/{id} [delete]
func (h *LiveHandler) Delete(c *gin.Context) {
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

// Sync godoc
// @Summary 同步直播源
// @Description 从外部直播源同步频道列表
// @Tags 管理端｜直播管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/sync [post]
func (h *LiveHandler) Sync(c *gin.Context) {
	result, err := h.svc.SyncFromSource(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

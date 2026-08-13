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

// ListSources
// @Summary 采集源列表
// @Description 获取采集源分页列表
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.CollectSourceListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.CollectSourceItem}}
// @Router /api/admin/v1/collect-sources [get]
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

// CreateSource
// @Summary 新建采集源
// @Description 创建一个新的采集源
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateCollectSourceRequest true "采集源参数"
// @Success 200 {object} response.Response{data=admindto.CollectSourceItem}
// @Router /api/admin/v1/collect-sources [post]
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

// UpdateSource
// @Summary 更新采集源
// @Description 更新指定采集源信息
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Param body body admindto.UpdateCollectSourceRequest true "采集源参数"
// @Success 200 {object} response.Response{data=admindto.CollectSourceItem}
// @Router /api/admin/v1/collect-sources/{id} [put]
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

// DeleteSource
// @Summary 删除采集源
// @Description 删除指定采集源
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/collect-sources/{id} [delete]
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

// ListCategories
// @Summary 采集源分类映射列表
// @Description 获取指定采集源的分类映射列表
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Success 200 {object} response.Response{data=[]admindto.CollectCategoryMapItem}
// @Router /api/admin/v1/collect-sources/{id}/categories [get]
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

// SetCategories
// @Summary 设置采集源分类映射
// @Description 替换指定采集源的分类映射
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Param body body admindto.SetCollectCategoriesRequest true "分类映射参数"
// @Success 200 {object} response.Response{data=[]admindto.CollectCategoryMapItem}
// @Router /api/admin/v1/collect-sources/{id}/categories [post]
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

// ListLogs
// @Summary 采集日志列表
// @Description 获取采集日志分页列表
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.CollectLogListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.CollectLogItem}}
// @Router /api/admin/v1/collect/logs [get]
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

// FetchRemoteCategories
// @Summary 获取远程分类
// @Description 从远程采集源拉取分类列表（仅苹果CMS）
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Success 200 {object} response.Response{data=admindto.RemoteCategoryResponse}
// @Router /api/admin/v1/collect-sources/{id}/remote-categories [get]
func (h *CollectHandler) FetchRemoteCategories(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	resp, err := h.svc.FetchRemoteCategories(c.Request.Context(), uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// EnableSchedule
// @Summary 启用定时采集
// @Description 启用指定采集源的定时采集（需已绑定分类和设置cron表达式），返回 data.enabled=true
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Success 200 {object} response.Response{data=map[string]bool}
// @Router /api/admin/v1/collect-sources/{id}/schedule/enable [post]
func (h *CollectHandler) EnableSchedule(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.EnableSchedule(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"enabled": true})
}

// DisableSchedule
// @Summary 禁用定时采集
// @Description 禁用指定采集源的定时采集，返回 data.disabled=true
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Success 200 {object} response.Response{data=map[string]bool}
// @Router /api/admin/v1/collect-sources/{id}/schedule/disable [post]
func (h *CollectHandler) DisableSchedule(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DisableSchedule(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"disabled": true})
}

// CollectNow
// @Summary 立即采集
// @Description 立即执行一次采集任务，返回 data.started=true
// @Tags 管理端｜采集管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "采集源ID"
// @Param body body admindto.CollectNowRequest true "立即采集参数"
// @Success 200 {object} response.Response{data=map[string]bool}
// @Router /api/admin/v1/collect-sources/{id}/collect [post]
func (h *CollectHandler) CollectNow(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.CollectNowRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.CollectNow(c.Request.Context(), uri.ID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"started": true})
}

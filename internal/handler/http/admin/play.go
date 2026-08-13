package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// PlayHandler handles admin play source and episode endpoints.
type PlayHandler struct {
	svc adminsvc.PlayService
}

// NewPlayHandler creates a PlayHandler.
func NewPlayHandler(svc adminsvc.PlayService) *PlayHandler {
	return &PlayHandler{svc: svc}
}

// ListSources
// @Summary 播放源列表
// @Description 获取全部播放源列表
// @Tags 管理端｜播放源管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.PlaySourceResponse}}
// @Router /api/admin/v1/play-sources [get]
func (h *PlayHandler) ListSources(c *gin.Context) {
	items, err := h.svc.ListSources(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	pageSize := len(items)
	if pageSize == 0 {
		pageSize = 20
	}
	response.SuccessPage(c, items, int64(len(items)), 1, pageSize, 1)
}

// CreateSource
// @Summary 新建播放源
// @Description 创建一个新的播放源
// @Tags 管理端｜播放源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreatePlaySourceRequest true "播放源参数"
// @Success 200 {object} response.Response{data=admindto.PlaySourceResponse}
// @Router /api/admin/v1/play-sources [post]
func (h *PlayHandler) CreateSource(c *gin.Context) {
	var req admindto.CreatePlaySourceRequest
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
// @Summary 更新播放源
// @Description 更新指定播放源信息
// @Tags 管理端｜播放源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "播放源ID"
// @Param body body admindto.UpdatePlaySourceRequest true "播放源参数"
// @Success 200 {object} response.Response{data=admindto.PlaySourceResponse}
// @Router /api/admin/v1/play-sources/{id} [put]
func (h *PlayHandler) UpdateSource(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdatePlaySourceRequest
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
// @Summary 删除播放源
// @Description 删除指定播放源
// @Tags 管理端｜播放源管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "播放源ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/play-sources/{id} [delete]
func (h *PlayHandler) DeleteSource(c *gin.Context) {
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

// ListEpisodes
// @Summary 剧集列表
// @Description 分页获取剧集列表
// @Tags 管理端｜播放源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.PlayEpisodeListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.PlayEpisodeResponse}}
// @Router /api/admin/v1/play-episodes [get]
func (h *PlayHandler) ListEpisodes(c *gin.Context) {
	var req admindto.PlayEpisodeListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListEpisodes(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateEpisode
// @Summary 新建剧集
// @Description 创建一条新的剧集
// @Tags 管理端｜播放源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreatePlayEpisodeRequest true "剧集参数"
// @Success 200 {object} response.Response{data=admindto.PlayEpisodeResponse}
// @Router /api/admin/v1/play-episodes [post]
func (h *PlayHandler) CreateEpisode(c *gin.Context) {
	var req admindto.CreatePlayEpisodeRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateEpisode(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateEpisode
// @Summary 更新剧集
// @Description 更新指定剧集信息
// @Tags 管理端｜播放源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "剧集ID"
// @Param body body admindto.UpdatePlayEpisodeRequest true "剧集参数"
// @Success 200 {object} response.Response{data=admindto.PlayEpisodeResponse}
// @Router /api/admin/v1/play-episodes/{id} [put]
func (h *PlayHandler) UpdateEpisode(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdatePlayEpisodeRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateEpisode(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteEpisode
// @Summary 删除剧集
// @Description 删除指定剧集
// @Tags 管理端｜播放源管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "剧集ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/play-episodes/{id} [delete]
func (h *PlayHandler) DeleteEpisode(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteEpisode(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// BatchUpdateEpisodeStatus 批量更新某影视下指定播放源的全部剧集上下架状态。
// @Summary 批量更新剧集上下架状态
// @Tags 管理端｜播放源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.BatchUpdateEpisodeStatusRequest true "批量更新剧集状态请求"
// @Success 200 {object} response.Response{data=admindto.BatchUpdateEpisodeStatusResponse}
// @Router /api/admin/v1/play-episodes/batch-status [post]
func (h *PlayHandler) BatchUpdateEpisodeStatus(c *gin.Context) {
	var req admindto.BatchUpdateEpisodeStatusRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.BatchUpdateEpisodeStatus(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

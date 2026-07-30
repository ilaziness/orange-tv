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
// @Tags 播放源管理
// @Accept json
// @Produce json
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

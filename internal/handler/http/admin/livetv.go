package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// LiveTVHandler handles admin livetv channel endpoints.
type LiveTVHandler struct {
	svc adminsvc.LiveTVService
}

// NewLiveTVHandler creates a LiveTVHandler.
func NewLiveTVHandler(svc adminsvc.LiveTVService) *LiveTVHandler {
	return &LiveTVHandler{svc: svc}
}

// List
// @Summary 直播频道列表
// @Description 分页获取直播频道列表
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.LiveTVListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.LiveTVChannelItem}}
// @Router /api/admin/v1/livetv [get]
func (h *LiveTVHandler) List(c *gin.Context) {
	var req admindto.LiveTVListRequest
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

// Create
// @Summary 新建直播频道
// @Description 创建一个新的直播频道
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateLiveTVRequest true "频道参数"
// @Success 200 {object} response.Response{data=admindto.LiveTVChannelItem}
// @Router /api/admin/v1/livetv [post]
func (h *LiveTVHandler) Create(c *gin.Context) {
	var req admindto.CreateLiveTVRequest
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
// @Summary 更新直播频道
// @Description 更新指定直播频道信息
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "频道ID"
// @Param body body admindto.UpdateLiveTVRequest true "频道参数"
// @Success 200 {object} response.Response{data=admindto.LiveTVChannelItem}
// @Router /api/admin/v1/livetv/{id} [put]
func (h *LiveTVHandler) Update(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateLiveTVRequest
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
// @Summary 删除直播频道
// @Description 删除指定直播频道
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "频道ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/livetv/{id} [delete]
func (h *LiveTVHandler) Delete(c *gin.Context) {
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

// GetSyncSource
// @Summary 获取上次同步的直播源地址
// @Description 返回 system_settings 中保存的最近一次直播源同步地址
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=admindto.LiveTVSyncSourceResponse}
// @Router /api/admin/v1/livetv/sync-source [get]
func (h *LiveTVHandler) GetSyncSource(c *gin.Context) {
	sourceURL, err := h.svc.GetSyncSourceURL(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, admindto.LiveTVSyncSourceResponse{SourceURL: sourceURL})
}

// Sync
// @Summary 同步直播源
// @Description 从外部直播源同步频道列表，支持 txt 和 m3u 格式
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.LiveTVSyncRequest true "直播源地址"
// @Success 200 {object} response.Response{data=admindto.LiveTVSyncResult}
// @Router /api/admin/v1/livetv/sync [post]
func (h *LiveTVHandler) Sync(c *gin.Context) {
	var req admindto.LiveTVSyncRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.SaveSyncSourceURL(c.Request.Context(), req.SourceURL); err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.SyncFromSource(c.Request.Context(), req.SourceURL)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

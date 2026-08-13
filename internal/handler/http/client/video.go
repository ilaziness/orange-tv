package client

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// VideoHandler handles client video endpoints.
type VideoHandler struct {
	svc clientsvc.VideoService
}

// NewVideoHandler creates a VideoHandler.
func NewVideoHandler(svc clientsvc.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

// List
// @Summary 影视列表
// @Description 分页获取影视列表
// @Tags 用户端｜影视浏览
// @Accept json
// @Produce json
// @Param req query clientdto.VideoListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.VideoListItem}}
// @Router /api/client/v1/videos [get]
func (h *VideoHandler) List(c *gin.Context) {
	var req clientdto.VideoListRequest
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

// Search
// @Summary 搜索影视
// @Description 根据关键词分页搜索影视
// @Tags 用户端｜影视浏览
// @Accept json
// @Produce json
// @Param req query clientdto.SearchRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.VideoListItem}}
// @Router /api/client/v1/search [get]
func (h *VideoHandler) Search(c *gin.Context) {
	var req clientdto.SearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.Search(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// Get
// @Summary 影视详情
// @Description 获取指定影视详情
// @Tags 用户端｜影视浏览
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response{data=clientdto.VideoDetailResponse}
// @Router /api/client/v1/videos/{id} [get]
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

// Related
// @Summary 相关影视
// @Description 获取指定影视的相关推荐列表
// @Tags 用户端｜影视浏览
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Param req query clientdto.RelatedRequest true "查询参数"
// @Success 200 {object} response.Response{data=[]clientdto.VideoListItem}
// @Router /api/client/v1/videos/{id}/related [get]
func (h *VideoHandler) Related(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req clientdto.RelatedRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, err := h.svc.Related(c.Request.Context(), uri.ID, req.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

// GetEpisode returns the play URL for a single episode.
// @Summary 获取单集播放地址
// @Description 根据影视ID和剧集主键ID获取播放地址
// @Tags 用户端｜影视浏览
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Param source_id path int true "播放源ID"
// @Param episode_id path int true "剧集主键ID"
// @Success 200 {object} response.Response{data=clientdto.PlayEpisodeResponse}
// @Router /api/client/v1/videos/{id}/episodes/{source_id}/{episode_id} [get]
func (h *VideoHandler) GetEpisode(c *gin.Context) {
	var uri clientdto.EpisodeURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	item, err := h.svc.GetEpisode(c.Request.Context(), uri.ID, uri.EpisodeID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

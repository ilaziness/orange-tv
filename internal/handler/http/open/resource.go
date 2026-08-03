package open

import (
	"github.com/gin-gonic/gin"
	opendto "github.com/ilaziness/orange-tv/internal/dto/open"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	opensvc "github.com/ilaziness/orange-tv/internal/service/open"
)

// ResourceHandler serves resource-station open APIs.
type ResourceHandler struct {
	svc opensvc.ResourceService
}

// NewResourceHandler creates a ResourceHandler.
func NewResourceHandler(svc opensvc.ResourceService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

// Service returns the underlying ResourceService.
func (h *ResourceHandler) Service() opensvc.ResourceService {
	return h.svc
}

// ListVideos handles GET /api/open/v1/videos
// @Summary 开放影视列表
// @Description 返回启用的影视列表，每项只包含 id、title、category_id、created_at
// @Tags 开放资源
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param limit query int false "每页数量（兼容）" default(20)
// @Param data_range query string false "数据范围（today/last1d/last3d/last1w/last1m/all）" default(all)
// @Param source query string false "播放源名称（精确匹配）"
// @Success 200 {object} response.Response{data=response.PageData}
// @Router /api/open/v1/videos [get]
func (h *ResourceHandler) ListVideos(c *gin.Context) {
	var req opendto.VideoListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListVideos(c.Request.Context(), req.GetPage(), req.GetLimit(), req.DataRange, req.Source)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetLimit(), req.GetTotalPages(total))
}

// GetVideo handles GET /api/open/v1/videos/detail
// @Summary 开放影视详情
// @Description 支持多个视频 id，返回视频详情数组
// @Tags 开放资源
// @Param id query []int true "视频 id，可重复" collectionFormat(multi)
// @Success 200 {object} response.Response{data=[]opendto.VideoDetailItem}
// @Router /api/open/v1/videos/detail [get]
func (h *ResourceHandler) GetVideo(c *gin.Context) {
	var req opendto.VideoDetailRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	data, err := h.svc.GetVideo(c.Request.Context(), req.IDs)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, data)
}

// ListCategories handles GET /api/open/v1/categories
// @Summary 开放分类列表
// @Description 返回启用中的分类扁平列表
// @Tags 开放资源
// @Success 200 {object} response.Response{data=[]opendto.CategoryItem}
// @Router /api/open/v1/categories [get]
func (h *ResourceHandler) ListCategories(c *gin.Context) {
	items, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

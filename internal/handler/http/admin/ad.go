package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// AdHandler handles advertisement CRUD endpoints.
type AdHandler struct {
	svc adminsvc.AdService
}

// NewAdHandler creates an AdHandler.
func NewAdHandler(svc adminsvc.AdService) *AdHandler {
	return &AdHandler{svc: svc}
}

// ListAds
// @Summary 广告列表
// @Description 分页获取广告列表
// @Tags 管理端｜广告管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query shareddto.PaginationRequest true "分页参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.AdItem}}
// @Router /api/admin/v1/ads [get]
func (h *AdHandler) ListAds(c *gin.Context) {
	var page shareddto.PaginationRequest
	if !httphandler.BindQuery(c, &page) {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), page.GetOffset(), page.GetPageSize())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), page.GetPage(), page.GetPageSize(), page.GetTotalPages(total))
}

// CreateAd
// @Summary 新建广告
// @Description 创建一个新的广告
// @Tags 管理端｜广告管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateAdRequest true "广告参数"
// @Success 200 {object} response.Response{data=admindto.AdItem}
// @Router /api/admin/v1/ads [post]
func (h *AdHandler) CreateAd(c *gin.Context) {
	var req admindto.CreateAdRequest
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

// UpdateAd
// @Summary 更新广告
// @Description 更新指定广告信息
// @Tags 管理端｜广告管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "广告ID"
// @Param body body admindto.UpdateAdRequest true "广告参数"
// @Success 200 {object} response.Response{data=admindto.AdItem}
// @Router /api/admin/v1/ads/{id} [put]
func (h *AdHandler) UpdateAd(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateAdRequest
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

// DeleteAd
// @Summary 删除广告
// @Description 删除指定广告
// @Tags 管理端｜广告管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "广告ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/ads/{id} [delete]
func (h *AdHandler) DeleteAd(c *gin.Context) {
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

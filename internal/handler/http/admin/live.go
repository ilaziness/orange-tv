package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// LiveHandler handles admin user-live endpoints.
type LiveHandler struct {
	svc adminsvc.LiveService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc adminsvc.LiveService) *LiveHandler {
	return &LiveHandler{svc: svc}
}

// ListCategories
// @Summary 直播分类列表
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param query query admindto.LiveCategoryListRequest true "分页筛选"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.LiveCategoryItem}}
// @Router /api/admin/v1/live/categories [get]
func (h *LiveHandler) ListCategories(c *gin.Context) {
	var req admindto.LiveCategoryListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListCategories(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateCategory
// @Summary 新建直播分类
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateLiveCategoryRequest true "分类参数"
// @Success 200 {object} response.Response{data=admindto.LiveCategoryItem}
// @Router /api/admin/v1/live/categories [post]
func (h *LiveHandler) CreateCategory(c *gin.Context) {
	var req admindto.CreateLiveCategoryRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateCategory
// @Summary 更新直播分类
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Param body body admindto.UpdateLiveCategoryRequest true "分类参数"
// @Success 200 {object} response.Response{data=admindto.LiveCategoryItem}
// @Router /api/admin/v1/live/categories/{id} [put]
func (h *LiveHandler) UpdateCategory(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateLiveCategoryRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateCategory(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteCategory
// @Summary 删除直播分类
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/categories/{id} [delete]
func (h *LiveHandler) DeleteCategory(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteCategory(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListApplies
// @Summary 主播申请列表
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param query query admindto.LiveApplyListRequest true "分页筛选"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.LiveApplyItem}}
// @Router /api/admin/v1/live/applies [get]
func (h *LiveHandler) ListApplies(c *gin.Context) {
	var req admindto.LiveApplyListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListApplies(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// ApproveApply
// @Summary 通过主播申请
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "申请ID"
// @Param body body admindto.ReviewLiveApplyRequest false "审核备注"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/applies/{id}/approve [put]
func (h *LiveHandler) ApproveApply(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.ReviewLiveApplyRequest
	if !httphandler.BindOptionalJSON(c, &req) {
		return
	}
	adminID := uint32(0)
	if claims := httpmiddleware.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}
	if err := h.svc.ApproveApply(c.Request.Context(), uri.ID, adminID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// RejectApply
// @Summary 拒绝主播申请
// @Tags 管理端｜直播管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "申请ID"
// @Param body body admindto.ReviewLiveApplyRequest false "审核备注"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/applies/{id}/reject [put]
func (h *LiveHandler) RejectApply(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.ReviewLiveApplyRequest
	if !httphandler.BindOptionalJSON(c, &req) {
		return
	}
	adminID := uint32(0)
	if claims := httpmiddleware.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}
	if err := h.svc.RejectApply(c.Request.Context(), uri.ID, adminID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListStreamers
// @Summary 主播列表
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param query query admindto.LiveStreamerListRequest true "分页筛选"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.LiveStreamerItem}}
// @Router /api/admin/v1/live/streamers [get]
func (h *LiveHandler) ListStreamers(c *gin.Context) {
	var req admindto.LiveStreamerListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListStreamers(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// GrantStreamer
// @Summary 直接设为主播
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/streamers/{userId}/grant [put]
func (h *LiveHandler) GrantStreamer(c *gin.Context) {
	var uri admindto.UserIDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	adminID := uint32(0)
	if claims := httpmiddleware.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}
	if err := h.svc.GrantStreamer(c.Request.Context(), uri.UserID, adminID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// RevokeStreamer
// @Summary 撤销主播资格
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/streamers/{userId}/revoke [put]
func (h *LiveHandler) RevokeStreamer(c *gin.Context) {
	var uri admindto.UserIDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.RevokeStreamer(c.Request.Context(), uri.UserID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListRooms
// @Summary 直播间列表
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param query query admindto.LiveRoomListRequest true "分页筛选"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.LiveRoomItem}}
// @Router /api/admin/v1/live/rooms [get]
func (h *LiveHandler) ListRooms(c *gin.Context) {
	var req admindto.LiveRoomListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListRooms(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// ForceStopRoom
// @Summary 强制下播
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "直播间ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/rooms/{id}/force-stop [post]
func (h *LiveHandler) ForceStopRoom(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.ForceStopRoom(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// BanRoom
// @Summary 禁播直播间
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "直播间ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/rooms/{id}/ban [put]
func (h *LiveHandler) BanRoom(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.BanRoom(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// UnbanRoom
// @Summary 解禁直播间
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "直播间ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/live/rooms/{id}/unban [put]
func (h *LiveHandler) UnbanRoom(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.UnbanRoom(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListSessions
// @Summary 直播场次列表
// @Tags 管理端｜直播管理
// @Produce json
// @Security BearerAuth
// @Param query query admindto.LiveSessionListRequest true "分页筛选"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.LiveSessionItem}}
// @Router /api/admin/v1/live/sessions [get]
func (h *LiveHandler) ListSessions(c *gin.Context) {
	var req admindto.LiveSessionListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListSessions(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

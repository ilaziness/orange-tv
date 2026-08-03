package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/audit"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// CommentHandler handles admin comment endpoints.
type CommentHandler struct {
	svc   adminsvc.CommentService
	audit *audit.Recorder
}

// NewCommentHandler creates a CommentHandler.
func NewCommentHandler(svc adminsvc.CommentService, recorder *audit.Recorder) *CommentHandler {
	return &CommentHandler{svc: svc, audit: recorder}
}

// List godoc
// @Summary 评论列表
// @Description 分页获取评论列表，支持按关键词、状态、影视ID筛选
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "关键词（评论内容）"
// @Param status query int false "状态：0隐藏 1正常"
// @Param video_id query int false "影视ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	var req admindto.CommentListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetLimit(), req.GetTotalPages(total))
}

// GetParents godoc
// @Summary 父级评论链
// @Description 获取指定评论的所有父级评论（从根评论到直接父评论）
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/comments/{id}/parents [get]
func (h *CommentHandler) GetParents(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	chain, err := h.svc.GetParents(c.Request.Context(), uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, chain)
}

// UpdateStatus godoc
// @Summary 更新评论状态
// @Description 审核/隐藏评论，status=1 为正常，status=0 为隐藏
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param body body admindto.UpdateCommentStatusRequest true "状态更新请求"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/comments/{id}/status [put]
func (h *CommentHandler) UpdateStatus(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateCommentStatusRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), uri.ID, &req); err != nil {
		response.Error(c, err)
		return
	}
	h.record(c, "comment_status", "update", "更新评论状态")
	response.Success(c, nil)
}

// Delete godoc
// @Summary 删除评论
// @Description 删除指定评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	h.record(c, "comment", "delete", "删除评论")
	response.Success(c, nil)
}

func (h *CommentHandler) record(c *gin.Context, resource, action, desc string) {
	if h.audit == nil {
		return
	}
	adminID := uint32(0)
	if claims := httpmiddleware.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}
	h.audit.AdminAction(c.Request.Context(), adminID, resource, action, desc, c.ClientIP())
}

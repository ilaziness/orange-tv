package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	"github.com/ilaziness/orange-tv/internal/service"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// MediaHandler handles admin media library and storage ping.
type MediaHandler struct {
	svc adminsvc.MediaService
}

// NewMediaHandler creates a MediaHandler.
func NewMediaHandler(svc adminsvc.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

// ListMedia
// @Summary 媒体库列表
// @Description 分页获取媒体库资源，可按 media_type 筛选
// @Tags 管理端｜媒体库
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query admindto.ListMediaQuery true "查询参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.MediaAsset}}
// @Router /api/admin/v1/media [get]
func (h *MediaHandler) List(c *gin.Context) {
	var q admindto.ListMediaQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), q.MediaType, q.GetOffset(), q.GetPageSize())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), q.GetPage(), q.GetPageSize(), q.GetTotalPages(total))
}

// UploadMedia
// @Summary 上传媒体
// @Description 上传图片到云存储并写入媒体库（multipart 字段 file）
// @Tags 管理端｜媒体库
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Response{data=admindto.MediaAsset}
// @Router /api/admin/v1/media [post]
func (h *MediaHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxImageUploadBytes+512*1024)
	var req admindto.UploadMediaForm
	if !httphandler.BindUploadAndValidate(c, &req) {
		return
	}
	adminID := uint32(0)
	if claims := httpmiddleware.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}
	if adminID == 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	item, err := h.svc.Upload(c.Request.Context(), adminID, req.File)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteMedia
// @Summary 删除媒体
// @Description 删除云对象与媒体库记录（不清理业务表中已引用的 URL）
// @Tags 管理端｜媒体库
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "媒体 ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/media/{id} [delete]
func (h *MediaHandler) Delete(c *gin.Context) {
	var uri admindto.MediaIDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// PingStorage
// @Summary 测试云存储连通性
// @Description 使用当前启用的云存储配置 Put 再 Delete 一个探针对象
// @Tags 管理端｜云存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=admindto.StoragePingResponse}
// @Router /api/admin/v1/storage/ping [post]
func (h *MediaHandler) Ping(c *gin.Context) {
	resp, err := h.svc.Ping(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

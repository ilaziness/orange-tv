package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	"github.com/ilaziness/orange-tv/internal/service"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// MediaHandler handles client media upload.
type MediaHandler struct {
	svc clientsvc.MediaService
}

// NewMediaHandler creates a MediaHandler.
func NewMediaHandler(svc clientsvc.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

// UploadMedia
// @Summary 上传图片
// @Description 登录用户上传图片到云存储并写入媒体库，返回加速域名 URL
// @Tags 用户端｜媒体
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Response{data=clientdto.UploadMediaResponse}
// @Router /api/client/v1/media [post]
func (h *MediaHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxImageUploadBytes+512*1024)
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.UploadMediaForm
	if !httphandler.BindUploadAndValidate(c, &req) {
		return
	}
	item, err := h.svc.Upload(c.Request.Context(), userID, req.File)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

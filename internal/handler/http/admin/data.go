package admin

import (
	"strings"

	"github.com/gin-gonic/gin"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// DataHandler handles database backup and controlled batch updates.
type DataHandler struct {
	svc adminsvc.DataService
}

// NewDataHandler creates a DataHandler.
func NewDataHandler(svc adminsvc.DataService) *DataHandler {
	return &DataHandler{svc: svc}
}

// Backup streams a full SQL dump as a downloadable file.
// @Summary 数据库备份
// @Description 导出全量 SQL 备份文件
// @Tags 管理端｜数据管理
// @Produce octet-stream
// @Param native query string false "是否使用原生备份工具：1/true 使用，0 使用 Go 实现"
// @Success 200 {file} file "SQL 备份文件"
// @Router /api/admin/v1/data/backup [get]
func (h *DataHandler) Backup(c *gin.Context) {
	var q admindto.BackupQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}

	filename := adminsvc.BackupFilename()
	c.Header("Content-Type", "application/sql")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Cache-Control", "no-cache")

	useNative := q.Native == "1" || strings.EqualFold(q.Native, "true")
	if err := h.svc.Backup(c.Request.Context(), c.Writer, useNative); err != nil {
		response.Error(c, err)
		return
	}
}

// BatchUpdatePreview returns the number of rows that would be affected.
// @Summary 批量更新预览
// @Description 预览批量替换操作的影响行数
// @Tags 管理端｜数据管理
// @Accept json
// @Produce json
// @Param body body admindto.BatchUpdatePreviewRequest true "批量更新预览请求"
// @Success 200 {object} response.Response{data=admindto.BatchUpdatePreviewResponse}
// @Router /api/admin/v1/data/batch-update/preview [post]
func (h *DataHandler) BatchUpdatePreview(c *gin.Context) {
	var req admindto.BatchUpdatePreviewRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	matched, err := h.svc.BatchUpdatePreview(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, admindto.BatchUpdatePreviewResponse{MatchedRows: matched})
}

// BatchUpdateExecute performs the batch replacement after confirmation.
// @Summary 批量更新执行
// @Description 执行批量替换操作并返回影响行数
// @Tags 管理端｜数据管理
// @Accept json
// @Produce json
// @Param body body admindto.BatchUpdateExecuteRequest true "批量更新执行请求"
// @Success 200 {object} response.Response{data=admindto.BatchUpdateExecuteResponse}
// @Router /api/admin/v1/data/batch-update/execute [post]
func (h *DataHandler) BatchUpdateExecute(c *gin.Context) {
	var req admindto.BatchUpdateExecuteRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}

	adminID := uint32(0)
	if claims := httpmiddleware.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}

	affected, err := h.svc.BatchUpdateExecute(c.Request.Context(), &req, adminID, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, admindto.BatchUpdateExecuteResponse{UpdatedRows: affected})
}

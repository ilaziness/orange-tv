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
func (h *DataHandler) Backup(c *gin.Context) {
	filename := adminsvc.BackupFilename()
	c.Header("Content-Type", "application/sql")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Cache-Control", "no-cache")

	useNative := c.Query("native") == "1" || strings.EqualFold(c.Query("native"), "true")
	if err := h.svc.Backup(c.Request.Context(), c.Writer, useNative); err != nil {
		response.Error(c, err)
		return
	}
}

// BatchUpdatePreview returns the number of rows that would be affected.
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
func (h *DataHandler) BatchUpdateExecute(c *gin.Context) {
	var req admindto.BatchUpdateExecuteRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}

	adminID := int64(0)
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

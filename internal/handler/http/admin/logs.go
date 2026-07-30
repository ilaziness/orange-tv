package admin

import (
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"

	"github.com/gin-gonic/gin"
)

// LogHandler handles admin log query endpoints.
type LogHandler struct {
	svc adminsvc.LogService
}

// NewLogHandler creates a LogHandler.
func NewLogHandler(svc adminsvc.LogService) *LogHandler {
	return &LogHandler{svc: svc}
}

func (h *LogHandler) ListSystemLogs(c *gin.Context) {
	var req admindto.SystemLogListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	items, total, err := h.svc.ListSystemLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := req.GetPage(), req.GetLimit()
	response.SuccessPage(c, items, int64(total), page, size, req.GetTotalPages(total))
}

func (h *LogHandler) ListAdminLoginLogs(c *gin.Context) {
	var req admindto.AdminLoginLogListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	items, total, err := h.svc.ListAdminLoginLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := req.GetPage(), req.GetLimit()
	response.SuccessPage(c, items, int64(total), page, size, req.GetTotalPages(total))
}

func (h *LogHandler) ListAppLogs(c *gin.Context) {
	var req admindto.AppLogListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	resp, err := h.svc.ListAppLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

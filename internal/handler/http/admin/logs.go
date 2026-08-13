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

// ListSystemLogs
// @Summary 系统日志列表
// @Description 分页获取系统操作日志
// @Tags 管理端｜日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.SystemLogListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.SystemLogItem}}
// @Router /api/admin/v1/system-logs [get]
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

// ListAdminLoginLogs
// @Summary 管理员登录日志列表
// @Description 分页获取管理员登录日志
// @Tags 管理端｜日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.AdminLoginLogListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.AdminLoginLogItem}}
// @Router /api/admin/v1/admin-login-logs [get]
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

// ListAppLogs
// @Summary 应用日志列表
// @Description 获取应用运行日志
// @Tags 管理端｜日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.AppLogListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=admindto.AppLogListResponse}
// @Router /api/admin/v1/app-logs [get]
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

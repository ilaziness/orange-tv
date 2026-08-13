package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// ManagementHandler handles dashboard, batch ops, admin/user/group/banner endpoints.
type ManagementHandler struct {
	svc adminsvc.ManagementService
}

// NewManagementHandler creates a ManagementHandler.
func NewManagementHandler(svc adminsvc.ManagementService) *ManagementHandler {
	return &ManagementHandler{svc: svc}
}

// ===== A1: Dashboard =====

// Dashboard
// @Summary 仪表盘
// @Description 获取管理后台仪表盘统计数据
// @Tags 管理端｜系统管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=admindto.DashboardResponse}
// @Router /api/admin/v1/dashboard [get]
func (h *ManagementHandler) Dashboard(c *gin.Context) {
	resp, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// ===== A2: Batch video ops =====

// BatchUpdatePublishStatus
// @Summary 批量更新发布状态
// @Description 批量更新影视的发布状态
// @Tags 管理端｜视频管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.BatchVideoRequest true "批量操作请求"
// @Success 200 {object} response.Response{data=admindto.BatchVideoResponse}
// @Router /api/admin/v1/videos/batch/publish-status [post]
func (h *ManagementHandler) BatchUpdatePublishStatus(c *gin.Context) {
	var req admindto.BatchVideoRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.BatchUpdatePublishStatus(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// BatchDeleteVideos
// @Summary 批量删除影视
// @Description 批量删除指定影视
// @Tags 管理端｜视频管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.BatchVideoRequest true "批量操作请求"
// @Success 200 {object} response.Response{data=admindto.BatchVideoResponse}
// @Router /api/admin/v1/videos/batch/delete [post]
func (h *ManagementHandler) BatchDeleteVideos(c *gin.Context) {
	var req admindto.BatchVideoRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.BatchDeleteVideos(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// ===== A3: Admin CRUD =====

// ListAdmins
// @Summary 管理员列表
// @Description 分页获取管理员列表
// @Tags 管理端｜管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.AdminListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.AdminItem}}
// @Router /api/admin/v1/admins [get]
func (h *ManagementHandler) ListAdmins(c *gin.Context) {
	var req admindto.AdminListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListAdmins(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateAdmin
// @Summary 新建管理员
// @Description 创建一个新的管理员
// @Tags 管理端｜管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateAdminRequest true "管理员参数"
// @Success 200 {object} response.Response{data=admindto.AdminItem}
// @Router /api/admin/v1/admins [post]
func (h *ManagementHandler) CreateAdmin(c *gin.Context) {
	var req admindto.CreateAdminRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateAdmin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateAdmin
// @Summary 更新管理员
// @Description 更新指定管理员信息
// @Tags 管理端｜管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "管理员ID"
// @Param body body admindto.UpdateAdminRequest true "管理员参数"
// @Success 200 {object} response.Response{data=admindto.AdminItem}
// @Router /api/admin/v1/admins/{id} [put]
func (h *ManagementHandler) UpdateAdmin(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateAdminRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateAdmin(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// ResetAdminPassword
// @Summary 重置管理员密码
// @Description 重置指定管理员的密码
// @Tags 管理端｜管理员管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "管理员ID"
// @Param body body admindto.ResetAdminPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/admins/{id}/password [put]
func (h *ManagementHandler) ResetAdminPassword(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.ResetAdminPasswordRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.ResetAdminPassword(c.Request.Context(), uri.ID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteAdmin
// @Summary 删除管理员
// @Description 删除指定管理员
// @Tags 管理端｜管理员管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "管理员ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/admins/{id} [delete]
func (h *ManagementHandler) DeleteAdmin(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteAdmin(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ===== A4: User group CRUD =====

// ListGroups
// @Summary 用户组列表
// @Description 分页获取用户组列表
// @Tags 管理端｜用户组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.UserGroupListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.UserGroupItem}}
// @Router /api/admin/v1/groups [get]
func (h *ManagementHandler) ListGroups(c *gin.Context) {
	var req admindto.UserGroupListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListGroups(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateGroup
// @Summary 新建用户组
// @Description 创建一个新的用户组
// @Tags 管理端｜用户组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateUserGroupRequest true "用户组参数"
// @Success 200 {object} response.Response{data=admindto.UserGroupItem}
// @Router /api/admin/v1/groups [post]
func (h *ManagementHandler) CreateGroup(c *gin.Context) {
	var req admindto.CreateUserGroupRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateGroup(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateGroup
// @Summary 更新用户组
// @Description 更新指定用户组信息
// @Tags 管理端｜用户组管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户组ID"
// @Param body body admindto.UpdateUserGroupRequest true "用户组参数"
// @Success 200 {object} response.Response{data=admindto.UserGroupItem}
// @Router /api/admin/v1/groups/{id} [put]
func (h *ManagementHandler) UpdateGroup(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateUserGroupRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateGroup(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteGroup
// @Summary 删除用户组
// @Description 删除指定用户组
// @Tags 管理端｜用户组管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户组ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/groups/{id} [delete]
func (h *ManagementHandler) DeleteGroup(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteGroup(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ===== A5: Regular user CRUD =====

// ListUsers
// @Summary 用户列表
// @Description 分页获取普通用户列表
// @Tags 管理端｜用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.UserListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.UserItem}}
// @Router /api/admin/v1/users [get]
func (h *ManagementHandler) ListUsers(c *gin.Context) {
	var req admindto.UserListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListUsers(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateUser
// @Summary 新建用户
// @Description 创建一个新的普通用户
// @Tags 管理端｜用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateUserRequest true "用户参数"
// @Success 200 {object} response.Response{data=admindto.UserItem}
// @Router /api/admin/v1/users [post]
func (h *ManagementHandler) CreateUser(c *gin.Context) {
	var req admindto.CreateUserRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateUser
// @Summary 更新用户
// @Description 更新指定普通用户信息
// @Tags 管理端｜用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body admindto.UpdateUserRequest true "用户参数"
// @Success 200 {object} response.Response{data=admindto.UserItem}
// @Router /api/admin/v1/users/{id} [put]
func (h *ManagementHandler) UpdateUser(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateUserRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateUser(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// ResetUserPassword
// @Summary 重置用户密码
// @Description 重置指定普通用户的密码
// @Tags 管理端｜用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body admindto.ResetUserPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/users/{id}/password [put]
func (h *ManagementHandler) ResetUserPassword(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.ResetUserPasswordRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.ResetUserPassword(c.Request.Context(), uri.ID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteUser
// @Summary 删除用户
// @Description 删除指定普通用户
// @Tags 管理端｜用户管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/users/{id} [delete]
func (h *ManagementHandler) DeleteUser(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteUser(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListUserLoginLogs
// @Summary 用户登录日志列表
// @Description 分页获取普通用户登录日志
// @Tags 管理端｜日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query admindto.UserLoginLogListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.UserLoginLogItem}}
// @Router /api/admin/v1/user-login-logs [get]
func (h *ManagementHandler) ListUserLoginLogs(c *gin.Context) {
	var req admindto.UserLoginLogListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListUserLoginLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// ===== C1: Banner CRUD =====

// ListBanners
// @Summary Banner列表
// @Description 分页获取Banner列表
// @Tags 管理端｜Banner管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req query shareddto.PaginationRequest true "分页参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]admindto.BannerItem}}
// @Router /api/admin/v1/banners [get]
func (h *ManagementHandler) ListBanners(c *gin.Context) {
	var page shareddto.PaginationRequest
	if !httphandler.BindQuery(c, &page) {
		return
	}
	list, total, err := h.svc.ListBanners(c.Request.Context(), page.GetOffset(), page.GetPageSize())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), page.GetPage(), page.GetPageSize(), page.GetTotalPages(total))
}

// CreateBanner
// @Summary 新建Banner
// @Description 创建一个新的Banner
// @Tags 管理端｜Banner管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body admindto.CreateBannerRequest true "Banner参数"
// @Success 200 {object} response.Response{data=admindto.BannerItem}
// @Router /api/admin/v1/banners [post]
func (h *ManagementHandler) CreateBanner(c *gin.Context) {
	var req admindto.CreateBannerRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateBanner(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateBanner
// @Summary 更新Banner
// @Description 更新指定Banner信息
// @Tags 管理端｜Banner管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "BannerID"
// @Param body body admindto.UpdateBannerRequest true "Banner参数"
// @Success 200 {object} response.Response{data=admindto.BannerItem}
// @Router /api/admin/v1/banners/{id} [put]
func (h *ManagementHandler) UpdateBanner(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateBannerRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateBanner(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteBanner
// @Summary 删除Banner
// @Description 删除指定Banner
// @Tags 管理端｜Banner管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "BannerID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/banners/{id} [delete]
func (h *ManagementHandler) DeleteBanner(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteBanner(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

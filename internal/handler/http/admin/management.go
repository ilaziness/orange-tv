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

func (h *ManagementHandler) Dashboard(c *gin.Context) {
	resp, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

// ===== A2: Batch video ops =====

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

func (h *ManagementHandler) ListUserLoginLogs(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var page shareddto.PaginationRequest
	if !httphandler.BindQuery(c, &page) {
		return
	}
	list, total, err := h.svc.ListUserLoginLogs(c.Request.Context(), uri.ID, page.GetOffset(), page.GetPageSize())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), page.GetPage(), page.GetPageSize(), page.GetTotalPages(total))
}

// ===== C1: Banner CRUD =====

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

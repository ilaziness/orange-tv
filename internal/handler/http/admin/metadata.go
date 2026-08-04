package admin

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// MetadataHandler handles directors/actors/tags admin endpoints.
type MetadataHandler struct {
	svc adminsvc.MetadataService
}

// NewMetadataHandler creates a MetadataHandler.
func NewMetadataHandler(svc adminsvc.MetadataService) *MetadataHandler {
	return &MetadataHandler{svc: svc}
}

// ListDirectors godoc
// @Summary 导演列表
// @Description 分页获取导演列表
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/directors [get]
func (h *MetadataHandler) ListDirectors(c *gin.Context) {
	var req admindto.NameSearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListDirectors(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateDirector godoc
// @Summary 新建导演
// @Description 创建一个新的导演
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param body body admindto.CreateNamedRequest true "导演参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/directors [post]
func (h *MetadataHandler) CreateDirector(c *gin.Context) {
	var req admindto.CreateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateDirector(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateDirector godoc
// @Summary 更新导演
// @Description 更新指定导演信息
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param id path int true "导演ID"
// @Param body body admindto.UpdateNamedRequest true "导演参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/directors/{id} [put]
func (h *MetadataHandler) UpdateDirector(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateDirector(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteDirector godoc
// @Summary 删除导演
// @Description 删除指定导演
// @Tags 管理端｜元数据管理
// @Produce json
// @Param id path int true "导演ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/directors/{id} [delete]
func (h *MetadataHandler) DeleteDirector(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteDirector(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListActors godoc
// @Summary 演员列表
// @Description 分页获取演员列表
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/actors [get]
func (h *MetadataHandler) ListActors(c *gin.Context) {
	var req admindto.NameSearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListActors(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateActor godoc
// @Summary 新建演员
// @Description 创建一个新的演员
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param body body admindto.CreateNamedRequest true "演员参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/actors [post]
func (h *MetadataHandler) CreateActor(c *gin.Context) {
	var req admindto.CreateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateActor(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateActor godoc
// @Summary 更新演员
// @Description 更新指定演员信息
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param id path int true "演员ID"
// @Param body body admindto.UpdateNamedRequest true "演员参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/actors/{id} [put]
func (h *MetadataHandler) UpdateActor(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateActor(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteActor godoc
// @Summary 删除演员
// @Description 删除指定演员
// @Tags 管理端｜元数据管理
// @Produce json
// @Param id path int true "演员ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/actors/{id} [delete]
func (h *MetadataHandler) DeleteActor(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteActor(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListTags godoc
// @Summary 标签列表
// @Description 分页获取标签列表
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/tags [get]
func (h *MetadataHandler) ListTags(c *gin.Context) {
	var req admindto.NameSearchRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.ListTags(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateTag godoc
// @Summary 新建标签
// @Description 创建一个新的标签
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param body body admindto.CreateNamedRequest true "标签参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/tags [post]
func (h *MetadataHandler) CreateTag(c *gin.Context) {
	var req admindto.CreateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateTag(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateTag godoc
// @Summary 更新标签
// @Description 更新指定标签信息
// @Tags 管理端｜元数据管理
// @Accept json
// @Produce json
// @Param id path int true "标签ID"
// @Param body body admindto.UpdateNamedRequest true "标签参数"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/tags/{id} [put]
func (h *MetadataHandler) UpdateTag(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req admindto.UpdateNamedRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateTag(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// DeleteTag godoc
// @Summary 删除标签
// @Description 删除指定标签
// @Tags 管理端｜元数据管理
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} response.Response
// @Router /api/admin/v1/tags/{id} [delete]
func (h *MetadataHandler) DeleteTag(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteTag(c.Request.Context(), uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

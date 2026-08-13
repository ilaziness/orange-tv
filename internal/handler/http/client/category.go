package client

import (
	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// CategoryHandler handles client category endpoints.
type CategoryHandler struct {
	svc clientsvc.CategoryService
}

// NewCategoryHandler creates a CategoryHandler.
func NewCategoryHandler(svc clientsvc.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// List
// @Summary 分类树
// @Description 获取公开分类树形列表
// @Tags 用户端｜分类浏览
// @Produce json
// @Success 200 {object} response.Response{data=[]clientdto.CategoryResponse}
// @Router /api/client/v1/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	var items []clientdto.CategoryResponse
	var err error
	items, err = h.svc.ListTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

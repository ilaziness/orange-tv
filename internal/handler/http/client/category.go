package client

import (
	"github.com/gin-gonic/gin"
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

func (h *CategoryHandler) List(c *gin.Context) {
	items, err := h.svc.ListTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

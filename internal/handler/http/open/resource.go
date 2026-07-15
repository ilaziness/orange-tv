package open

import (
	"strings"

	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	opensvc "github.com/ilaziness/orange-tv/internal/service/open"
)

// ResourceHandler serves resource-station open APIs.
type ResourceHandler struct {
	svc opensvc.ResourceService
}

// NewResourceHandler creates a ResourceHandler.
func NewResourceHandler(svc opensvc.ResourceService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

func extractAPIKey(c *gin.Context) string {
	key := c.GetHeader("X-API-Key")
	if key == "" {
		key = c.Query("key")
	}
	if key == "" {
		key = c.Query("api_key")
	}
	return key
}

// ListVideos handles GET /api/open/v1/videos
func (h *ResourceHandler) ListVideos(c *gin.Context) {
	cfg, err := h.svc.Authorize(c.Request.Context(), extractAPIKey(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 0)
	if pageSize == 0 {
		pageSize = queryInt(c, "limit", 20)
	}
	format := strings.TrimSpace(c.Query("format"))
	if format == "" {
		format = cfg.APIOutputFormat
	}
	data, err := h.svc.ListVideos(c.Request.Context(), page, pageSize, format)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(200, data)
}

// GetVideo handles GET /api/open/v1/videos/:id
func (h *ResourceHandler) GetVideo(c *gin.Context) {
	cfg, err := h.svc.Authorize(c.Request.Context(), extractAPIKey(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	format := strings.TrimSpace(c.Query("format"))
	if format == "" {
		format = cfg.APIOutputFormat
	}
	data, err := h.svc.GetVideo(c.Request.Context(), uri.ID, format)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(200, data)
}

// ListCategories handles GET /api/open/v1/categories
func (h *ResourceHandler) ListCategories(c *gin.Context) {
	if _, err := h.svc.Authorize(c.Request.Context(), extractAPIKey(c)); err != nil {
		response.Error(c, err)
		return
	}
	items, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

func queryInt(c *gin.Context, name string, def int) int {
	v := strings.TrimSpace(c.Query(name))
	if v == "" {
		return def
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}

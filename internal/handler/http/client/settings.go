package client

import (
	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// SettingsHandler exposes public site info and client settings.
type SettingsHandler struct {
	settings clientsvc.ClientSettingsService
}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler(settings clientsvc.ClientSettingsService) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

// GetSettings
// @Summary 获取客户端设置
// @Description 按分组获取客户端设置（支持多分组：site/feature/seo）。单分组时 data 为对应结构，多分组时 data 为 "分组名→结构" 的 map
// @Tags 用户端｜站点设置
// @Accept json
// @Produce json
// @Param q query clientdto.GetSettingsQuery true "查询参数"
// @Success 200 {object} response.Response{data=map[string]any}
// @Router /api/client/v1/settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	var q clientdto.GetSettingsQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}
	resp, err := h.settings.GetByGroups(c.Request.Context(), q.Groups)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

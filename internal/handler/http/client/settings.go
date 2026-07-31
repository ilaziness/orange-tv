package client

import (
	"github.com/gin-gonic/gin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
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

// GetSettingsQuery binds the multi-value groups query parameter.
type GetSettingsQuery struct {
	Groups []string `form:"groups" binding:"required,min=1,dive,oneof=site ad feature"`
}

// GetSettings godoc
// @Summary 获取客户端设置
// @Description 按分组获取客户端设置（支持多分组：site/ad/feature）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Param groups query []string true "设置分组 (site/ad/feature)" collectionFormat(multi)
// @Success 200 {object} response.Response
// @Router /api/client/v1/settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	var q GetSettingsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.WithMessage(errcode.ParamError, "groups 参数无效，仅支持 site/ad/feature"))
		return
	}
	resp, err := h.settings.GetByGroups(c.Request.Context(), q.Groups)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

package client

import (
	"strings"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// LiveHandler handles public live channel endpoints.
type LiveHandler struct {
	svc      clientsvc.LiveService
	proxySvc clientsvc.LiveProxyService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc clientsvc.LiveService, proxySvc clientsvc.LiveProxyService) *LiveHandler {
	return &LiveHandler{svc: svc, proxySvc: proxySvc}
}

// List
// @Summary 直播频道列表
// @Description 分页获取直播频道列表。web 端不返回 stream_url（走 /live/play/:id 代理播放）；app/tv/desktop 端（X-Client-Type: app|tv|desktop 或对应 UA）额外返回 stream_url 字段。
// @Tags 用户端｜直播观看
// @Accept json
// @Produce json
// @Param req query clientdto.LiveListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.LiveChannelItem}}
// @Router /api/client/v1/live [get]
func (h *LiveHandler) List(c *gin.Context) {
	var req clientdto.LiveListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// Play proxies a live stream by delegating to LiveProxyService.
// @Summary 直播播放代理，浏览器端有CORS限制，需通过本接口代理播放流
// @Description 代理指定直播频道的播放流
// @Tags 用户端｜直播观看
// @Produce octet-stream
// @Param id path int true "频道ID"
// @Param u query string false "分片 URL（用于代理具体分片）"
// @Success 200 {file} binary "直播流"
// @Router /api/client/v1/live/play/{id} [get]
func (h *LiveHandler) Play(c *gin.Context) {
	var req clientdto.LivePlayRequest
	if !httphandler.BindURI(c, &req) {
		return
	}
	if !httphandler.BindQuery(c, &req) {
		return
	}

	segURL := strings.TrimSpace(req.U)
	if err := h.proxySvc.Proxy(c, req.ID, segURL); err != nil {
		response.Error(c, err)
	}
}

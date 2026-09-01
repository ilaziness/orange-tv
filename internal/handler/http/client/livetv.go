package client

import (
	"strings"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// LiveTVHandler handles public livetv channel endpoints.
type LiveTVHandler struct {
	svc      clientsvc.LiveTVService
	proxySvc clientsvc.LiveTVProxyService
}

// NewLiveTVHandler creates a LiveTVHandler.
func NewLiveTVHandler(svc clientsvc.LiveTVService, proxySvc clientsvc.LiveTVProxyService) *LiveTVHandler {
	return &LiveTVHandler{svc: svc, proxySvc: proxySvc}
}

// List
// @Summary 直播频道列表
// @Description 分页获取直播频道列表。web 端不返回 stream_url（走 /livetv/play/:id 代理播放）；app/tv/desktop 端（X-Client-Type: app|tv|desktop 或对应 UA）额外返回 stream_url 字段。
// @Tags 用户端｜直播观看
// @Accept json
// @Produce json
// @Param req query clientdto.LiveTVListRequest true "筛选参数"
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.LiveTVChannelItem}}
// @Router /api/client/v1/livetv [get]
func (h *LiveTVHandler) List(c *gin.Context) {
	var req clientdto.LiveTVListRequest
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

// Play proxies a live stream by delegating to LiveTVProxyService.
// @Summary 直播播放代理，浏览器端有CORS限制，需通过本接口代理播放流
// @Description 代理指定直播频道的播放流
// @Tags 用户端｜直播观看
// @Produce octet-stream
// @Param id path int true "频道ID"
// @Param u query string false "分片 URL（用于代理具体分片）"
// @Success 200 {file} binary "直播流"
// @Router /api/client/v1/livetv/play/{id} [get]
func (h *LiveTVHandler) Play(c *gin.Context) {
	var req clientdto.LiveTVPlayRequest
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

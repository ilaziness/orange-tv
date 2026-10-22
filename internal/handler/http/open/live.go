package open

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	opensvc "github.com/ilaziness/orange-tv/internal/service/open"
)

// LiveHandler serves self-hosted SRS callbacks and Nginx play auth.
// These routes are not behind user JWT or the client live feature switch.
type LiveHandler struct {
	svc opensvc.LiveHookService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc opensvc.LiveHookService) *LiveHandler {
	return &LiveHandler{svc: svc}
}

type srsHookQuery struct {
	HookSecret string `form:"hook_secret" binding:"required,max=256"`
}

type srsHookBody struct {
	Action string `json:"action" binding:"required,max=32"`
	App    string `json:"app" binding:"omitempty,max=128"`
	Stream string `json:"stream" binding:"omitempty,max=256"`
	Param  string `json:"param" binding:"omitempty,max=2048"`
}

type playAuthQuery struct {
	HookSecret string `form:"hook_secret" binding:"required,max=256"`
}

// SRSHook
// @Summary SRS 推流/播放回调
// @Description 校验 hook_secret、应用名、直播中的流名和自建 HMAC。on_unpublish 只记日志，不结束场次。成功正文为 0。
// @Tags 开放端｜直播鉴权
// @Accept json
// @Produce plain
// @Param hook_secret query string true "回调密钥"
// @Param body body srsHookBody true "SRS http_hooks 正文"
// @Success 200 {string} string "0"
// @Failure 403 {string} string "1"
// @Router /api/open/v1/live/hooks/srs [post]
func (h *LiveHandler) SRSHook(c *gin.Context) {
	var q srsHookQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}
	var body srsHookBody
	if !httphandler.BindAndValidate(c, &body) {
		return
	}
	err := h.svc.SRSHook(c.Request.Context(), q.HookSecret, opensvc.SRSHookBody{
		Action: body.Action,
		App:    body.App,
		Stream: body.Stream,
		Param:  body.Param,
	})
	writeHook(c, err)
}

// PlayAuth
// @Summary Nginx 播放鉴权
// @Description 只校验 m3u8 与 flv。Nginx auth_request 必须把 $request_uri 放到 X-Original-URI。不要把 .ts 打到此接口。
// @Tags 开放端｜直播鉴权
// @Produce plain
// @Param hook_secret query string true "回调密钥"
// @Param X-Original-URI header string true "原始播放路径和查询串"
// @Success 200 {string} string "0"
// @Failure 403 {string} string "1"
// @Router /api/open/v1/live/auth/play [get]
func (h *LiveHandler) PlayAuth(c *gin.Context) {
	var q playAuthQuery
	if !httphandler.BindQuery(c, &q) {
		return
	}
	original := c.GetHeader("X-Original-URI")
	err := h.svc.PlayAuth(c.Request.Context(), q.HookSecret, original)
	writeHook(c, err)
}

func writeHook(c *gin.Context, err error) {
	if err == nil {
		c.String(http.StatusOK, "0")
		return
	}
	// SRS and Nginx auth_request only need a non-2xx status.
	c.String(http.StatusForbidden, "1")
}

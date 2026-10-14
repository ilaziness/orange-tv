package client

import (
	"github.com/gin-gonic/gin"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// LiveHandler handles client user-live endpoints.
type LiveHandler struct {
	svc clientsvc.LiveService
}

// NewLiveHandler creates a LiveHandler.
func NewLiveHandler(svc clientsvc.LiveService) *LiveHandler {
	return &LiveHandler{svc: svc}
}

// CreateApply
// @Summary 申请成为主播
// @Tags 用户端｜直播
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body clientdto.CreateLiveApplyRequest true "申请参数"
// @Success 200 {object} response.Response{data=clientdto.LiveApplyItem}
// @Router /api/client/v1/live/applies [post]
func (h *LiveHandler) CreateApply(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.CreateLiveApplyRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateApply(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// GetMyApply
// @Summary 我的主播申请
// @Tags 用户端｜直播
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=clientdto.LiveApplyItem}
// @Router /api/client/v1/live/applies/me [get]
func (h *LiveHandler) GetMyApply(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	item, err := h.svc.GetMyApply(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// ListCategories
// @Summary 直播分类列表
// @Tags 用户端｜直播
// @Produce json
// @Success 200 {object} response.Response{data=[]clientdto.LiveCategoryItem}
// @Router /api/client/v1/live/categories [get]
func (h *LiveHandler) ListCategories(c *gin.Context) {
	items, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, items)
}

// ListRooms
// @Summary 直播中房间列表
// @Tags 用户端｜直播
// @Produce json
// @Param query query clientdto.LiveRoomListRequest true "分页筛选"
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.LiveRoomListItem}}
// @Router /api/client/v1/live/rooms [get]
func (h *LiveHandler) ListRooms(c *gin.Context) {
	var req clientdto.LiveRoomListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListLivingRooms(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// GetRoom
// @Summary 直播间详情
// @Tags 用户端｜直播
// @Produce json
// @Param id path int true "直播间ID"
// @Success 200 {object} response.Response{data=clientdto.LiveRoomDetail}
// @Router /api/client/v1/live/rooms/{id} [get]
func (h *LiveHandler) GetRoom(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	item, err := h.svc.GetRoom(c.Request.Context(), uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// GetMyRoom
// @Summary 我的直播间
// @Tags 用户端｜直播
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=clientdto.MyLiveRoom}
// @Router /api/client/v1/live/rooms/me [get]
func (h *LiveHandler) GetMyRoom(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	item, err := h.svc.GetMyRoom(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// UpdateMyRoom
// @Summary 更新我的直播间
// @Tags 用户端｜直播
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body clientdto.UpdateMyLiveRoomRequest true "直播间资料"
// @Success 200 {object} response.Response{data=clientdto.MyLiveRoom}
// @Router /api/client/v1/live/rooms/me [put]
func (h *LiveHandler) UpdateMyRoom(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.UpdateMyLiveRoomRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.UpdateMyRoom(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// RegenerateKey
// @Summary 重置推流密钥
// @Tags 用户端｜直播
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=clientdto.MyLiveRoom}
// @Router /api/client/v1/live/rooms/me/regenerate-key [post]
func (h *LiveHandler) RegenerateKey(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	item, err := h.svc.RegenerateKey(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// Start
// @Summary 开播
// @Tags 用户端｜直播
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body clientdto.StartLiveRequest false "开播参数"
// @Success 200 {object} response.Response{data=clientdto.StartLiveResponse}
// @Router /api/client/v1/live/rooms/me/start [post]
func (h *LiveHandler) Start(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.StartLiveRequest
	if !httphandler.BindOptionalJSON(c, &req) {
		return
	}
	item, err := h.svc.Start(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

// Stop
// @Summary 停播
// @Tags 用户端｜直播
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/client/v1/live/rooms/me/stop [post]
func (h *LiveHandler) Stop(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	if err := h.svc.Stop(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

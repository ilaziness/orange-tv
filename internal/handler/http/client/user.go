package client

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/auth"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/response"
	clientsvc "github.com/ilaziness/orange-tv/internal/service/client"
)

// UserHandler handles client user auth, favorites, history, comments.
type UserHandler struct {
	svc clientsvc.UserService
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(svc clientsvc.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// ===== C5: Auth =====

func (h *UserHandler) Register(c *gin.Context) {
	var req clientdto.RegisterRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	profile, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, profile)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req clientdto.LoginRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), &req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	profile, err := h.svc.Profile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, profile)
}

// ===== C6: Favorites =====

// ListFavorites returns the current user's favorite list.
// @Summary 获取收藏列表
// @Description 分页获取当前用户的收藏列表
// @Tags client-user
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.Page{list=[]clientdto.FavoriteItem}}
// @Router /api/client/v1/favorites [get]
func (h *UserHandler) ListFavorites(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.FavoriteListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListFavorites(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// AddFavorite adds a video to the current user's favorites.
// @Summary 添加收藏
// @Description 将指定影视添加到当前用户的收藏
// @Tags client-user
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response
// @Router /api/client/v1/favorites/{id} [post]
func (h *UserHandler) AddFavorite(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.AddFavorite(c.Request.Context(), userID, uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// RemoveFavorite removes a video from the current user's favorites.
// @Summary 取消收藏
// @Description 将指定影视从当前用户的收藏中移除
// @Tags client-user
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response
// @Router /api/client/v1/favorites/{id} [delete]
func (h *UserHandler) RemoveFavorite(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.RemoveFavorite(c.Request.Context(), userID, uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// CheckFavorite checks if a video is in the current user's favorites.
// @Summary 检查收藏状态
// @Description 检查指定影视是否已被当前用户收藏
// @Tags client-user
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response{data=clientdto.FavoriteCheckResult}
// @Router /api/client/v1/favorites/{id} [get]
func (h *UserHandler) CheckFavorite(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	favorited, err := h.svc.CheckFavorite(c.Request.Context(), userID, uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, &clientdto.FavoriteCheckResult{Favorited: favorited})
}

// ===== C6: History =====

func (h *UserHandler) ListHistory(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.HistoryListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListHistory(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// GetHistory returns the current user's play history for a single video.
// @Summary 获取单条播放历史
// @Description 根据影视ID获取当前用户的播放历史（用于恢复播放进度）
// @Tags client-user
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response{data=clientdto.HistoryItem}
// @Router /api/client/v1/history/{id} [get]
func (h *UserHandler) GetHistory(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	item, err := h.svc.GetHistory(c.Request.Context(), userID, uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *UserHandler) UpsertHistory(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.UpsertHistoryRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.UpsertHistory(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) DeleteHistory(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteHistory(c.Request.Context(), userID, uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) ClearHistory(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	if err := h.svc.ClearHistory(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ===== C6: Comments =====

func (h *UserHandler) ListComments(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req clientdto.CommentListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListComments(c.Request.Context(), uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

func (h *UserHandler) CreateComment(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.CreateCommentRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.CreateComment(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *UserHandler) DeleteComment(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	if err := h.svc.DeleteComment(c.Request.Context(), userID, uri.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ===== C1: Banners (public) =====

// ListBanners is kept for backwards-compatibility; the public banner list
// is served via ClientBannerHandler (see banner.go). This handler returns
// an empty list and should not be routed anymore.
func (h *UserHandler) ListBanners(c *gin.Context) {
	response.Success(c, []any{})
}

// currentUserID extracts the user ID from JWT claims; returns 0 if not a user token.
func currentUserID(c *gin.Context) int64 {
	claims := httpmiddleware.GetClaims(c)
	if claims == nil {
		return 0
	}
	// Only accept user-subject tokens for client user endpoints
	if claims.Subject != "" && claims.Subject != auth.SubjectUser {
		return 0
	}
	return claims.UserID
}

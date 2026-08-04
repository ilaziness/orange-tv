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

// Login godoc
// @Summary 用户登录
// @Description 用户账号密码登录，返回 JWT
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param body body clientdto.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=clientdto.LoginResponse}
// @Router /api/client/v1/auth/login [post]
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

// Profile godoc
// @Summary 获取用户资料
// @Description 获取当前登录用户的个人资料
// @Tags 用户端｜用户中心
// @Produce json
// @Success 200 {object} response.Response{data=clientdto.Profile}
// @Router /api/client/v1/auth/profile [get]
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

// UpdateProfile updates the current user's profile.
// @Summary 更新用户资料
// @Description 更新当前登录用户的昵称、邮箱、头像
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param body body clientdto.UpdateProfileRequest true "用户资料"
// @Success 200 {object} response.Response{data=clientdto.Profile}
// @Router /api/client/v1/auth/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.UpdateProfileRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	profile, err := h.svc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, profile)
}

// ChangePassword changes the current user's password.
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param body body clientdto.ChangePasswordRequest true "修改密码"
// @Success 200 {object} response.Response{}
// @Router /api/client/v1/auth/profile/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.ChangePasswordRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// LoginHistory returns the current user's login history for the last 3 months.
// @Summary 获取登录历史
// @Description 分页获取当前用户最近 3 个月的登录历史
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.LoginHistoryItem}}
// @Router /api/client/v1/auth/login-history [get]
func (h *UserHandler) LoginHistory(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var req clientdto.LoginHistoryListRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	list, total, err := h.svc.LoginHistory(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// ===== C6: Favorites =====

// ListFavorites returns the current user's favorite list.
// @Summary 获取收藏列表
// @Description 分页获取当前用户的收藏列表
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.FavoriteItem}}
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
// @Tags 用户端｜用户中心
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
// @Tags 用户端｜用户中心
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
// @Tags 用户端｜用户中心
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

// ListHistory godoc
// @Summary 播放历史列表
// @Description 分页获取当前用户的播放历史
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]clientdto.HistoryItem}}
// @Router /api/client/v1/history [get]
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
// @Tags 用户端｜用户中心
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

// UpsertHistory godoc
// @Summary 上报播放进度
// @Description 新增或更新当前用户的播放历史（用于恢复播放进度）
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param body body clientdto.UpsertHistoryRequest true "播放历史请求"
// @Success 200 {object} response.Response
// @Router /api/client/v1/history [post]
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

// DeleteHistory godoc
// @Summary 删除单条播放历史
// @Description 删除当前用户指定影视的播放历史
// @Tags 用户端｜用户中心
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response
// @Router /api/client/v1/history/{id} [delete]
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

// ClearHistory godoc
// @Summary 清空播放历史
// @Description 清空当前用户的全部播放历史
// @Tags 用户端｜用户中心
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/client/v1/history [delete]
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

// ListComments godoc
// @Summary 视频评论列表
// @Description 分页获取视频顶级评论（parent_id=0），未登录时 my_vote 为 0
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/client/v1/videos/{id}/comments [get]
func (h *UserHandler) ListComments(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req clientdto.CommentListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListComments(c.Request.Context(), uri.ID, currentUserID(c), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// ListReplies godoc
// @Summary 评论回复列表
// @Description 分页获取某条评论的直接子回复
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/client/v1/comments/{id}/replies [get]
func (h *UserHandler) ListReplies(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req clientdto.CommentListRequest
	if !httphandler.BindQuery(c, &req) {
		return
	}
	list, total, err := h.svc.ListReplies(c.Request.Context(), uri.ID, currentUserID(c), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, list, int64(total), req.GetPage(), req.GetPageSize(), req.GetTotalPages(total))
}

// CreateComment godoc
// @Summary 发表评论/回复
// @Description 对视频发表评论或回复指定评论
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param body body clientdto.CreateCommentRequest true "评论请求"
// @Success 200 {object} response.Response
// @Router /api/client/v1/comments [post]
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

// VoteComment godoc
// @Summary 评论顶/踩
// @Description 对评论进行顶（like）、踩（dislike）或取消（cancel）
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param body body clientdto.VoteCommentRequest true "投票请求"
// @Success 200 {object} response.Response{data=clientdto.VoteCommentResult}
// @Router /api/client/v1/comments/{id}/vote [post]
func (h *UserHandler) VoteComment(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req clientdto.VoteCommentRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	result, err := h.svc.VoteComment(c.Request.Context(), userID, uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// ===== C6: Ratings =====

// GetRating returns the video's rating stats and the current user's score.
// @Summary 获取影视评分
// @Description 根据影视ID获取视频评分统计与当前用户评分（未登录时 my_score 为 0）
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Success 200 {object} response.Response{data=clientdto.RatingResult}
// @Router /api/client/v1/ratings/{id} [get]
func (h *UserHandler) GetRating(c *gin.Context) {
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	userID := currentUserID(c)
	result, err := h.svc.GetRating(c.Request.Context(), userID, uri.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// RateVideo submits or updates the current user's rating for a video.
// @Summary 评分影视
// @Description 当前用户对影视评分（0.5-10.0，步进 0.5），需登录
// @Tags 用户端｜用户中心
// @Accept json
// @Produce json
// @Param id path int true "影视ID"
// @Param body body clientdto.RateVideoRequest true "评分请求"
// @Success 200 {object} response.Response{data=clientdto.RatingResult}
// @Router /api/client/v1/ratings/{id} [post]
func (h *UserHandler) RateVideo(c *gin.Context) {
	userID := currentUserID(c)
	if userID <= 0 {
		response.Error(c, errcode.AuthFailed)
		return
	}
	var uri shareddto.IDURI
	if !httphandler.BindURI(c, &uri) {
		return
	}
	var req clientdto.RateVideoRequest
	if !httphandler.BindAndValidate(c, &req) {
		return
	}
	result, err := h.svc.RateVideo(c.Request.Context(), userID, uri.ID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// ===== C1: Banners (public) =====

// ListBanners is kept for backwards-compatibility; the public banner list
// is served via ClientBannerHandler (see banner.go). This handler returns
// an empty list and should not be routed anymore.
func (h *UserHandler) ListBanners(c *gin.Context) {
	response.Success(c, []any{})
}

// currentUserID extracts the user ID from JWT claims; returns 0 if not a user token.
func currentUserID(c *gin.Context) uint32 {
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

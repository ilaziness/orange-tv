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

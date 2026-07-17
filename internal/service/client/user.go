package client

import (
	"context"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/crypto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// UserService handles client user auth, favorites, history, comments.
type UserService interface {
	// C5: Auth
	Register(ctx context.Context, req *clientdto.RegisterRequest) (*clientdto.Profile, error)
	Login(ctx context.Context, req *clientdto.LoginRequest, ip, ua string) (*clientdto.LoginResponse, error)
	Profile(ctx context.Context, userID int64) (*clientdto.Profile, error)

	// C6: Favorites
	ListFavorites(ctx context.Context, userID int64, req *clientdto.FavoriteListRequest) ([]clientdto.FavoriteItem, int, error)
	AddFavorite(ctx context.Context, userID, videoID int64) error
	RemoveFavorite(ctx context.Context, userID, videoID int64) error

	// C6: History
	ListHistory(ctx context.Context, userID int64, req *clientdto.HistoryListRequest) ([]clientdto.HistoryItem, int, error)
	UpsertHistory(ctx context.Context, userID int64, req *clientdto.UpsertHistoryRequest) error
	DeleteHistory(ctx context.Context, userID, videoID int64) error
	ClearHistory(ctx context.Context, userID int64) error

	// C6: Comments
	ListComments(ctx context.Context, videoID int64, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error)
	CreateComment(ctx context.Context, userID int64, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error)
	DeleteComment(ctx context.Context, userID, commentID int64) error
}

type userService struct {
	adminRepo repository.AdminRepository
	userRepo  repository.UserFeatureRepository
	videoRepo repository.VideoRepository
	jwtMgr    *auth.JWTManager
	accessTTL int
}

// NewUserService creates a client UserService.
func NewUserService(
	adminRepo repository.AdminRepository,
	userRepo repository.UserFeatureRepository,
	videoRepo repository.VideoRepository,
	jwtMgr *auth.JWTManager,
	accessTTL int,
) UserService {
	if accessTTL <= 0 {
		accessTTL = 7200
	}
	return &userService{
		adminRepo: adminRepo,
		userRepo:  userRepo,
		videoRepo: videoRepo,
		jwtMgr:    jwtMgr,
		accessTTL: accessTTL,
	}
}

// ===== C5: Auth =====

func (s *userService) Register(ctx context.Context, req *clientdto.RegisterRequest) (*clientdto.Profile, error) {
	username := strings.TrimSpace(req.Username)
	exists, err := s.adminRepo.ExistsUserUsername(ctx, username)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.UserAlreadyExists
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	u := &model.Users{
		Username: username,
		Password: hash,
		Email:    strings.TrimSpace(req.Email),
		Status:   constant.StatusEnabled,
	}
	if err := s.adminRepo.CreateUser(ctx, u); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toUserProfile(u), nil
}

func (s *userService) Login(ctx context.Context, req *clientdto.LoginRequest, ip, ua string) (*clientdto.LoginResponse, error) {
	if s.jwtMgr == nil {
		return nil, errcode.WithMessage(errcode.ServiceUnavailable, "JWT 未配置")
	}
	username := strings.TrimSpace(req.Username)
	u, err := s.adminRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	userID := int64(0)
	if u != nil {
		userID = u.ID
	}
	recordLog := func(success bool, msg string) {
		_ = s.userRepo.CreateUserLoginLog(ctx, &model.UserLoginLogs{
			UserID:    userID,
			Username:  username,
			IP:        ip,
			UserAgent: ua,
			Status:    boolToStatus(success),
			Message:   msg,
		})
	}
	if u == nil {
		recordLog(false, "用户不存在")
		return nil, errcode.InvalidCredentials
	}
	if u.Status != constant.StatusEnabled {
		recordLog(false, "账号已禁用")
		return nil, errcode.UserDisabled
	}
	if err := crypto.CheckPassword(req.Password, u.Password); err != nil {
		recordLog(false, "密码错误")
		return nil, errcode.InvalidCredentials
	}
	token, err := s.jwtMgr.GenerateAccessTokenFor(u.ID, auth.SubjectUser)
	if err != nil {
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	now := time.Now()
	u.LastLoginAt = &now
	_ = s.adminRepo.UpdateUser(ctx, u)
	recordLog(true, "登录成功")
	return &clientdto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.accessTTL,
		User:        toUserProfile(u),
	}, nil
}

func (s *userService) Profile(ctx context.Context, userID int64) (*clientdto.Profile, error) {
	u, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return nil, errcode.UserNotFound
	}
	return toUserProfile(u), nil
}

// ===== C6: Favorites =====

func (s *userService) ListFavorites(ctx context.Context, userID int64, req *clientdto.FavoriteListRequest) ([]clientdto.FavoriteItem, int, error) {
	favs, total, err := s.userRepo.ListFavorites(ctx, userID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.FavoriteItem, 0, len(favs))
	for _, f := range favs {
		v, _ := s.videoRepo.GetByID(ctx, f.VideoID)
		item := clientdto.FavoriteItem{
			VideoID:   f.VideoID,
			CreatedAt: f.CreatedAt.Format(time.RFC3339),
		}
		if v != nil {
			item.Title = v.Title
			item.Cover = v.CoverImage
			item.Year = v.Year
			item.Rating = v.Rating
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) AddFavorite(ctx context.Context, userID, videoID int64) error {
	v, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return errcode.VideoNotFound
	}
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing != nil {
		return errcode.FavoriteExists
	}
	return s.userRepo.AddFavorite(ctx, &model.UserFavorites{
		UserID:  userID,
		VideoID: videoID,
	})
}

func (s *userService) RemoveFavorite(ctx context.Context, userID, videoID int64) error {
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing == nil {
		return errcode.FavoriteNotFound
	}
	return s.userRepo.RemoveFavorite(ctx, userID, videoID)
}

// ===== C6: History =====

func (s *userService) ListHistory(ctx context.Context, userID int64, req *clientdto.HistoryListRequest) ([]clientdto.HistoryItem, int, error) {
	items, total, err := s.userRepo.ListHistory(ctx, userID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.HistoryItem, 0, len(items))
	for _, h := range items {
		v, _ := s.videoRepo.GetByID(ctx, h.VideoID)
		item := clientdto.HistoryItem{
			VideoID:      h.VideoID,
			PlaySourceID: h.PlaySourceID,
			EpisodeID:    h.EpisodeID,
			Progress:     h.Progress,
			Duration:     h.Duration,
			LastPlayedAt: h.LastPlayedAt.Format(time.RFC3339),
		}
		if v != nil {
			item.Title = v.Title
			item.Cover = v.CoverImage
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) UpsertHistory(ctx context.Context, userID int64, req *clientdto.UpsertHistoryRequest) error {
	v, err := s.videoRepo.GetByID(ctx, req.VideoID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return errcode.VideoNotFound
	}
	now := time.Now()
	return s.userRepo.UpsertHistory(ctx, &model.UserPlayHistory{
		UserID:       userID,
		VideoID:      req.VideoID,
		PlaySourceID: req.PlaySourceID,
		EpisodeID:    req.EpisodeID,
		Progress:     req.Progress,
		Duration:     req.Duration,
		LastPlayedAt: now,
	})
}

func (s *userService) DeleteHistory(ctx context.Context, userID, videoID int64) error {
	return s.userRepo.DeleteHistory(ctx, userID, videoID)
}

func (s *userService) ClearHistory(ctx context.Context, userID int64) error {
	return s.userRepo.ClearHistory(ctx, userID)
}

// ===== C6: Comments =====

func (s *userService) ListComments(ctx context.Context, videoID int64, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error) {
	comments, total, err := s.userRepo.ListComments(ctx, videoID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.CommentItem, 0, len(comments))
	for _, c := range comments {
		item := clientdto.CommentItem{
			ID:        c.ID,
			VideoID:   c.VideoID,
			UserID:    c.UserID,
			ParentID:  c.ParentID,
			Content:   c.Content,
			LikeCount: c.LikeCount,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		}
		// Fill username
		if u, _ := s.adminRepo.GetUserByID(ctx, c.UserID); u != nil {
			item.Username = u.Username
			item.Avatar = u.Avatar
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) CreateComment(ctx context.Context, userID int64, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error) {
	v, err := s.videoRepo.GetByID(ctx, req.VideoID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return nil, errcode.VideoNotFound
	}
	if len(req.Content) > 500 {
		return nil, errcode.CommentTooLong
	}
	c := &model.VideoComments{
		VideoID:  req.VideoID,
		UserID:   userID,
		ParentID: req.ParentID,
		Content:  strings.TrimSpace(req.Content),
		Status:   1,
	}
	if err := s.userRepo.CreateComment(ctx, c); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	item := &clientdto.CommentItem{
		ID:        c.ID,
		VideoID:   c.VideoID,
		UserID:    c.UserID,
		ParentID:  c.ParentID,
		Content:   c.Content,
		LikeCount: c.LikeCount,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
	if u, _ := s.adminRepo.GetUserByID(ctx, userID); u != nil {
		item.Username = u.Username
		item.Avatar = u.Avatar
	}
	return item, nil
}

func (s *userService) DeleteComment(ctx context.Context, userID, commentID int64) error {
	c, err := s.userRepo.GetComment(ctx, commentID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.CommentNotFound
	}
	if c.UserID != userID {
		return errcode.InsufficientPermission
	}
	return s.userRepo.DeleteComment(ctx, commentID)
}

// ===== helpers =====

func toUserProfile(u *model.Users) *clientdto.Profile {
	return &clientdto.Profile{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.Avatar,
		Status:   u.Status,
	}
}

func boolToStatus(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

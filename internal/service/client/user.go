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
	"go.uber.org/zap"
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
	log       *zap.Logger
}

// NewUserService creates a client UserService.
func NewUserService(
	adminRepo repository.AdminRepository,
	userRepo repository.UserFeatureRepository,
	videoRepo repository.VideoRepository,
	jwtMgr *auth.JWTManager,
	accessTTL int,
	log *zap.Logger,
) UserService {
	if accessTTL <= 0 {
		accessTTL = 7200
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &userService{
		adminRepo: adminRepo,
		userRepo:  userRepo,
		videoRepo: videoRepo,
		jwtMgr:    jwtMgr,
		accessTTL: accessTTL,
		log:       log,
	}
}

// ===== C5: Auth =====

func (s *userService) Register(ctx context.Context, req *clientdto.RegisterRequest) (*clientdto.Profile, error) {
	username := strings.TrimSpace(req.Username)
	exists, err := s.adminRepo.ExistsUserUsername(ctx, username)
	if err != nil {
		s.log.Error("client user: check username exists failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.UserAlreadyExists
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.log.Error("client user: hash password for register failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	u := &model.Users{
		Username: username,
		Password: hash,
		Email:    strings.TrimSpace(req.Email),
		Status:   constant.StatusEnabled,
	}
	if err := s.adminRepo.CreateUser(ctx, u); err != nil {
		s.log.Error("client user: create user failed", zap.String("username", username), zap.Error(err))
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
		s.log.Error("client user: get user by username for login failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	userID := int64(0)
	if u != nil {
		userID = int64(u.ID)
	}
	recordLog := func(success bool) {
		_ = s.userRepo.CreateUserLoginLog(ctx, &model.UserLoginLogs{
			UserID:    uint64(userID),
			Username:  username,
			IP:        ip,
			UserAgent: ua,
			Status:    boolToStatus(success),
		})
	}
	if u == nil {
		recordLog(false)
		return nil, errcode.InvalidCredentials
	}
	if u.Status != constant.StatusEnabled {
		recordLog(false)
		return nil, errcode.UserDisabled
	}
	if err := crypto.CheckPassword(req.Password, u.Password); err != nil {
		recordLog(false)
		return nil, errcode.InvalidCredentials
	}
	token, err := s.jwtMgr.GenerateAccessTokenFor(int64(u.ID), auth.SubjectUser)
	if err != nil {
		s.log.Error("client user: generate access token failed", zap.Int64("user_id", int64(u.ID)), zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	now := time.Now()
	u.LastLoginAt = &now
	_ = s.adminRepo.UpdateUser(ctx, u)
	recordLog(true)
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
		s.log.Error("client user: get profile failed", zap.Int64("user_id", userID), zap.Error(err))
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
		s.log.Error("client user: list favorites failed", zap.Int64("user_id", userID), zap.Error(err))
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
			item.Year = uint32(v.Year)
			item.Rating = v.Rating
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) AddFavorite(ctx context.Context, userID, videoID int64) error {
	v, err := s.videoRepo.GetByID(ctx, uint64(videoID))
	if err != nil {
		s.log.Error("client user: get video for add favorite failed", zap.Int64("video_id", videoID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return errcode.VideoNotFound
	}
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: get favorite failed", zap.Int64("user_id", userID), zap.Int64("video_id", videoID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing != nil {
		return errcode.FavoriteExists
	}
	return s.userRepo.AddFavorite(ctx, &model.UserFavorites{
		UserID:  uint64(userID),
		VideoID: uint64(videoID),
	})
}

func (s *userService) RemoveFavorite(ctx context.Context, userID, videoID int64) error {
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: get favorite for remove failed", zap.Int64("user_id", userID), zap.Int64("video_id", videoID), zap.Error(err))
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
		s.log.Error("client user: list history failed", zap.Int64("user_id", userID), zap.Error(err))
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
		s.log.Error("client user: get video for upsert history failed", zap.Uint64("video_id", req.VideoID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return errcode.VideoNotFound
	}
	now := time.Now()
	return s.userRepo.UpsertHistory(ctx, &model.UserPlayHistory{
		UserID:       uint64(userID),
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
		s.log.Error("client user: list comments failed", zap.Int64("video_id", videoID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.CommentItem, 0, len(comments))
	for _, c := range comments {
		item := clientdto.CommentItem{
			ID:           c.ID,
			VideoID:      c.VideoID,
			UserID:       c.UserID,
			ParentID:     c.ParentID,
			Content:      c.Content,
			LikeCount:    c.LikeCount,
			DislikeCount: c.DislikeCount,
			CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		}
		if c.User != nil {
			item.Username = c.User.Username
			item.Avatar = c.User.Avatar
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) CreateComment(ctx context.Context, userID int64, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error) {
	v, err := s.videoRepo.GetByID(ctx, req.VideoID)
	if err != nil {
		s.log.Error("client user: get video for create comment failed", zap.Uint64("video_id", req.VideoID), zap.Error(err))
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
		UserID:   uint64(userID),
		ParentID: req.ParentID,
		Content:  strings.TrimSpace(req.Content),
		Status:   0,
	}
	if err := s.userRepo.CreateComment(ctx, c); err != nil {
		s.log.Error("client user: create comment failed", zap.Uint64("video_id", req.VideoID), zap.Int64("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	item := &clientdto.CommentItem{
		ID:           c.ID,
		VideoID:      c.VideoID,
		UserID:       c.UserID,
		ParentID:     c.ParentID,
		Content:      c.Content,
		LikeCount:    c.LikeCount,
		DislikeCount: c.DislikeCount,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
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
		s.log.Error("client user: get comment for delete failed", zap.Int64("comment_id", commentID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.CommentNotFound
	}
	if c.UserID != uint64(userID) {
		return errcode.InsufficientPermission
	}
	return s.userRepo.DeleteComment(ctx, commentID)
}

// ===== helpers =====

func toUserProfile(u *model.Users) *clientdto.Profile {
	return &clientdto.Profile{
		ID:       uint64(u.ID),
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.Avatar,
		Status:   uint8(u.Status),
	}
}

func boolToStatus(b bool) uint8 {
	if b {
		return constant.LoginStatusSuccess
	}
	return constant.LoginStatusFailed
}

package client

import (
	"context"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/crypto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// htmlTagRegex matches both opening and closing HTML tags.
var htmlTagRegex = regexp.MustCompile(`<(/)?[a-zA-Z][^>]*>`)

// UserService handles client user auth, favorites, history, comments.
type UserService interface {
	// C5: Auth
	Register(ctx context.Context, req *clientdto.RegisterRequest) (*clientdto.Profile, error)
	Login(ctx context.Context, req *clientdto.LoginRequest, ip, ua string) (*clientdto.LoginResponse, error)
	Profile(ctx context.Context, userID int64) (*clientdto.Profile, error)
	UpdateProfile(ctx context.Context, userID int64, req *clientdto.UpdateProfileRequest) (*clientdto.Profile, error)
	ChangePassword(ctx context.Context, userID int64, req *clientdto.ChangePasswordRequest) error
	LoginHistory(ctx context.Context, userID int64, req *clientdto.LoginHistoryListRequest) ([]clientdto.LoginHistoryItem, int, error)

	// C6: Favorites
	ListFavorites(ctx context.Context, userID int64, req *clientdto.FavoriteListRequest) ([]clientdto.FavoriteItem, int, error)
	AddFavorite(ctx context.Context, userID, videoID int64) error
	RemoveFavorite(ctx context.Context, userID, videoID int64) error
	CheckFavorite(ctx context.Context, userID, videoID int64) (bool, error)

	// C6: History
	ListHistory(ctx context.Context, userID int64, req *clientdto.HistoryListRequest) ([]clientdto.HistoryItem, int, error)
	GetHistory(ctx context.Context, userID, videoID int64) (*clientdto.HistoryItem, error)
	UpsertHistory(ctx context.Context, userID int64, req *clientdto.UpsertHistoryRequest) error
	DeleteHistory(ctx context.Context, userID, videoID int64) error
	ClearHistory(ctx context.Context, userID int64) error

	// C6: Comments
	ListComments(ctx context.Context, videoID int64, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error)
	CreateComment(ctx context.Context, userID int64, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error)
	DeleteComment(ctx context.Context, userID, commentID int64) error

	// C6: Ratings
	RateVideo(ctx context.Context, userID, videoID int64, req *clientdto.RateVideoRequest) (*clientdto.RatingResult, error)
	GetRating(ctx context.Context, userID, videoID int64) (*clientdto.RatingResult, error)
}

type userService struct {
	adminRepo    repository.AdminRepository
	userRepo     repository.UserFeatureRepository
	videoRepo    repository.VideoRepository
	categoryRepo repository.CategoryRepository
	jwtMgr       *auth.JWTManager
	accessTTL    int
	settingsSvc  service.SettingsService
	log          *zap.Logger
}

// NewUserService creates a client UserService.
func NewUserService(
	adminRepo repository.AdminRepository,
	userRepo repository.UserFeatureRepository,
	videoRepo repository.VideoRepository,
	categoryRepo repository.CategoryRepository,
	jwtMgr *auth.JWTManager,
	accessTTL int,
	settingsSvc service.SettingsService,
	log *zap.Logger,
) UserService {
	if accessTTL <= 0 {
		accessTTL = 7200
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &userService{
		adminRepo:    adminRepo,
		userRepo:     userRepo,
		videoRepo:    videoRepo,
		categoryRepo: categoryRepo,
		jwtMgr:       jwtMgr,
		accessTTL:    accessTTL,
		settingsSvc:  settingsSvc,
		log:          log,
	}
}

// ===== C5: Auth =====

func (s *userService) Register(ctx context.Context, req *clientdto.RegisterRequest) (*clientdto.Profile, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	exists, err := s.adminRepo.ExistsUserUsername(ctx, username)
	if err != nil {
		s.log.Error("client user: check username exists failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.UserAlreadyExists
	}
	hash, err := crypto.HashPassword(password)
	if err != nil {
		s.log.Error("client user: hash password for register failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	strID, err := utils.GenerateUniqueNumericID(ctx, 10, s.adminRepo.ExistsUserStrID)
	if err != nil {
		s.log.Error("client user: generate str_id for register failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	u := &model.Users{
		Username: username,
		Password: hash,
		Email:    strings.TrimSpace(req.Email),
		Status:   constant.StatusEnabled,
		StrID:    strID,
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
	password := strings.TrimSpace(req.Password)
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
		if userID > 0 {
			_ = s.userRepo.DeleteUserLoginLogsBefore(ctx, userID, time.Now().AddDate(0, -3, 0))
		}
	}
	if u == nil {
		recordLog(false)
		return nil, errcode.InvalidCredentials
	}
	if u.Status != constant.StatusEnabled {
		recordLog(false)
		return nil, errcode.UserDisabled
	}
	if err := crypto.CheckPassword(password, u.Password); err != nil {
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

func (s *userService) UpdateProfile(ctx context.Context, userID int64, req *clientdto.UpdateProfileRequest) (*clientdto.Profile, error) {
	u, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("client user: get user for update profile failed", zap.Int64("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return nil, errcode.UserNotFound
	}
	if nickname := strings.TrimSpace(req.Nickname); nickname != "" && nickname != u.Nickname {
		u.Nickname = nickname
	}
	if email := strings.TrimSpace(req.Email); email != "" && email != u.Email {
		u.Email = email
	}
	if avatar := strings.TrimSpace(req.Avatar); avatar != "" && avatar != u.Avatar {
		u.Avatar = avatar
	}
	if err := s.adminRepo.UpdateUser(ctx, u); err != nil {
		s.log.Error("client user: update profile failed", zap.Int64("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toUserProfile(u), nil
}

func (s *userService) ChangePassword(ctx context.Context, userID int64, req *clientdto.ChangePasswordRequest) error {
	u, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("client user: get user for change password failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return errcode.UserNotFound
	}
	if err := crypto.CheckPassword(req.CurrentPassword, u.Password); err != nil {
		return errcode.InvalidCredentials
	}
	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("client user: hash new password failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}
	u.Password = hash
	if err := s.adminRepo.UpdateUser(ctx, u); err != nil {
		s.log.Error("client user: update password failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *userService) LoginHistory(ctx context.Context, userID int64, req *clientdto.LoginHistoryListRequest) ([]clientdto.LoginHistoryItem, int, error) {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	items, total, err := s.userRepo.ListUserLoginLogs(ctx, repository.UserLoginLogFilter{
		UserID:    &userID,
		StartTime: &threeMonthsAgo,
		Offset:    req.GetOffset(),
		Limit:     req.GetPageSize(),
	})
	if err != nil {
		s.log.Error("client user: list login history failed", zap.Int64("user_id", userID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.LoginHistoryItem, 0, len(items))
	for _, item := range items {
		out = append(out, clientdto.LoginHistoryItem{
			ID:        item.ID,
			IP:        item.IP,
			UserAgent: item.UserAgent,
			Status:    item.Status,
			CreatedAt: utils.FormatTimeStr(&item.CreatedAt),
		})
	}
	return out, total, nil
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
			CreatedAt: utils.FormatTimeStr(&f.CreatedAt),
		}
		if v != nil {
			item.Title = v.Title
			item.Cover = v.CoverImage
			item.Year = uint32(v.Year)
			item.Rating = v.Rating
			if v.CategoryID > 0 {
				cat, _ := s.categoryRepo.GetByID(ctx, int64(v.CategoryID))
				if cat != nil {
					item.CategoryName = cat.Name
				}
			}
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

func (s *userService) CheckFavorite(ctx context.Context, userID, videoID int64) (bool, error) {
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: check favorite failed", zap.Int64("user_id", userID), zap.Int64("video_id", videoID), zap.Error(err))
		return false, errcode.Wrap(errcode.DatabaseError, err)
	}
	return existing != nil, nil
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
			LastPlayedAt: utils.FormatTimeStr(&h.LastPlayedAt),
		}
		if v != nil {
			item.Title = v.Title
			item.Cover = v.CoverImage
			if v.Year > 0 {
				item.Year = strconv.FormatUint(uint64(v.Year), 10)
			}
			if v.CategoryID > 0 {
				if cat, _ := s.categoryRepo.GetByID(ctx, int64(v.CategoryID)); cat != nil {
					item.CategoryName = cat.Name
				}
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) GetHistory(ctx context.Context, userID, videoID int64) (*clientdto.HistoryItem, error) {
	h, err := s.userRepo.GetHistory(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: get history failed", zap.Int64("user_id", userID), zap.Int64("video_id", videoID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if h == nil {
		return nil, errcode.HistoryNotFound
	}
	item := clientdto.HistoryItem{
		VideoID:      h.VideoID,
		PlaySourceID: h.PlaySourceID,
		EpisodeID:    h.EpisodeID,
		Progress:     h.Progress,
		Duration:     h.Duration,
		LastPlayedAt: utils.FormatTimeStr(&h.LastPlayedAt),
	}
	v, _ := s.videoRepo.GetByID(ctx, h.VideoID)
	if v != nil {
		item.Title = v.Title
		item.Cover = v.CoverImage
		if v.Year > 0 {
			item.Year = strconv.FormatUint(uint64(v.Year), 10)
		}
		if v.CategoryID > 0 {
			if cat, _ := s.categoryRepo.GetByID(ctx, int64(v.CategoryID)); cat != nil {
				item.CategoryName = cat.Name
			}
		}
	}
	return &item, nil
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
			CreatedAt:    utils.FormatTimeStr(&c.CreatedAt),
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
	// Check feature toggles
	featureMap, err := s.settingsSvc.LoadMapByGroup(ctx, constant.SettingGroupFeature)
	if err != nil {
		s.log.Error("client user: load feature settings for create comment failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if !service.BoolVal(featureMap, constant.SettingFeatureCommentEnabled, true) {
		return nil, errcode.CommentDisabled
	}

	v, err := s.videoRepo.GetByID(ctx, req.VideoID)
	if err != nil {
		s.log.Error("client user: get video for create comment failed", zap.Uint64("video_id", req.VideoID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return nil, errcode.VideoNotFound
	}
	content := strings.TrimSpace(htmlTagRegex.ReplaceAllString(html.UnescapeString(req.Content), ""))
	if content == "" {
		return nil, errcode.ParamError
	}
	if utf8.RuneCountInString(content) > 200 {
		return nil, errcode.CommentTooLong
	}

	// Determine comment status based on review toggle
	status := constant.CommentStatusNormal
	if service.BoolVal(featureMap, constant.SettingFeatureCommentReview, true) {
		status = constant.CommentStatusHidden
	}

	c := &model.VideoComments{
		VideoID:  req.VideoID,
		UserID:   uint64(userID),
		ParentID: req.ParentID,
		Content:  content,
		Status:   status,
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
		CreatedAt:    utils.FormatTimeStr(&c.CreatedAt),
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

// ===== Ratings (C6) =====

// RateVideo submits or updates a user's rating for a video.
func (s *userService) RateVideo(ctx context.Context, userID, videoID int64, req *clientdto.RateVideoRequest) (*clientdto.RatingResult, error) {
	// Check feature toggle
	featureMap, err := s.settingsSvc.LoadMapByGroup(ctx, constant.SettingGroupFeature)
	if err != nil {
		s.log.Error("client user: load feature settings for rate video failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if !service.BoolVal(featureMap, constant.SettingFeatureRatingEnabled, true) {
		return nil, errcode.RatingDisabled
	}

	// Validate score: 0.5-10.0 in 0.5 increments
	score := req.Score
	doubled := score * 2
	if doubled != float64(int(doubled)) || doubled < 1 || doubled > 20 {
		return nil, errcode.RatingInvalid
	}

	// Validate video exists
	v, err := s.videoRepo.GetByID(ctx, uint64(videoID))
	if err != nil {
		s.log.Error("client user: get video for rate failed", zap.Int64("video_id", videoID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return nil, errcode.VideoNotFound
	}

	// Upsert rating + recompute stats + update video in a transaction
	var avg float64
	var count int
	err = s.userRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txUserRepo := s.userRepo.WithTx(tx)
		txVideoRepo := s.videoRepo.WithTx(tx)

		ur := &model.UserRatings{
			UserID:  uint64(userID),
			VideoID: uint64(videoID),
			Score:   score,
		}
		if err := txUserRepo.UpsertRating(ctx, ur); err != nil {
			s.log.Error("client user: upsert rating failed", zap.Int64("video_id", videoID), zap.Int64("user_id", userID), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, err)
		}

		avg, count, err = txUserRepo.GetRatingStats(ctx, videoID)
		if err != nil {
			s.log.Error("client user: get rating stats failed", zap.Int64("video_id", videoID), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, err)
		}

		// Round to 1 decimal place
		rounded := math.Round(avg*10) / 10
		if err := txVideoRepo.UpdateRatingStats(ctx, uint64(videoID), rounded, uint32(count)); err != nil {
			s.log.Error("client user: update video rating stats failed", zap.Int64("video_id", videoID), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &clientdto.RatingResult{
		MyScore:     score,
		Rating:      math.Round(avg*10) / 10,
		RatingCount: uint32(count),
	}, nil
}

// GetRating returns the video's rating stats and the current user's score (0 if not logged in / not rated).
func (s *userService) GetRating(ctx context.Context, userID, videoID int64) (*clientdto.RatingResult, error) {
	v, err := s.videoRepo.GetByID(ctx, uint64(videoID))
	if err != nil {
		s.log.Error("client user: get video for get rating failed", zap.Int64("video_id", videoID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return nil, errcode.VideoNotFound
	}

	result := &clientdto.RatingResult{
		MyScore:     0,
		Rating:      v.Rating,
		RatingCount: v.RatingCount,
	}

	if userID > 0 {
		rating, err := s.userRepo.GetRating(ctx, userID, videoID)
		if err != nil {
			s.log.Error("client user: get user rating failed", zap.Int64("video_id", videoID), zap.Int64("user_id", userID), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if rating != nil {
			result.MyScore = rating.Score
		}
	}

	return result, nil
}

// ===== helpers =====

func toUserProfile(u *model.Users) *clientdto.Profile {
	return &clientdto.Profile{
		ID:       uint64(u.ID),
		StrID:    u.StrID,
		Username: u.Username,
		Nickname: u.Nickname,
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

package client

import (
	"context"
	"errors"
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
	internallock "github.com/ilaziness/orange-tv/internal/lock"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/utils"
	pkglock "github.com/ilaziness/orange-tv/pkg/lock"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// htmlTagRegex matches both opening and closing HTML tags.
var htmlTagRegex = regexp.MustCompile(`<(/)?[a-zA-Z][^>]*>`)

// registerLockTTL 注册接口邮箱锁的过期时间。
// 注册流程通常在秒级完成，10s 足以覆盖正常处理；
// 锁失效后 DB 唯一索引仍可兜底，避免重复插入。
const registerLockTTL = 10 * time.Second

// UserService handles client user auth, favorites, history, comments.
type UserService interface {
	// C5: Auth
	Register(ctx context.Context, req *clientdto.RegisterRequest) error
	Login(ctx context.Context, req *clientdto.LoginRequest, ip, ua string) (*clientdto.LoginResponse, error)
	Profile(ctx context.Context, userID uint32) (*clientdto.Profile, error)
	UpdateProfile(ctx context.Context, userID uint32, req *clientdto.UpdateProfileRequest) (*clientdto.Profile, error)
	ChangePassword(ctx context.Context, userID uint32, req *clientdto.ChangePasswordRequest) error
	LoginHistory(ctx context.Context, userID uint32, req *clientdto.LoginHistoryListRequest) ([]clientdto.LoginHistoryItem, int, error)

	// C6: Favorites
	ListFavorites(ctx context.Context, userID uint32, req *clientdto.FavoriteListRequest) ([]clientdto.FavoriteItem, int, error)
	AddFavorite(ctx context.Context, userID, videoID uint32) error
	RemoveFavorite(ctx context.Context, userID, videoID uint32) error
	CheckFavorite(ctx context.Context, userID, videoID uint32) (bool, error)

	// C6: History
	ListHistory(ctx context.Context, userID uint32, req *clientdto.HistoryListRequest) ([]clientdto.HistoryItem, int, error)
	GetHistory(ctx context.Context, userID, videoID uint32) (*clientdto.HistoryItem, error)
	UpsertHistory(ctx context.Context, userID uint32, req *clientdto.UpsertHistoryRequest) error
	DeleteHistory(ctx context.Context, userID, videoID uint32) error
	ClearHistory(ctx context.Context, userID uint32) error

	// C6: Comments
	ListComments(ctx context.Context, videoID, userID uint32, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error)
	ListReplies(ctx context.Context, commentID, userID uint32, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error)
	CreateComment(ctx context.Context, userID uint32, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error)
	VoteComment(ctx context.Context, userID, commentID uint32, req *clientdto.VoteCommentRequest) (*clientdto.VoteCommentResult, error)

	// C6: Ratings
	RateVideo(ctx context.Context, userID, videoID uint32, req *clientdto.RateVideoRequest) (*clientdto.RatingResult, error)
	GetRating(ctx context.Context, userID, videoID uint32) (*clientdto.RatingResult, error)
}

type userService struct {
	adminRepo    repository.AdminRepository
	userRepo     repository.UserFeatureRepository
	videoRepo    repository.VideoRepository
	categoryRepo repository.CategoryRepository
	jwtMgr       *auth.JWTManager
	accessTTL    int
	settingsSvc  service.SettingsService
	locker       pkglock.Locker
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
	locker pkglock.Locker,
	log *zap.Logger,
) UserService {
	if accessTTL <= 0 {
		accessTTL = 7200
	}
	return &userService{
		adminRepo:    adminRepo,
		userRepo:     userRepo,
		videoRepo:    videoRepo,
		categoryRepo: categoryRepo,
		jwtMgr:       jwtMgr,
		accessTTL:    accessTTL,
		settingsSvc:  settingsSvc,
		locker:       locker,
		log:          log,
	}
}

// ===== C5: Auth =====

func (s *userService) Register(ctx context.Context, req *clientdto.RegisterRequest) error {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := strings.TrimSpace(req.Password)

	// 邮箱维度加分布式锁，防并发同邮箱注册产生两次 ExistsUserEmail 都通过的竞态。
	// 锁被其他请求持有（ErrLockNotHeld）视为并发同邮箱注册，返回 UserAlreadyExists。
	// 其它锁故障（网络/Redis 异常等）按操作频繁处理，避免暴露底层细节。
	lockKey := internallock.UserRegisterKey(email)
	lock, err := s.locker.Lock(ctx, lockKey, pkglock.WithTTL(registerLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return errcode.UserAlreadyExists
	}
	if err != nil {
		s.log.Error("client user: acquire register lock failed", zap.String("email", email), zap.Error(err))
		return errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	exists, err := s.adminRepo.ExistsUserEmail(ctx, email)
	if err != nil {
		s.log.Error("client user: check email exists failed", zap.String("email", email), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return errcode.UserAlreadyExists
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		s.log.Error("client user: hash password for register failed", zap.String("email", email), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}
	strID, err := utils.GenerateUniqueNumericID(ctx, 10, s.adminRepo.ExistsUserStrID)
	if err != nil {
		s.log.Error("client user: generate str_id for register failed", zap.String("email", email), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = deriveNicknameFromEmail(email)
	}

	u := &model.Users{
		Password: hash,
		Email:    email,
		Nickname: nickname,
		Status:   constant.StatusEnabled,
		StrID:    strID,
	}
	if err := s.adminRepo.CreateUser(ctx, u); err != nil {
		s.log.Error("client user: create user failed", zap.String("email", email), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// deriveNicknameFromEmail 从邮箱推导默认昵称：取 @ 前部分，按 rune 截断到 15 字符。
// 无 @ 或空前缀时回退用完整 email 截断 15 字符，匹配 DB nickname VARCHAR(15)。
func deriveNicknameFromEmail(email string) string {
	prefix := email
	if idx := strings.Index(email, "@"); idx > 0 {
		prefix = email[:idx]
	}
	if prefix == "" {
		prefix = email
	}
	r := []rune(prefix)
	if len(r) > 15 {
		r = r[:15]
	}
	return string(r)
}

func (s *userService) Login(ctx context.Context, req *clientdto.LoginRequest, ip, ua string) (*clientdto.LoginResponse, error) {
	if s.jwtMgr == nil {
		return nil, errcode.WithMessage(errcode.ServiceUnavailable, "JWT 未配置")
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := strings.TrimSpace(req.Password)
	u, err := s.adminRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Error("client user: get user by email for login failed", zap.String("email", email), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	userID := uint32(0)
	if u != nil {
		userID = u.ID
	}
	recordLog := func(success bool) {
		_ = s.userRepo.CreateUserLoginLog(ctx, &model.UserLoginLogs{
			UserID:    userID,
			Email:     email,
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
	if pwdErr := crypto.CheckPassword(password, u.Password); pwdErr != nil {
		recordLog(false)
		return nil, errcode.InvalidCredentials
	}
	token, err := s.jwtMgr.GenerateAccessTokenFor(u.ID, auth.SubjectUser)
	if err != nil {
		s.log.Error("client user: generate access token failed", zap.Uint32("user_id", u.ID), zap.Error(err))
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

func (s *userService) Profile(ctx context.Context, userID uint32) (*clientdto.Profile, error) {
	u, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("client user: get profile failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return nil, errcode.UserNotFound
	}
	return toUserProfile(u), nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uint32, req *clientdto.UpdateProfileRequest) (*clientdto.Profile, error) {
	u, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("client user: get user for update profile failed", zap.Uint32("user_id", userID), zap.Error(err))
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
		s.log.Error("client user: update profile failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toUserProfile(u), nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uint32, req *clientdto.ChangePasswordRequest) error {
	u, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("client user: get user for change password failed", zap.Uint32("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return errcode.UserNotFound
	}
	if pwdErr := crypto.CheckPassword(req.CurrentPassword, u.Password); pwdErr != nil {
		return errcode.InvalidCredentials
	}
	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("client user: hash new password failed", zap.Uint32("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}
	u.Password = hash
	if err := s.adminRepo.UpdateUser(ctx, u); err != nil {
		s.log.Error("client user: update password failed", zap.Uint32("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *userService) LoginHistory(ctx context.Context, userID uint32, req *clientdto.LoginHistoryListRequest) ([]clientdto.LoginHistoryItem, int, error) {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	items, total, err := s.userRepo.ListUserLoginLogs(ctx, repository.UserLoginLogFilter{
		UserID:    &userID,
		StartTime: &threeMonthsAgo,
		Offset:    req.GetOffset(),
		Limit:     req.GetPageSize(),
	})
	if err != nil {
		s.log.Error("client user: list login history failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.LoginHistoryItem, 0, len(items))
	for _, item := range items {
		out = append(out, clientdto.LoginHistoryItem{
			ID:        item.ID,
			IP:        item.IP,
			UserAgent: item.UserAgent,
			Status:    item.Status,
			CreatedAt: utils.FormatTimeStr(item.CreatedAt),
		})
	}
	return out, total, nil
}

// ===== C6: Favorites =====

func (s *userService) ListFavorites(ctx context.Context, userID uint32, req *clientdto.FavoriteListRequest) ([]clientdto.FavoriteItem, int, error) {
	favs, total, err := s.userRepo.ListFavorites(ctx, userID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		s.log.Error("client user: list favorites failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.FavoriteItem, 0, len(favs))
	for _, f := range favs {
		v, _ := s.videoRepo.GetByID(ctx, f.VideoID)
		item := clientdto.FavoriteItem{
			VideoID:   f.VideoID,
			CreatedAt: utils.FormatTimeStr(f.CreatedAt),
		}
		if v != nil {
			item.Title = v.Title
			item.Cover = v.CoverImage
			item.Year = v.Year
			item.Rating = v.Rating
			if v.CategoryID > 0 {
				cat, _ := s.categoryRepo.GetByID(ctx, v.CategoryID)
				if cat != nil {
					item.CategoryName = cat.Name
				}
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) AddFavorite(ctx context.Context, userID, videoID uint32) error {
	v, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("client user: get video for add favorite failed", zap.Uint32("video_id", videoID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if v == nil {
		return errcode.VideoNotFound
	}
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: get favorite failed", zap.Uint32("user_id", userID), zap.Uint32("video_id", videoID), zap.Error(err))
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

func (s *userService) RemoveFavorite(ctx context.Context, userID, videoID uint32) error {
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: get favorite for remove failed", zap.Uint32("user_id", userID), zap.Uint32("video_id", videoID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing == nil {
		return errcode.FavoriteNotFound
	}
	return s.userRepo.RemoveFavorite(ctx, userID, videoID)
}

func (s *userService) CheckFavorite(ctx context.Context, userID, videoID uint32) (bool, error) {
	existing, err := s.userRepo.GetFavorite(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: check favorite failed", zap.Uint32("user_id", userID), zap.Uint32("video_id", videoID), zap.Error(err))
		return false, errcode.Wrap(errcode.DatabaseError, err)
	}
	return existing != nil, nil
}

// ===== C6: History =====

func (s *userService) ListHistory(ctx context.Context, userID uint32, req *clientdto.HistoryListRequest) ([]clientdto.HistoryItem, int, error) {
	items, total, err := s.userRepo.ListHistory(ctx, userID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		s.log.Error("client user: list history failed", zap.Uint32("user_id", userID), zap.Error(err))
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
			LastPlayedAt: utils.FormatTimeStr(h.LastPlayedAt),
		}
		if v != nil {
			item.Title = v.Title
			item.Cover = v.CoverImage
			if v.Year > 0 {
				item.Year = strconv.FormatUint(uint64(v.Year), 10)
			}
			if v.CategoryID > 0 {
				if cat, _ := s.categoryRepo.GetByID(ctx, v.CategoryID); cat != nil {
					item.CategoryName = cat.Name
				}
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *userService) GetHistory(ctx context.Context, userID, videoID uint32) (*clientdto.HistoryItem, error) {
	h, err := s.userRepo.GetHistory(ctx, userID, videoID)
	if err != nil {
		s.log.Error("client user: get history failed", zap.Uint32("user_id", userID), zap.Uint32("video_id", videoID), zap.Error(err))
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
		LastPlayedAt: utils.FormatTimeStr(h.LastPlayedAt),
	}
	v, _ := s.videoRepo.GetByID(ctx, h.VideoID)
	if v != nil {
		item.Title = v.Title
		item.Cover = v.CoverImage
		if v.Year > 0 {
			item.Year = strconv.FormatUint(uint64(v.Year), 10)
		}
		if v.CategoryID > 0 {
			if cat, _ := s.categoryRepo.GetByID(ctx, v.CategoryID); cat != nil {
				item.CategoryName = cat.Name
			}
		}
	}
	return &item, nil
}

func (s *userService) UpsertHistory(ctx context.Context, userID uint32, req *clientdto.UpsertHistoryRequest) error {
	v, err := s.videoRepo.GetByID(ctx, req.VideoID)
	if err != nil {
		s.log.Error("client user: get video for upsert history failed", zap.Uint32("video_id", req.VideoID), zap.Error(err))
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

func (s *userService) DeleteHistory(ctx context.Context, userID, videoID uint32) error {
	return s.userRepo.DeleteHistory(ctx, userID, videoID)
}

func (s *userService) ClearHistory(ctx context.Context, userID uint32) error {
	return s.userRepo.ClearHistory(ctx, userID)
}

// ===== C6: Comments =====

func (s *userService) mapComments(ctx context.Context, comments []model.VideoComments, userID uint32) ([]clientdto.CommentItem, error) {
	if len(comments) == 0 {
		return []clientdto.CommentItem{}, nil
	}
	ids := make([]uint32, len(comments))
	for i, c := range comments {
		ids[i] = c.ID
	}
	replyCounts, err := s.userRepo.CountRepliesByParents(ctx, ids)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	myVotes, err := s.userRepo.BatchGetCommentVotes(ctx, userID, ids)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
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
			MyVote:       myVotes[c.ID],
			ReplyCount:   replyCounts[c.ID],
			Replies:      make([]*clientdto.CommentItem, 0),
			CreatedAt:    utils.FormatTimeStr(c.CreatedAt),
		}
		if c.User != nil {
			item.Nickname = c.User.Nickname
			item.Avatar = c.User.Avatar
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *userService) ListComments(ctx context.Context, videoID, userID uint32, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error) {
	comments, total, err := s.userRepo.ListComments(ctx, videoID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		s.log.Error("client user: list comments failed", zap.Uint32("video_id", videoID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out, err := s.mapComments(ctx, comments, userID)
	if err != nil {
		s.log.Error("client user: map comments failed", zap.Uint32("video_id", videoID), zap.Error(err))
		return nil, 0, err
	}
	return out, total, nil
}

func (s *userService) ListReplies(ctx context.Context, commentID, userID uint32, req *clientdto.CommentListRequest) ([]clientdto.CommentItem, int, error) {
	comments, total, err := s.userRepo.ListReplies(ctx, commentID, req.GetOffset(), req.GetPageSize())
	if err != nil {
		s.log.Error("client user: list replies failed", zap.Uint32("parent_id", commentID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out, err := s.mapComments(ctx, comments, userID)
	if err != nil {
		s.log.Error("client user: map replies failed", zap.Uint32("parent_id", commentID), zap.Error(err))
		return nil, 0, err
	}
	return out, total, nil
}

func (s *userService) CreateComment(ctx context.Context, userID uint32, req *clientdto.CreateCommentRequest) (*clientdto.CommentItem, error) {
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
		s.log.Error("client user: get video for create comment failed", zap.Uint32("video_id", req.VideoID), zap.Error(err))
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

	if req.ParentID > 0 {
		parent, err := s.userRepo.GetComment(ctx, req.ParentID)
		if err != nil {
			s.log.Error("client user: get parent comment failed", zap.Uint32("parent_id", req.ParentID), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if parent == nil || parent.Status != constant.CommentStatusNormal {
			return nil, errcode.CommentNotFound
		}
		if parent.VideoID != req.VideoID {
			return nil, errcode.ParamError
		}
	}

	// Determine comment status based on review toggle
	status := constant.CommentStatusNormal
	if service.BoolVal(featureMap, constant.SettingFeatureCommentReview, true) {
		status = constant.CommentStatusHidden
	}

	c := &model.VideoComments{
		VideoID:  req.VideoID,
		UserID:   userID,
		ParentID: req.ParentID,
		Content:  content,
		Status:   status,
	}
	if err := s.userRepo.CreateComment(ctx, c); err != nil {
		s.log.Error("client user: create comment failed", zap.Uint32("video_id", req.VideoID), zap.Uint32("user_id", userID), zap.Error(err))
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
		MyVote:       0,
		ReplyCount:   0,
		Replies:      make([]*clientdto.CommentItem, 0),
		CreatedAt:    utils.FormatTimeStr(c.CreatedAt),
	}
	if u, _ := s.adminRepo.GetUserByID(ctx, userID); u != nil {
		item.Nickname = u.Nickname
		item.Avatar = u.Avatar
	}
	return item, nil
}

func (s *userService) VoteComment(ctx context.Context, userID, commentID uint32, req *clientdto.VoteCommentRequest) (*clientdto.VoteCommentResult, error) {
	c, err := s.userRepo.GetComment(ctx, commentID)
	if err != nil {
		s.log.Error("client user: get comment for vote failed", zap.Uint32("comment_id", commentID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil || c.Status != constant.CommentStatusNormal {
		return nil, errcode.CommentNotFound
	}

	var newDir int
	switch req.Action {
	case "like":
		newDir = 1
	case "dislike":
		newDir = -1
	case "cancel":
		newDir = 0
	}

	err = s.userRepo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := s.userRepo.WithTx(tx)
		oldVote, voteErr := txRepo.GetCommentVote(ctx, userID, commentID)
		if voteErr != nil {
			return voteErr
		}
		oldDir := 0
		if oldVote != nil {
			oldDir = int(oldVote.Direction)
		}
		if oldDir == newDir {
			return nil
		}

		likeDelta, dislikeDelta := 0, 0
		if oldDir == 1 {
			likeDelta--
		}
		if oldDir == -1 {
			dislikeDelta--
		}
		if newDir == 1 {
			likeDelta++
		}
		if newDir == -1 {
			dislikeDelta++
		}

		if newDir != 0 {
			if voteErr := txRepo.UpsertCommentVote(ctx, &model.UserCommentVotes{
				UserID:    userID,
				CommentID: commentID,
				Direction: utils.IntToInt8(newDir),
			}); voteErr != nil {
				return voteErr
			}
		} else {
			if voteErr := txRepo.DeleteCommentVote(ctx, userID, commentID); voteErr != nil {
				return voteErr
			}
		}
		if likeDelta != 0 || dislikeDelta != 0 {
			if voteErr := txRepo.UpdateCommentVoteCounts(ctx, commentID, likeDelta, dislikeDelta); voteErr != nil {
				return voteErr
			}
		}
		return nil
	})
	if err != nil {
		s.log.Error("client user: vote comment transaction failed", zap.Uint32("comment_id", commentID), zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	c, err = s.userRepo.GetComment(ctx, commentID)
	if err != nil {
		s.log.Error("client user: get comment after vote failed", zap.Uint32("comment_id", commentID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	myVote := int8(0)
	vote, err := s.userRepo.GetCommentVote(ctx, userID, commentID)
	if err != nil {
		s.log.Error("client user: get comment vote after vote failed", zap.Uint32("comment_id", commentID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if vote != nil {
		myVote = vote.Direction
	}
	return &clientdto.VoteCommentResult{
		LikeCount:    c.LikeCount,
		DislikeCount: c.DislikeCount,
		MyVote:       myVote,
	}, nil
}

// ===== Ratings (C6) =====

// RateVideo submits or updates a user's rating for a video.
func (s *userService) RateVideo(ctx context.Context, userID, videoID uint32, req *clientdto.RateVideoRequest) (*clientdto.RatingResult, error) {
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
	v, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("client user: get video for rate failed", zap.Uint32("video_id", videoID), zap.Error(err))
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
			UserID:  userID,
			VideoID: videoID,
			Score:   score,
		}
		if upsertErr := txUserRepo.UpsertRating(ctx, ur); upsertErr != nil {
			s.log.Error("client user: upsert rating failed", zap.Uint32("video_id", videoID), zap.Uint32("user_id", userID), zap.Error(upsertErr))
			return errcode.Wrap(errcode.DatabaseError, upsertErr)
		}

		avg, count, err = txUserRepo.GetRatingStats(ctx, videoID)
		if err != nil {
			s.log.Error("client user: get rating stats failed", zap.Uint32("video_id", videoID), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, err)
		}

		// Round to 1 decimal place
		rounded := math.Round(avg*10) / 10
		if updateErr := txVideoRepo.UpdateRatingStats(ctx, videoID, rounded, utils.IntToUint32(count)); updateErr != nil {
			s.log.Error("client user: update video rating stats failed", zap.Uint32("video_id", videoID), zap.Error(updateErr))
			return errcode.Wrap(errcode.DatabaseError, updateErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &clientdto.RatingResult{
		MyScore:     score,
		Rating:      math.Round(avg*10) / 10,
		RatingCount: utils.IntToUint32(count),
	}, nil
}

// GetRating returns the video's rating stats and the current user's score (0 if not logged in / not rated).
func (s *userService) GetRating(ctx context.Context, userID, videoID uint32) (*clientdto.RatingResult, error) {
	v, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		s.log.Error("client user: get video for get rating failed", zap.Uint32("video_id", videoID), zap.Error(err))
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
			s.log.Error("client user: get user rating failed", zap.Uint32("video_id", videoID), zap.Uint32("user_id", userID), zap.Error(err))
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
		ID:       u.ID,
		StrID:    u.StrID,
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

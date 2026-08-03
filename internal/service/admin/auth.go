package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/audit"
	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/crypto"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// LoginMeta carries request client info for login audit.
type LoginMeta struct {
	IP        string
	UserAgent string
}

// AuthService handles admin authentication.
type AuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest, meta *LoginMeta) (*dto.LoginResponse, error)
	Profile(ctx context.Context, adminID uint32) (*dto.Profile, error)
	// EnsureSuperAdmin loads admin+group and validates super_admin access for each request.
	EnsureSuperAdmin(ctx context.Context, adminID uint32) (*model.Admins, *model.UserGroups, error)
	UpdateProfile(ctx context.Context, adminID uint32, req *dto.UpdateProfileRequest) (*dto.Profile, error)
	ChangePassword(ctx context.Context, adminID uint32, req *dto.ChangePasswordRequest) error
}

type authService struct {
	adminRepo repository.AdminRepository
	jwtMgr    *auth.JWTManager
	accessTTL int
	audit     *audit.Recorder
	log       *zap.Logger
}

// NewAuthService creates an AuthService.
func NewAuthService(adminRepo repository.AdminRepository, jwtMgr *auth.JWTManager, cfg *config.Config, recorder *audit.Recorder, log *zap.Logger) AuthService {
	ttl := 7200
	if cfg != nil && cfg.JWT.AccessTokenTTL > 0 {
		ttl = cfg.JWT.AccessTokenTTL
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &authService{adminRepo: adminRepo, jwtMgr: jwtMgr, accessTTL: ttl, audit: recorder, log: log}
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest, meta *LoginMeta) (*dto.LoginResponse, error) {
	ip, ua := "", ""
	if meta != nil {
		ip, ua = meta.IP, meta.UserAgent
	}
	recordFail := func(userID uint32, username string) {
		if s.audit != nil {
			s.audit.AdminLogin(ctx, userID, username, ip, ua, false)
		}
	}
	recordOK := func(userID uint32, username string) {
		if s.audit != nil {
			s.audit.AdminLogin(ctx, userID, username, ip, ua, true)
		}
	}

	if s.jwtMgr == nil {
		return nil, errcode.WithMessage(errcode.ServiceUnavailable, "JWT 未配置")
	}
	username := strings.TrimSpace(req.Username)
	admin, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		s.log.Error("auth: get admin by username failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		recordFail(0, username)
		return nil, errcode.InvalidCredentials
	}
	if admin.Status != constant.StatusEnabled {
		recordFail(admin.ID, username)
		return nil, errcode.AdminDisabled
	}
	if err := crypto.CheckPassword(req.Password, admin.Password); err != nil {
		recordFail(admin.ID, username)
		return nil, errcode.InvalidCredentials
	}

	group, err := s.adminRepo.GetGroupByID(ctx, admin.GroupID)
	if err != nil {
		s.log.Error("auth: get group by id failed", zap.Uint32("admin_id", admin.ID), zap.Uint32("group_id", admin.GroupID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if group == nil || group.Name != constant.RoleSuperAdmin {
		recordFail(admin.ID, username)
		return nil, errcode.InsufficientPermission
	}

	token, err := s.jwtMgr.GenerateAccessToken(admin.ID)
	if err != nil {
		s.log.Error("auth: generate access token failed", zap.Uint32("admin_id", admin.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	now := time.Now()
	if err := s.adminRepo.UpdateLastLogin(ctx, admin.ID, now); err != nil {
		s.log.Error("auth: update last login failed", zap.Uint32("admin_id", admin.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	recordOK(admin.ID, username)

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.accessTTL,
		Admin:       toAdminProfile(admin, group.Name),
	}, nil
}

func (s *authService) Profile(ctx context.Context, adminID uint32) (*dto.Profile, error) {
	admin, group, err := s.EnsureSuperAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return toAdminProfile(admin, group.Name), nil
}

func (s *authService) UpdateProfile(ctx context.Context, adminID uint32, req *dto.UpdateProfileRequest) (*dto.Profile, error) {
	admin, group, err := s.EnsureSuperAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	nickname := strings.TrimSpace(req.Nickname)
	if len(nickname) > 50 {
		return nil, errcode.WithMessage(errcode.ParamError, "昵称长度不能超过50")
	}
	email := strings.TrimSpace(req.Email)
	if len(email) > 100 {
		return nil, errcode.WithMessage(errcode.ParamError, "邮箱长度不能超过100")
	}
	avatar := strings.TrimSpace(req.Avatar)
	if len(avatar) > 500 {
		return nil, errcode.WithMessage(errcode.ParamError, "头像URL长度不能超过500")
	}
	if err := s.adminRepo.UpdateProfile(ctx, adminID, nickname, email, avatar); err != nil {
		s.log.Error("auth: update admin profile failed", zap.Uint32("admin_id", adminID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	admin.Nickname = nickname
	admin.Email = email
	admin.Avatar = avatar
	return toAdminProfile(admin, group.Name), nil
}

func (s *authService) ChangePassword(ctx context.Context, adminID uint32, req *dto.ChangePasswordRequest) error {
	admin, _, err := s.EnsureSuperAdmin(ctx, adminID)
	if err != nil {
		return err
	}
	if err := crypto.CheckPassword(req.OldPassword, admin.Password); err != nil {
		return errcode.InvalidCredentials
	}
	if req.OldPassword == req.NewPassword {
		return errcode.WithMessage(errcode.ParamError, "新密码不能与旧密码相同")
	}
	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("auth: hash new password failed", zap.Uint32("admin_id", adminID), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}
	if err := s.adminRepo.UpdatePassword(ctx, adminID, hash); err != nil {
		s.log.Error("auth: update admin password failed", zap.Uint32("admin_id", adminID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *authService) EnsureSuperAdmin(ctx context.Context, adminID uint32) (*model.Admins, *model.UserGroups, error) {
	if adminID == 0 {
		return nil, nil, errcode.AuthFailed
	}
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		s.log.Error("auth: get admin by id failed", zap.Uint32("admin_id", adminID), zap.Error(err))
		return nil, nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		return nil, nil, errcode.AuthFailed
	}
	if admin.Status != constant.StatusEnabled {
		return nil, nil, errcode.AdminDisabled
	}
	group, err := s.adminRepo.GetGroupByID(ctx, admin.GroupID)
	if err != nil {
		s.log.Error("auth: get group by id for super admin check failed", zap.Uint32("admin_id", adminID), zap.Uint32("group_id", admin.GroupID), zap.Error(err))
		return nil, nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if group == nil || group.Name != constant.RoleSuperAdmin {
		return nil, nil, errcode.InsufficientPermission
	}
	return admin, group, nil
}

func toAdminProfile(admin *model.Admins, role string) *dto.Profile {
	return &dto.Profile{
		ID:       admin.ID,
		Username: admin.Username,
		Nickname: admin.Nickname,
		Email:    admin.Email,
		Avatar:   admin.Avatar,
		Role:     role,
		Status:   admin.Status,
	}
}

// CreateAdmin seeds an admin with super_admin group (CLI use).
func CreateAdmin(ctx context.Context, adminRepo repository.AdminRepository, username, password, email string) (*model.Admins, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 50 {
		return nil, fmt.Errorf("username length must be 3-50")
	}
	if len(password) < 6 || len(password) > 72 {
		return nil, fmt.Errorf("password length must be 6-72")
	}

	exists, err := adminRepo.ExistsUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	group, err := adminRepo.GetGroupByName(ctx, constant.RoleSuperAdmin)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("super_admin group not found; run migrations first")
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}
	admin := &model.Admins{
		Username: username,
		Password: hash,
		Email:    strings.TrimSpace(email),
		GroupID:  group.ID,
		Status:   constant.StatusEnabled,
	}
	if err := adminRepo.Create(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

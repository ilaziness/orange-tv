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
)

// LoginMeta carries request client info for login audit.
type LoginMeta struct {
	IP        string
	UserAgent string
}

// AuthService handles admin authentication.
type AuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest, meta *LoginMeta) (*dto.LoginResponse, error)
	Profile(ctx context.Context, adminID int64) (*dto.Profile, error)
	// EnsureSuperAdmin loads admin+group and validates super_admin access for each request.
	EnsureSuperAdmin(ctx context.Context, adminID int64) (*model.Admins, *model.UserGroups, error)
}

type authService struct {
	adminRepo repository.AdminRepository
	jwtMgr    *auth.JWTManager
	accessTTL int
	audit     *audit.Recorder
}

// NewAuthService creates an AuthService.
func NewAuthService(adminRepo repository.AdminRepository, jwtMgr *auth.JWTManager, cfg *config.Config, recorder *audit.Recorder) AuthService {
	ttl := 7200
	if cfg != nil && cfg.JWT.AccessTokenTTL > 0 {
		ttl = cfg.JWT.AccessTokenTTL
	}
	return &authService{adminRepo: adminRepo, jwtMgr: jwtMgr, accessTTL: ttl, audit: recorder}
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest, meta *LoginMeta) (*dto.LoginResponse, error) {
	ip, ua := "", ""
	if meta != nil {
		ip, ua = meta.IP, meta.UserAgent
	}
	recordFail := func(userID int64, username string) {
		if s.audit != nil {
			s.audit.Login(ctx, constant.LoginUserTypeAdmin, userID, username, ip, ua, false)
		}
	}
	recordOK := func(userID int64, username string) {
		if s.audit != nil {
			s.audit.Login(ctx, constant.LoginUserTypeAdmin, userID, username, ip, ua, true)
		}
	}

	if s.jwtMgr == nil {
		return nil, errcode.WithMessage(errcode.ServiceUnavailable, "JWT 未配置")
	}
	username := strings.TrimSpace(req.Username)
	admin, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		recordFail(0, username)
		return nil, errcode.InvalidCredentials
	}
	if admin.Status != constant.StatusEnabled {
		recordFail(int64(admin.ID), username)
		return nil, errcode.AdminDisabled
	}
	if err := crypto.CheckPassword(req.Password, admin.Password); err != nil {
		recordFail(int64(admin.ID), username)
		return nil, errcode.InvalidCredentials
	}

	group, err := s.adminRepo.GetGroupByID(ctx, int64(admin.GroupID))
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if group == nil || group.Name != constant.RoleSuperAdmin {
		recordFail(int64(admin.ID), username)
		return nil, errcode.InsufficientPermission
	}

	token, err := s.jwtMgr.GenerateAccessToken(int64(admin.ID))
	if err != nil {
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	now := time.Now()
	if err := s.adminRepo.UpdateLastLogin(ctx, int64(admin.ID), now); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	recordOK(int64(admin.ID), username)

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.accessTTL,
		Admin:       toAdminProfile(admin, group.Name),
	}, nil
}

func (s *authService) Profile(ctx context.Context, adminID int64) (*dto.Profile, error) {
	admin, group, err := s.EnsureSuperAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return toAdminProfile(admin, group.Name), nil
}

func (s *authService) EnsureSuperAdmin(ctx context.Context, adminID int64) (*model.Admins, *model.UserGroups, error) {
	if adminID <= 0 {
		return nil, nil, errcode.AuthFailed
	}
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		return nil, nil, errcode.AuthFailed
	}
	if admin.Status != constant.StatusEnabled {
		return nil, nil, errcode.AdminDisabled
	}
	group, err := s.adminRepo.GetGroupByID(ctx, int64(admin.GroupID))
	if err != nil {
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

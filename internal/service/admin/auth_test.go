package admin

import (
	"context"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/crypto"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/require"
)

type fakeAdminRepo struct {
	admins map[uint64]*model.Admins
	groups map[uint64]*model.UserGroups
}

func (f *fakeAdminRepo) GetByUsername(ctx context.Context, username string) (*model.Admins, error) {
	for _, a := range f.admins {
		if a.Username == username && a.DeletedAt == nil {
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeAdminRepo) GetByID(ctx context.Context, id int64) (*model.Admins, error) {
	a, ok := f.admins[uint64(id)]
	if !ok || a.DeletedAt != nil {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (f *fakeAdminRepo) GetGroupByID(ctx context.Context, id int64) (*model.UserGroups, error) {
	g, ok := f.groups[uint64(id)]
	if !ok || g.DeletedAt != nil {
		return nil, nil
	}
	cp := *g
	return &cp, nil
}

func (f *fakeAdminRepo) GetGroupByName(ctx context.Context, name string) (*model.UserGroups, error) {
	for _, g := range f.groups {
		if g.Name == name && g.DeletedAt == nil {
			cp := *g
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeAdminRepo) Create(ctx context.Context, admin *model.Admins) error {
	admin.ID = uint64(len(f.admins) + 1)
	cp := *admin
	f.admins[admin.ID] = &cp
	return nil
}

func (f *fakeAdminRepo) UpdateLastLogin(ctx context.Context, id int64, at time.Time) error {
	if a, ok := f.admins[uint64(id)]; ok {
		a.LastLoginAt = &at
	}
	return nil
}

func (f *fakeAdminRepo) ExistsUsername(ctx context.Context, username string) (bool, error) {
	for _, a := range f.admins {
		if a.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeAdminRepo) UpdateProfile(ctx context.Context, adminID int64, nickname, email, avatar string) error {
	if a, ok := f.admins[uint64(adminID)]; ok {
		a.Nickname = nickname
		a.Email = email
		a.Avatar = avatar
	}
	return nil
}

func (f *fakeAdminRepo) UpdatePassword(ctx context.Context, adminID int64, hashedPassword string) error {
	if a, ok := f.admins[uint64(adminID)]; ok {
		a.Password = hashedPassword
	}
	return nil
}

// Stub implementations for new AdminRepository methods (not used in auth tests)
func (f *fakeAdminRepo) ListAdmins(ctx context.Context, fl repository.AdminListFilter) ([]model.Admins, int, error) {
	return nil, 0, nil
}
func (f *fakeAdminRepo) UpdateAdmin(ctx context.Context, admin *model.Admins) error { return nil }
func (f *fakeAdminRepo) SoftDeleteAdmin(ctx context.Context, id int64) error        { return nil }
func (f *fakeAdminRepo) ExistsUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error) {
	return false, nil
}
func (f *fakeAdminRepo) ListGroups(ctx context.Context, fl repository.UserGroupListFilter) ([]model.UserGroups, int, error) {
	return nil, 0, nil
}
func (f *fakeAdminRepo) CreateGroup(ctx context.Context, g *model.UserGroups) error { return nil }
func (f *fakeAdminRepo) UpdateGroup(ctx context.Context, g *model.UserGroups) error { return nil }
func (f *fakeAdminRepo) SoftDeleteGroup(ctx context.Context, id int64) error        { return nil }
func (f *fakeAdminRepo) ExistsGroupNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error) {
	return false, nil
}
func (f *fakeAdminRepo) ListUsers(ctx context.Context, fl repository.UserListFilter) ([]model.Users, int, error) {
	return nil, 0, nil
}
func (f *fakeAdminRepo) GetUserByID(ctx context.Context, id int64) (*model.Users, error) {
	return nil, nil
}
func (f *fakeAdminRepo) GetUserByUsername(ctx context.Context, username string) (*model.Users, error) {
	return nil, nil
}
func (f *fakeAdminRepo) CreateUser(ctx context.Context, u *model.Users) error { return nil }
func (f *fakeAdminRepo) UpdateUser(ctx context.Context, u *model.Users) error { return nil }
func (f *fakeAdminRepo) SoftDeleteUser(ctx context.Context, id int64) error   { return nil }
func (f *fakeAdminRepo) ExistsUserUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (f *fakeAdminRepo) ExistsUserUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error) {
	return false, nil
}

func TestAuthService_LoginSuccess(t *testing.T) {
	hash, err := crypto.HashPassword("secret12")
	require.NoError(t, err)
	repo := &fakeAdminRepo{
		admins: map[uint64]*model.Admins{
			1: {ID: 1, Username: "admin", Password: hash, GroupID: 1, Status: constant.StatusEnabled},
		},
		groups: map[uint64]*model.UserGroups{
			1: {ID: 1, Name: constant.RoleSuperAdmin},
		},
	}
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	svc := NewAuthService(repo, jwtMgr, &config.Config{JWT: config.JWTConfig{AccessTokenTTL: 3600}}, nil, nil)

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "admin", Password: "secret12"}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, constant.RoleSuperAdmin, resp.Admin.Role)
}

func TestAuthService_LoginRejectsWrongPassword(t *testing.T) {
	hash, err := crypto.HashPassword("secret12")
	require.NoError(t, err)
	repo := &fakeAdminRepo{
		admins: map[uint64]*model.Admins{
			1: {ID: 1, Username: "admin", Password: hash, GroupID: 1, Status: constant.StatusEnabled},
		},
		groups: map[uint64]*model.UserGroups{
			1: {ID: 1, Name: constant.RoleSuperAdmin},
		},
	}
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	svc := NewAuthService(repo, jwtMgr, &config.Config{JWT: config.JWTConfig{AccessTokenTTL: 3600}}, nil, nil)

	_, err = svc.Login(context.Background(), &dto.LoginRequest{Username: "admin", Password: "bad-pass"}, nil)
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.InvalidCredentials.Code, code.Code)
}

func TestAuthService_EnsureSuperAdminRejectsDisabled(t *testing.T) {
	repo := &fakeAdminRepo{
		admins: map[uint64]*model.Admins{
			1: {ID: 1, Username: "admin", GroupID: 1, Status: constant.StatusDisabled},
		},
		groups: map[uint64]*model.UserGroups{
			1: {ID: 1, Name: constant.RoleSuperAdmin},
		},
	}
	svc := NewAuthService(repo, nil, nil, nil, nil)
	_, _, err := svc.EnsureSuperAdmin(context.Background(), 1)
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.AdminDisabled.Code, code.Code)
}

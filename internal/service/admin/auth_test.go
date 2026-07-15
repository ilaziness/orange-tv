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
	"github.com/stretchr/testify/require"
)

type fakeAdminRepo struct {
	admins map[int64]*model.Admins
	groups map[int64]*model.UserGroups
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
	a, ok := f.admins[id]
	if !ok || a.DeletedAt != nil {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (f *fakeAdminRepo) GetGroupByID(ctx context.Context, id int64) (*model.UserGroups, error) {
	g, ok := f.groups[id]
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
	admin.ID = int64(len(f.admins) + 1)
	cp := *admin
	f.admins[admin.ID] = &cp
	return nil
}

func (f *fakeAdminRepo) UpdateLastLogin(ctx context.Context, id int64, at time.Time) error {
	if a, ok := f.admins[id]; ok {
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

func TestAuthService_LoginSuccess(t *testing.T) {
	hash, err := crypto.HashPassword("secret12")
	require.NoError(t, err)
	repo := &fakeAdminRepo{
		admins: map[int64]*model.Admins{
			1: {ID: 1, Username: "admin", Password: hash, GroupID: 1, Status: constant.StatusEnabled},
		},
		groups: map[int64]*model.UserGroups{
			1: {ID: 1, Name: constant.RoleSuperAdmin},
		},
	}
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	svc := NewAuthService(repo, jwtMgr, &config.Config{JWT: config.JWTConfig{AccessTokenTTL: 3600}})

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "admin", Password: "secret12"})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, constant.RoleSuperAdmin, resp.Admin.Role)
}

func TestAuthService_LoginRejectsWrongPassword(t *testing.T) {
	hash, err := crypto.HashPassword("secret12")
	require.NoError(t, err)
	repo := &fakeAdminRepo{
		admins: map[int64]*model.Admins{
			1: {ID: 1, Username: "admin", Password: hash, GroupID: 1, Status: constant.StatusEnabled},
		},
		groups: map[int64]*model.UserGroups{
			1: {ID: 1, Name: constant.RoleSuperAdmin},
		},
	}
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	svc := NewAuthService(repo, jwtMgr, &config.Config{JWT: config.JWTConfig{AccessTokenTTL: 3600}})

	_, err = svc.Login(context.Background(), &dto.LoginRequest{Username: "admin", Password: "bad-pass"})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.InvalidCredentials.Code, code.Code)
}

func TestAuthService_EnsureSuperAdminRejectsDisabled(t *testing.T) {
	repo := &fakeAdminRepo{
		admins: map[int64]*model.Admins{
			1: {ID: 1, Username: "admin", GroupID: 1, Status: constant.StatusDisabled},
		},
		groups: map[int64]*model.UserGroups{
			1: {ID: 1, Name: constant.RoleSuperAdmin},
		},
	}
	svc := NewAuthService(repo, nil, nil)
	_, _, err := svc.EnsureSuperAdmin(context.Background(), 1)
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.AdminDisabled.Code, code.Code)
}

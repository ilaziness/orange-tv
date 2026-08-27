package client

import (
	"context"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeClientUserAdminRepo struct {
	users map[uint32]*model.Users
}

func (f *fakeClientUserAdminRepo) GetByUsername(context.Context, string) (*model.Admins, error) {
	return nil, nil
}
func (f *fakeClientUserAdminRepo) GetByID(context.Context, uint32) (*model.Admins, error) {
	return nil, nil
}
func (f *fakeClientUserAdminRepo) GetGroupByID(context.Context, uint32) (*model.UserGroups, error) {
	return nil, nil
}
func (f *fakeClientUserAdminRepo) GetGroupByName(context.Context, string) (*model.UserGroups, error) {
	return nil, nil
}
func (f *fakeClientUserAdminRepo) Create(context.Context, *model.Admins) error { return nil }
func (f *fakeClientUserAdminRepo) UpdateLastLogin(context.Context, uint32, time.Time) error {
	return nil
}
func (f *fakeClientUserAdminRepo) ExistsUsername(context.Context, string) (bool, error) {
	return false, nil
}
func (f *fakeClientUserAdminRepo) UpdateProfile(context.Context, uint32, string, string, string) error {
	return nil
}
func (f *fakeClientUserAdminRepo) UpdatePassword(context.Context, uint32, string) error { return nil }
func (f *fakeClientUserAdminRepo) ListAdmins(context.Context, repository.AdminListFilter) ([]model.Admins, int, error) {
	return nil, 0, nil
}
func (f *fakeClientUserAdminRepo) UpdateAdmin(context.Context, *model.Admins) error { return nil }
func (f *fakeClientUserAdminRepo) SoftDeleteAdmin(context.Context, uint32) error    { return nil }
func (f *fakeClientUserAdminRepo) ExistsUsernameExcludeID(context.Context, string, uint32) (bool, error) {
	return false, nil
}
func (f *fakeClientUserAdminRepo) ListGroups(context.Context, repository.UserGroupListFilter) ([]model.UserGroups, int, error) {
	return nil, 0, nil
}
func (f *fakeClientUserAdminRepo) CreateGroup(context.Context, *model.UserGroups) error { return nil }
func (f *fakeClientUserAdminRepo) UpdateGroup(context.Context, *model.UserGroups) error { return nil }
func (f *fakeClientUserAdminRepo) SoftDeleteGroup(context.Context, uint32) error        { return nil }
func (f *fakeClientUserAdminRepo) ExistsGroupNameExcludeID(context.Context, string, uint32) (bool, error) {
	return false, nil
}
func (f *fakeClientUserAdminRepo) ListUsers(context.Context, repository.UserListFilter) ([]model.Users, int, error) {
	return nil, 0, nil
}
func (f *fakeClientUserAdminRepo) GetUserByID(_ context.Context, id uint32) (*model.Users, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}
func (f *fakeClientUserAdminRepo) GetUserByEmail(context.Context, string) (*model.Users, error) {
	return nil, nil
}
func (f *fakeClientUserAdminRepo) GetUserByStrID(context.Context, string) (*model.Users, error) {
	return nil, nil
}
func (f *fakeClientUserAdminRepo) CreateUser(context.Context, *model.Users) error { return nil }
func (f *fakeClientUserAdminRepo) UpdateUser(context.Context, *model.Users) error { return nil }
func (f *fakeClientUserAdminRepo) SoftDeleteUser(context.Context, uint32) error   { return nil }
func (f *fakeClientUserAdminRepo) ExistsUserEmail(context.Context, string) (bool, error) {
	return false, nil
}
func (f *fakeClientUserAdminRepo) ExistsUserEmailExcludeID(context.Context, string, uint32) (bool, error) {
	return false, nil
}
func (f *fakeClientUserAdminRepo) ExistsUserStrID(context.Context, string) (bool, error) {
	return false, nil
}

func newRefreshTestUserService(jwtMgr *auth.JWTManager, repo repository.AdminRepository) UserService {
	return NewUserService(repo, nil, nil, nil, jwtMgr, 3600, 86400, nil, nil, nil, zap.NewNop())
}

func TestUserService_RefreshTokenSuccess(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	repo := &fakeClientUserAdminRepo{
		users: map[uint32]*model.Users{
			42: {ID: 42, Email: "user@example.com", Status: constant.StatusEnabled},
		},
	}
	svc := newRefreshTestUserService(jwtMgr, repo)

	refreshToken, err := jwtMgr.GenerateRefreshTokenFor(42, auth.SubjectUser)
	require.NoError(t, err)

	resp, err := svc.RefreshToken(context.Background(), &clientdto.RefreshTokenRequest{RefreshToken: refreshToken})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.NotEmpty(t, resp.RefreshToken)
	require.Equal(t, "Bearer", resp.TokenType)
	require.Equal(t, 3600, resp.ExpiresIn)
	require.Equal(t, 86400, resp.RefreshExpiresIn)
}

func TestUserService_RefreshTokenRejectsAccessToken(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	repo := &fakeClientUserAdminRepo{
		users: map[uint32]*model.Users{
			42: {ID: 42, Email: "user@example.com", Status: constant.StatusEnabled},
		},
	}
	svc := newRefreshTestUserService(jwtMgr, repo)

	accessToken, err := jwtMgr.GenerateAccessTokenFor(42, auth.SubjectUser)
	require.NoError(t, err)

	_, err = svc.RefreshToken(context.Background(), &clientdto.RefreshTokenRequest{RefreshToken: accessToken})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.InvalidToken.Code, code.Code)
}

func TestUserService_RefreshTokenRejectsDisabledUser(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 86400)
	repo := &fakeClientUserAdminRepo{
		users: map[uint32]*model.Users{
			42: {ID: 42, Email: "user@example.com", Status: constant.StatusDisabled},
		},
	}
	svc := newRefreshTestUserService(jwtMgr, repo)

	refreshToken, err := jwtMgr.GenerateRefreshTokenFor(42, auth.SubjectUser)
	require.NoError(t, err)

	_, err = svc.RefreshToken(context.Background(), &clientdto.RefreshTokenRequest{RefreshToken: refreshToken})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.UserDisabled.Code, code.Code)
}

func TestUserService_RefreshTokenRejectsExpiredToken(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret-key", 3600, 1)
	repo := &fakeClientUserAdminRepo{
		users: map[uint32]*model.Users{
			42: {ID: 42, Email: "user@example.com", Status: constant.StatusEnabled},
		},
	}
	svc := newRefreshTestUserService(jwtMgr, repo)

	refreshToken, err := jwtMgr.GenerateRefreshTokenFor(42, auth.SubjectUser)
	require.NoError(t, err)
	time.Sleep(1100 * time.Millisecond)

	_, err = svc.RefreshToken(context.Background(), &clientdto.RefreshTokenRequest{RefreshToken: refreshToken})
	require.Error(t, err)
	code, ok := errcode.As(err)
	require.True(t, ok)
	require.Equal(t, errcode.TokenExpired.Code, code.Code)
}

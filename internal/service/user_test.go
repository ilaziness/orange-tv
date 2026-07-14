package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/ilaziness/orange-tv/internal/crypto"
	"github.com/ilaziness/orange-tv/internal/dto"
	"github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uptrace/bun"
)

// MockUserRepository 是 repository.UserRepository 的 mock 实现
type MockUserRepository struct {
	mock.Mock
}

// 确保 MockUserRepository 实现了 repository.UserRepository 接口
var _ repository.UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockUserRepository) WithTx(db bun.IDB) repository.UserRepository {
	args := m.Called(db)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.UserRepository)
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.CreateUserRequest
		mockFn  func(*MockUserRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "创建用户成功",
			req: &dto.CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
				Phone:    "1234567890",
			},
			mockFn: func(m *MockUserRepository) {
				// 检查邮箱不存在
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, sql.ErrNoRows)
				// 创建成功
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "查询邮箱失败",
			req: &dto.CreateUserRequest{
				Email:    "db-error@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			mockFn: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "db-error@example.com").Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "failed to check user email",
		},
		{
			name: "邮箱已存在",
			req: &dto.CreateUserRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			mockFn: func(m *MockUserRepository) {
				existingUser := &model.User{ID: 1, Email: "existing@example.com"}
				m.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			wantErr: true,
			errMsg:  "用户已存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockFn(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.Create(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Email, resp.Email)
				assert.Equal(t, tt.req.Name, resp.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		mockFn   func(*MockUserRepository)
		wantUser *model.User
		wantErr  bool
	}{
		{
			name: "获取用户成功",
			id:   1,
			mockFn: func(m *MockUserRepository) {
				user := &model.User{
					ID:     1,
					Email:  "test@example.com",
					Name:   "Test User",
					Status: 1,
				}
				m.On("GetByID", mock.Anything, int64(1)).Return(user, nil)
			},
			wantUser: &model.User{
				ID:     1,
				Email:  "test@example.com",
				Name:   "Test User",
				Status: 1,
			},
			wantErr: false,
		},
		{
			name: "用户不存在",
			id:   999,
			mockFn: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("not found"))
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockFn(mockRepo)

			service := NewUserService(mockRepo)
			user, err := service.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser.ID, user.ID)
				assert.Equal(t, tt.wantUser.Email, user.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ValidateCredentials(t *testing.T) {
	// 首先创建一个带哈希密码的用户
	hashedPassword, _ := crypto.HashPassword("correctpassword")

	tests := []struct {
		name     string
		email    string
		password string
		mockFn   func(*MockUserRepository)
		wantUser *model.User
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "正确凭据",
			email:    "test@example.com",
			password: "correctpassword",
			mockFn: func(m *MockUserRepository) {
				user := &model.User{
					ID:       1,
					Email:    "test@example.com",
					Password: hashedPassword,
					Name:     "Test User",
					Status:   1,
				}
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			wantUser: &model.User{
				ID:     1,
				Email:  "test@example.com",
				Name:   "Test User",
				Status: 1,
			},
			wantErr: false,
		},
		{
			name:     "错误密码",
			email:    "test@example.com",
			password: "wrongpassword",
			mockFn: func(m *MockUserRepository) {
				user := &model.User{
					ID:       1,
					Email:    "test@example.com",
					Password: hashedPassword,
					Status:   1,
				}
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			wantUser: nil,
			wantErr:  true,
			errMsg:   "认证失败",
		},
		{
			name:     "用户不存在",
			email:    "nonexistent@example.com",
			password: "password123",
			mockFn: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("not found"))
			},
			wantUser: nil,
			wantErr:  true,
			errMsg:   "认证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockFn(mockRepo)

			service := NewUserService(mockRepo)
			user, err := service.ValidateCredentials(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.wantUser.ID, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	req := &dto.UpdateUserRequest{
		ID:   1,
		Name: "Updated Name",
	}

	existingUser := &model.User{ID: 1, Email: "test@example.com", Name: "Original", Status: 1}
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(existingUser, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	resp, err := service.Update(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated Name", resp.Name)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	existingUser := &model.User{ID: 1, Email: "test@example.com", Name: "Test", Status: 1}
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(existingUser, nil)
	mockRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

	err := service.Delete(context.Background(), 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("not found"))

	err := service.Delete(context.Background(), 999)
	assert.Error(t, err)
	codeErr, ok := errcode.As(err)
	assert.True(t, ok)
	assert.Equal(t, errcode.UserNotFound.Code, codeErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ListWithCount(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		mockFn    func(*MockUserRepository)
		wantUsers int
		wantTotal int
		wantErr   bool
	}{
		{
			name:   "正常分页",
			limit:  10,
			offset: 0,
			mockFn: func(m *MockUserRepository) {
				users := []*model.User{
					{ID: 1, Email: "user1@example.com"},
					{ID: 2, Email: "user2@example.com"},
				}
				m.On("List", mock.Anything, 10, 0).Return(users, nil)
				m.On("Count", mock.Anything).Return(2, nil)
			},
			wantUsers: 2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:   "空列表",
			limit:  10,
			offset: 100,
			mockFn: func(m *MockUserRepository) {
				m.On("List", mock.Anything, 10, 100).Return([]*model.User{}, nil)
				m.On("Count", mock.Anything).Return(0, nil)
			},
			wantUsers: 0,
			wantTotal: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockFn(mockRepo)

			service := NewUserService(mockRepo)
			users, total, err := service.ListWithCount(context.Background(), tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.wantUsers)
				assert.Equal(t, tt.wantTotal, total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ListWithCount_ListError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("List", mock.Anything, 10, 0).Return(nil, fmt.Errorf("database error"))

	users, total, err := service.ListWithCount(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, 0, total)
}

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService 是 service.UserService 的 mock 实现
type MockUserService struct {
	mock.Mock
}

var _ service.UserService = (*MockUserService)(nil)

func (m *MockUserService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) ValidateCredentials(ctx context.Context, email, password string) (*model.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserService) ListWithCount(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}

func setupTestHandler() (*UserHandler, *MockUserService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := gin.New()
	return handler, mockService, router
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockFn     func(*MockUserService)
		wantStatus int
		wantErr    bool
	}{
		{
			name: "获取用户成功",
			id:   "1",
			mockFn: func(m *MockUserService) {
				user := &model.User{
					ID:        1,
					Email:     "test@example.com",
					Name:      "Test User",
					Status:    1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("GetByID", mock.Anything, int64(1)).Return(user, nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "无效ID",
			id:         "invalid",
			mockFn:     func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "用户不存在",
			id:   "999",
			mockFn: func(m *MockUserService) {
				m.On("GetByID", mock.Anything, int64(999)).Return(nil, errcode.Wrap(errcode.UserNotFound, errors.New("not found")))
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, router := setupTestHandler()
			tt.mockFn(mockService)

			router.GET("/users/:id", handler.GetUser)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/users/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		mockFn     func(*MockUserService)
		wantStatus int
		wantErr    bool
	}{
		{
			name: "创建用户成功",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
				"phone":    "1234567890",
			},
			mockFn: func(m *MockUserService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).Return(&dto.UserResponse{ID: 1, Email: "test@example.com", Name: "Test User", Phone: "1234567890", Status: 1}, nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "缺少邮箱",
			body: map[string]any{
				"password": "password123",
				"name":     "Test User",
			},
			mockFn:     func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "无效邮箱格式",
			body: map[string]any{
				"email":    "invalid-email",
				"password": "password123",
				"name":     "Test User",
			},
			mockFn:     func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "密码太短",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "123",
				"name":     "Test User",
			},
			mockFn:     func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "邮箱已存在",
			body: map[string]any{
				"email":    "existing@example.com",
				"password": "password123",
				"name":     "Test User",
			},
			mockFn: func(m *MockUserService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).Return(nil, errcode.UserAlreadyExists)
			},
			wantStatus: http.StatusConflict,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, router := setupTestHandler()
			tt.mockFn(mockService)

			router.POST("/users", handler.CreateUser)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       map[string]any
		mockFn     func(*MockUserService)
		wantStatus int
		wantErr    bool
	}{
		{
			name: "更新用户成功",
			id:   "1",
			body: map[string]any{
				"name": "Updated Name",
			},
			mockFn: func(m *MockUserService) {
				m.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateUserRequest")).Return(&dto.UserResponse{ID: 1, Email: "test@example.com", Name: "Updated Name", Status: 1}, nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "用户不存在",
			id:   "999",
			body: map[string]any{
				"name": "Updated Name",
			},
			mockFn: func(m *MockUserService) {
				m.On("Update", mock.Anything, int64(999), mock.AnythingOfType("*dto.UpdateUserRequest")).Return(nil, errcode.Wrap(errcode.UserNotFound, errors.New("not found")))
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, router := setupTestHandler()
			tt.mockFn(mockService)

			router.PUT("/users/:id", handler.UpdateUser)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/users/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockFn     func(*MockUserService)
		wantStatus int
		wantErr    bool
	}{
		{
			name: "删除用户成功",
			id:   "1",
			mockFn: func(m *MockUserService) {
				m.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "无效ID",
			id:         "invalid",
			mockFn:     func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "用户不存在",
			id:   "999",
			mockFn: func(m *MockUserService) {
				m.On("Delete", mock.Anything, int64(999)).Return(errcode.Wrap(errcode.UserNotFound, errors.New("not found")))
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, router := setupTestHandler()
			tt.mockFn(mockService)

			router.DELETE("/users/:id", handler.DeleteUser)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/users/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		mockFn     func(*MockUserService)
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:  "获取用户列表成功",
			query: "?page=1&limit=10",
			mockFn: func(m *MockUserService) {
				users := []*model.User{
					{ID: 1, Email: "user1@example.com", Name: "User 1"},
					{ID: 2, Email: "user2@example.com", Name: "User 2"},
				}
				m.On("ListWithCount", mock.Anything, 10, 0).Return(users, 2, nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result struct {
					Code int `json:"code"`
					Data struct {
						List       []dto.UserResponse `json:"list"`
						Total      int64              `json:"total"`
						Page       int                `json:"page"`
						PageSize   int                `json:"page_size"`
						TotalPages int                `json:"total_pages"`
					} `json:"data"`
				}
				err := json.Unmarshal(body, &result)
				assert.NoError(t, err)
				assert.Len(t, result.Data.List, 2)
				assert.Equal(t, int64(2), result.Data.Total)
				assert.Equal(t, 1, result.Data.Page)
				assert.Equal(t, 10, result.Data.PageSize)
				assert.Equal(t, 1, result.Data.TotalPages)
			},
		},
		{
			name:  "第二页",
			query: "?page=2&limit=5",
			mockFn: func(m *MockUserService) {
				m.On("ListWithCount", mock.Anything, 5, 5).Return([]*model.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result struct {
					Code int `json:"code"`
					Data struct {
						List     []dto.UserResponse `json:"list"`
						Total    int64              `json:"total"`
						Page     int                `json:"page"`
						PageSize int                `json:"page_size"`
					} `json:"data"`
				}
				err := json.Unmarshal(body, &result)
				assert.NoError(t, err)
				assert.Len(t, result.Data.List, 0)
				assert.Equal(t, 2, result.Data.Page)
			},
		},
		{
			name:  "默认分页参数",
			query: "",
			mockFn: func(m *MockUserService) {
				m.On("ListWithCount", mock.Anything, 10, 0).Return([]*model.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			checkBody:  func(t *testing.T, body []byte) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, router := setupTestHandler()
			tt.mockFn(mockService)

			router.GET("/users", handler.ListUsers)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/users"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestParsePaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{
			name:      "默认参数",
			query:     "",
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "自定义参数",
			query:     "?page=2&limit=20",
			wantPage:  2,
			wantLimit: 20,
		},
		{
			name:      "无效参数使用默认",
			query:     "?page=0&limit=0",
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "限制最大 limit",
			query:     "?page=1&limit=200",
			wantPage:  1,
			wantLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			var gotPage, gotLimit int

			router.GET("/test", func(c *gin.Context) {
				var req dto.ListUsersRequest
				_ = c.ShouldBindQuery(&req)
				gotPage = req.GetPage()
				gotLimit = req.GetLimit()
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantPage, gotPage)
			assert.Equal(t, tt.wantLimit, gotLimit)
		})
	}
}

func TestCalcTotalPages(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		limit     int
		wantPages int
	}{
		{
			name:      "整除",
			total:     100,
			limit:     10,
			wantPages: 10,
		},
		{
			name:      "有余数",
			total:     95,
			limit:     10,
			wantPages: 10,
		},
		{
			name:      "总数为0",
			total:     0,
			limit:     10,
			wantPages: 0,
		},
		{
			name:      "limit为0",
			total:     10,
			limit:     0,
			wantPages: 1,
		},
		{
			name:      "总数小于limit",
			total:     5,
			limit:     10,
			wantPages: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagReq := dto.PaginationRequest{Limit: tt.limit}
			got := pagReq.GetTotalPages(tt.total)
			assert.Equal(t, tt.wantPages, got)
		})
	}
}

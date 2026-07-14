package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database/testutil"
	"github.com/ilaziness/orange-tv/internal/dto"
	httphandler "github.com/ilaziness/orange-tv/internal/handler/http"
	"github.com/ilaziness/orange-tv/internal/health"
	httpmiddleware "github.com/ilaziness/orange-tv/internal/middleware/http"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/router"
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
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}

func (m *MockUserService) ValidateCredentials(ctx context.Context, email, password string) (*model.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, id int64, newPassword string) error {
	args := m.Called(ctx, id, newPassword)
	return args.Error(0)
}

func newIntegrationHealthHandler(t *testing.T, cfg *config.Config) *httphandler.HealthHandler {
	t.Helper()
	healthHandler := httphandler.NewHealthHandler(cfg)
	healthHandler.AddChecker(health.NewDatabaseChecker(testutil.OpenBunDB(t)))
	return healthHandler
}

func setupTestRouter(t *testing.T) (*gin.Engine, *MockUserService) {
	gin.SetMode(gin.TestMode)

	// 创建 mock service
	mockUserService := new(MockUserService)

	// 创建 config
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "orange-tv",
			Version: "1.0.0",
			Env:     constant.EnvTest,
		},
	}

	healthHandler := newIntegrationHealthHandler(t, cfg)
	userHandler := httphandler.NewUserHandler(mockUserService)

	handlers, err := router.NewHandlers(healthHandler, userHandler)
	if err != nil {
		panic(err)
	}

	engine := gin.New()
	engine.Use(injectTestAuthClaims())
	if err := router.RegisterRoutes(engine, handlers); err != nil {
		panic(err)
	}

	return engine, mockUserService
}

func injectTestAuthClaims() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(httpmiddleware.ClaimsKey, &auth.Claims{
			UserID:    1,
			TokenType: auth.TokenTypeAccess,
		})
		c.Next()
	}
}

func setupTestRouterWithInternalKey(t *testing.T, serviceKey string) (*gin.Engine, *MockUserService) {
	gin.SetMode(gin.TestMode)

	mockUserService := new(MockUserService)
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "orange-tv",
			Version: "1.0.0",
			Env:     constant.EnvTest,
		},
	}

	healthHandler := newIntegrationHealthHandler(t, cfg)
	userHandler := httphandler.NewUserHandler(mockUserService)

	handlers, err := router.NewHandlers(healthHandler, userHandler)
	if err != nil {
		panic(err)
	}
	handlers.InternalServiceKey = serviceKey

	engine := gin.New()
	if serviceKey == "" {
		engine.Use(injectTestAuthClaims())
	}
	if err := router.RegisterRoutes(engine, handlers); err != nil {
		panic(err)
	}

	return engine, mockUserService
}

func TestHealthCheckEndpoints(t *testing.T) {
	router, _ := setupTestRouter(t)

	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"health", "/health", "alive"},
		{"readiness", "/readiness", "ready"},
		{"liveness", "/liveness", "alive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.path)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var body map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&body)
			assert.NoError(t, err)
			resp.Body.Close()

			data, ok := body["data"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, tt.expected, data["status"])

			if tt.path == "/readiness" {
				checks, ok := data["checks"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, true, checks["database"])
			}
		})
	}
}

func TestUserCRUD(t *testing.T) {
	router, mockService := setupTestRouter(t)

	server := httptest.NewServer(router)
	defer server.Close()

	now := time.Now()

	t.Run("CreateUser", func(t *testing.T) {
		mockService.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).Return(&dto.UserResponse{ID: 1, Email: "test@example.com", Name: "Test User", Phone: "1234567890", Status: 1}, nil).Once()

		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			"name":     "Test User",
			"phone":    "1234567890",
		}
		body, _ := json.Marshal(reqBody)

		resp, err := http.Post(server.URL+"/api/client/v1/users", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		mockService.AssertExpectations(t)
	})

	t.Run("GetUser", func(t *testing.T) {
		user := &model.User{
			ID:        1,
			Email:     "test@example.com",
			Name:      "Test User",
			Phone:     "1234567890",
			Status:    1,
			CreatedAt: now,
			UpdatedAt: now,
		}
		mockService.On("GetByID", mock.Anything, int64(1)).Return(user, nil).Once()

		resp, err := http.Get(server.URL + "/api/client/v1/users/1")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		mockService.AssertExpectations(t)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		mockService.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*dto.UpdateUserRequest")).Return(&dto.UserResponse{ID: 1, Name: "Updated Name"}, nil).Once()

		reqBody := map[string]any{
			"name": "Updated Name",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", server.URL+"/api/client/v1/users/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		mockService.AssertExpectations(t)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		mockService.On("Delete", mock.Anything, int64(1)).Return(nil).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/api/client/v1/users/1", nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		mockService.AssertExpectations(t)
	})

	t.Run("ListUsers", func(t *testing.T) {
		users := []*model.User{
			{ID: 1, Email: "user1@example.com", Name: "User 1"},
			{ID: 2, Email: "user2@example.com", Name: "User 2"},
		}
		mockService.On("ListWithCount", mock.Anything, 10, 0).Return(users, 2, nil).Once()

		resp, err := http.Get(server.URL + "/api/client/v1/users?page=1&limit=10")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		mockService.AssertExpectations(t)
	})
}

func TestClientListUsersDeprecatedHeaders(t *testing.T) {
	engine, mockService := setupTestRouter(t)
	server := httptest.NewServer(engine)
	defer server.Close()

	users := []*model.User{{ID: 1, Email: "user1@example.com", Name: "User 1"}}
	mockService.On("ListWithCount", mock.Anything, 10, 0).Return(users, 1, nil).Once()

	resp, err := http.Get(server.URL + "/api/client/v1/users?page=1&limit=10")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "true", resp.Header.Get("Deprecation"))
	assert.NotEmpty(t, resp.Header.Get("Sunset"))
	assert.Contains(t, resp.Header.Get("Link"), router.PathClientV2Users)
	io.Copy(io.Discard, resp.Body)

	mockService.AssertExpectations(t)
}

func TestAdminGetUser(t *testing.T) {
	engine, mockService := setupTestRouter(t)
	server := httptest.NewServer(engine)
	defer server.Close()

	now := time.Now()
	user := &model.User{
		ID:        2,
		Email:     "admin@example.com",
		Name:      "Admin User",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockService.On("GetByID", mock.Anything, int64(2)).Return(user, nil).Once()

	resp, err := http.Get(server.URL + "/api/admin/v1/users/2")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	io.Copy(io.Discard, resp.Body)

	mockService.AssertExpectations(t)
}

func TestInternalGetUserWithServiceKey(t *testing.T) {
	engine, mockService := setupTestRouterWithInternalKey(t, "integration-internal-key")
	server := httptest.NewServer(engine)
	defer server.Close()

	now := time.Now()
	user := &model.User{
		ID:        3,
		Email:     "internal@example.com",
		Name:      "Internal User",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockService.On("GetByID", mock.Anything, int64(3)).Return(user, nil).Once()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/internal/v1/users/3", nil)
	assert.NoError(t, err)
	req.Header.Set("X-Internal-Service-Key", "integration-internal-key")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	io.Copy(io.Discard, resp.Body)

	mockService.AssertExpectations(t)
}

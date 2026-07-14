// Package service provides business logic layer.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/crypto"
	"github.com/ilaziness/orange-tv/internal/dto"
	"github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// UserService defines the contract for user business logic.
type UserService interface {
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (*model.User, error)
	Update(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
	ListWithCount(ctx context.Context, limit, offset int) ([]*model.User, int, error)
}

// userService handles user business logic.
type userService struct {
	userRepo repository.UserRepository
}

// 确保 userService 实现了 UserService 接口
var _ UserService = (*userService)(nil)

// NewUserService creates a new user service.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// Create creates a new user.
func (s *userService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errcode.UserAlreadyExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check user email: %w", err)
	}

	// 加密密码
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Phone:    req.Phone,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return toUserResponse(user), nil
}

// GetByID retrieves a user by ID.
func (s *userService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// ValidateCredentials validates user credentials (email and password).
func (s *userService) ValidateCredentials(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errcode.AuthFailed
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errcode.UserDisabled
	}

	// 验证密码
	if err := crypto.CheckPassword(password, user.Password); err != nil {
		return nil, errcode.AuthFailed
	}

	return user, nil
}

// Update updates an existing user.
func (s *userService) Update(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.UserNotFound, err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.Wrap(errcode.UserNotFound, err)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return toUserResponse(user), nil
}

// Delete deletes a user by ID.
func (s *userService) Delete(ctx context.Context, id int64) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.UserNotFound, err)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errcode.Wrap(errcode.UserNotFound, err)
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// List retrieves a list of users with pagination.
func (s *userService) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// ListWithCount retrieves a list of users with pagination and returns total count.
func (s *userService) ListWithCount(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

func toUserResponse(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

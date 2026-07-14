// Package repository provides data access layer.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

const (
	// userNotDeletedCond 是用户软删除过滤条件
	userNotDeletedCond = "deleted_at IS NULL"
	userByIDCond       = "id = ? AND "
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
	Count(ctx context.Context) (int, error)
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error
	WithTx(db bun.IDB) UserRepository
}

// UserRepo handles user data access operations.
type UserRepo struct {
	db bun.IDB
}

// 确保 UserRepo 实现了 UserRepository 接口
var _ UserRepository = (*UserRepo)(nil)

// NewUserRepo creates a new user repository.
func NewUserRepo(db *database.DB) UserRepository {
	return &UserRepo{db: db.DB}
}

// RunInTx runs a function within a database transaction.
func (r *UserRepo) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return r.db.RunInTx(ctx, &sql.TxOptions{}, fn)
}

// WithTx returns a new UserRepository that uses the provided transaction.
func (r *UserRepo) WithTx(db bun.IDB) UserRepository {
	return &UserRepo{db: db}
}

// Create creates a new user.
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where(userByIDCond+userNotDeletedCond, id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("email = ? AND "+userNotDeletedCond, email).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// Update updates an existing user.
func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	_, err := r.db.NewUpdate().Model(user).Where(userByIDCond+userNotDeletedCond, user.ID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user by ID (soft delete).
func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.User)(nil)).
		Set("deleted_at = ?", now).
		Where(userByIDCond+userNotDeletedCond, id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("failed to delete user: %w", sql.ErrNoRows)
	}
	return nil
}

// List retrieves a list of users with pagination.
func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.NewSelect().Model(&users).
		Where(userNotDeletedCond).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// Count returns the total count of users.
func (r *UserRepo) Count(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().Model((*model.User)(nil)).
		Where(userNotDeletedCond).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

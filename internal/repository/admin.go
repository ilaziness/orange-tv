package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

// AdminRepository provides admin and user-group persistence.
type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.Admins, error)
	GetByID(ctx context.Context, id int64) (*model.Admins, error)
	GetGroupByID(ctx context.Context, id int64) (*model.UserGroups, error)
	GetGroupByName(ctx context.Context, name string) (*model.UserGroups, error)
	Create(ctx context.Context, admin *model.Admins) error
	UpdateLastLogin(ctx context.Context, id int64, at time.Time) error
	ExistsUsername(ctx context.Context, username string) (bool, error)
}

type adminRepo struct {
	db *database.DB
}

// NewAdminRepo creates an AdminRepository.
func NewAdminRepo(db *database.DB) AdminRepository {
	return &adminRepo{db: db}
}

func (r *adminRepo) GetByUsername(ctx context.Context, username string) (*model.Admins, error) {
	admin := new(model.Admins)
	err := r.db.NewSelect().Model(admin).
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get admin by username: %w", err)
	}
	return admin, nil
}

func (r *adminRepo) GetByID(ctx context.Context, id int64) (*model.Admins, error) {
	admin := new(model.Admins)
	err := r.db.NewSelect().Model(admin).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get admin by id: %w", err)
	}
	return admin, nil
}

func (r *adminRepo) GetGroupByID(ctx context.Context, id int64) (*model.UserGroups, error) {
	group := new(model.UserGroups)
	err := r.db.NewSelect().Model(group).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user group by id: %w", err)
	}
	return group, nil
}

func (r *adminRepo) GetGroupByName(ctx context.Context, name string) (*model.UserGroups, error) {
	group := new(model.UserGroups)
	err := r.db.NewSelect().Model(group).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user group by name: %w", err)
	}
	return group, nil
}

func (r *adminRepo) Create(ctx context.Context, admin *model.Admins) error {
	_, err := r.db.NewInsert().Model(admin).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}

func (r *adminRepo) UpdateLastLogin(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.NewUpdate().Model((*model.Admins)(nil)).
		Set("last_login_at = ?", at).
		Set("updated_at = ?", at).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update admin last login: %w", err)
	}
	return nil
}

func (r *adminRepo) ExistsUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.Admins)(nil)).
		Where("username = ?", username).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check admin username: %w", err)
	}
	return exists, nil
}

// DB exposes bun.IDB for transaction helpers when needed by CLI.
func (r *adminRepo) DB() bun.IDB {
	return r.db
}

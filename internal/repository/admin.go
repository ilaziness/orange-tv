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

	// Admin CRUD (A3)
	ListAdmins(ctx context.Context, f AdminListFilter) ([]model.Admins, int, error)
	UpdateAdmin(ctx context.Context, admin *model.Admins) error
	SoftDeleteAdmin(ctx context.Context, id int64) error
	ExistsUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error)

	// User group CRUD (A4)
	ListGroups(ctx context.Context, f UserGroupListFilter) ([]model.UserGroups, int, error)
	CreateGroup(ctx context.Context, g *model.UserGroups) error
	UpdateGroup(ctx context.Context, g *model.UserGroups) error
	SoftDeleteGroup(ctx context.Context, id int64) error
	ExistsGroupNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error)

	// Regular users (A5)
	ListUsers(ctx context.Context, f UserListFilter) ([]model.Users, int, error)
	GetUserByID(ctx context.Context, id int64) (*model.Users, error)
	GetUserByUsername(ctx context.Context, username string) (*model.Users, error)
	CreateUser(ctx context.Context, u *model.Users) error
	UpdateUser(ctx context.Context, u *model.Users) error
	SoftDeleteUser(ctx context.Context, id int64) error
	ExistsUserUsername(ctx context.Context, username string) (bool, error)
	ExistsUserUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error)
}

// AdminListFilter filters admin list queries.
type AdminListFilter struct {
	Keyword string
	Status  *uint8
	GroupID *uint64
	Offset  int
	Limit   int
}

// UserGroupListFilter filters user group list queries.
type UserGroupListFilter struct {
	Keyword string
	Offset  int
	Limit   int
}

// UserListFilter filters regular user list queries.
type UserListFilter struct {
	Keyword string
	Status  *uint8
	Offset  int
	Limit   int
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

// ===== Admin CRUD (A3) =====

func (r *adminRepo) ListAdmins(ctx context.Context, f AdminListFilter) ([]model.Admins, int, error) {
	items := make([]model.Admins, 0, f.Limit)
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if f.Keyword != "" {
		q = q.Where("username LIKE ? OR email LIKE ?", "%"+f.Keyword+"%", "%"+f.Keyword+"%")
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.GroupID != nil {
		q = q.Where("group_id = ?", *f.GroupID)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count admins: %w", err)
	}
	if err := q.Order("id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list admins: %w", err)
	}
	return items, total, nil
}

func (r *adminRepo) UpdateAdmin(ctx context.Context, admin *model.Admins) error {
	_, err := r.db.NewUpdate().Model(admin).
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update admin: %w", err)
	}
	return nil
}

func (r *adminRepo) SoftDeleteAdmin(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Admins)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete admin: %w", err)
	}
	return nil
}

func (r *adminRepo) ExistsUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.Admins)(nil)).
		Where("username = ?", username).
		Where("id != ?", excludeID).
		Where("deleted_at IS NULL").
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check admin username: %w", err)
	}
	return exists, nil
}

// ===== User group CRUD (A4) =====

func (r *adminRepo) ListGroups(ctx context.Context, f UserGroupListFilter) ([]model.UserGroups, int, error) {
	items := make([]model.UserGroups, 0, f.Limit)
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if f.Keyword != "" {
		q = q.Where("name LIKE ? OR description LIKE ?", "%"+f.Keyword+"%", "%"+f.Keyword+"%")
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count user groups: %w", err)
	}
	if err := q.Order("id ASC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list user groups: %w", err)
	}
	return items, total, nil
}

func (r *adminRepo) CreateGroup(ctx context.Context, g *model.UserGroups) error {
	_, err := r.db.NewInsert().Model(g).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create user group: %w", err)
	}
	return nil
}

func (r *adminRepo) UpdateGroup(ctx context.Context, g *model.UserGroups) error {
	_, err := r.db.NewUpdate().Model(g).
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update user group: %w", err)
	}
	return nil
}

func (r *adminRepo) SoftDeleteGroup(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.UserGroups)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete user group: %w", err)
	}
	return nil
}

func (r *adminRepo) ExistsGroupNameExcludeID(ctx context.Context, name string, excludeID int64) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.UserGroups)(nil)).
		Where("name = ?", name).
		Where("id != ?", excludeID).
		Where("deleted_at IS NULL").
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check user group name: %w", err)
	}
	return exists, nil
}

// ===== Regular users (A5) =====

func (r *adminRepo) ListUsers(ctx context.Context, f UserListFilter) ([]model.Users, int, error) {
	items := make([]model.Users, 0, f.Limit)
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if f.Keyword != "" {
		q = q.Where("username LIKE ? OR email LIKE ?", "%"+f.Keyword+"%", "%"+f.Keyword+"%")
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	if err := q.Order("id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return items, total, nil
}

func (r *adminRepo) GetUserByID(ctx context.Context, id int64) (*model.Users, error) {
	u := new(model.Users)
	err := r.db.NewSelect().Model(u).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *adminRepo) GetUserByUsername(ctx context.Context, username string) (*model.Users, error) {
	u := new(model.Users)
	err := r.db.NewSelect().Model(u).
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

func (r *adminRepo) CreateUser(ctx context.Context, u *model.Users) error {
	_, err := r.db.NewInsert().Model(u).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *adminRepo) UpdateUser(ctx context.Context, u *model.Users) error {
	_, err := r.db.NewUpdate().Model(u).
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *adminRepo) SoftDeleteUser(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Users)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *adminRepo) ExistsUserUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.Users)(nil)).
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check user username: %w", err)
	}
	return exists, nil
}

func (r *adminRepo) ExistsUserUsernameExcludeID(ctx context.Context, username string, excludeID int64) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.Users)(nil)).
		Where("username = ?", username).
		Where("id != ?", excludeID).
		Where("deleted_at IS NULL").
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check user username: %w", err)
	}
	return exists, nil
}

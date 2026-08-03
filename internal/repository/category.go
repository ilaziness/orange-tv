package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
)

// CategoryRepository provides category persistence.
type CategoryRepository interface {
	List(ctx context.Context, onlyEnabled bool) ([]model.Categories, error)
	GetByID(ctx context.Context, id uint32) (*model.Categories, error)
	ExistsName(ctx context.Context, name string, excludeID uint32) (bool, error)
	Create(ctx context.Context, c *model.Categories) error
	Update(ctx context.Context, c *model.Categories) error
	SoftDelete(ctx context.Context, id uint32) error
	CountChildren(ctx context.Context, parentID uint32) (int, error)
	CountVideos(ctx context.Context, categoryID uint32) (int, error)
	ListIDs(ctx context.Context) ([]model.Categories, error)
}

type categoryRepo struct {
	db *database.DB
}

// NewCategoryRepo creates a CategoryRepository.
func NewCategoryRepo(db *database.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) List(ctx context.Context, onlyEnabled bool) ([]model.Categories, error) {
	var items []model.Categories
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if onlyEnabled {
		q = q.Where("status = ?", 1)
	}
	err := q.OrderExpr("sort_order ASC, id ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return items, nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id uint32) (*model.Categories, error) {
	item := new(model.Categories)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return item, nil
}

func (r *categoryRepo) ExistsName(ctx context.Context, name string, excludeID uint32) (bool, error) {
	q := r.db.NewSelect().Model((*model.Categories)(nil)).
		Where("name = ?", name).
		Where("deleted_at IS NULL")
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check category name: %w", err)
	}
	return exists, nil
}

func (r *categoryRepo) Create(ctx context.Context, c *model.Categories) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (r *categoryRepo) Update(ctx context.Context, c *model.Categories) error {
	_, err := r.db.NewUpdate().Model(c).
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (r *categoryRepo) SoftDelete(ctx context.Context, id uint32) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Categories)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (r *categoryRepo) CountChildren(ctx context.Context, parentID uint32) (int, error) {
	n, err := r.db.NewSelect().Model((*model.Categories)(nil)).
		Where("parent_id = ?", parentID).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count category children: %w", err)
	}
	return n, nil
}

func (r *categoryRepo) CountVideos(ctx context.Context, categoryID uint32) (int, error) {
	n, err := r.db.NewSelect().Model((*model.Videos)(nil)).
		Where("category_id = ?", categoryID).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count category videos: %w", err)
	}
	return n, nil
}

func (r *categoryRepo) ListIDs(ctx context.Context) ([]model.Categories, error) {
	return r.List(ctx, false)
}

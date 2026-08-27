package repository

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

// AdRepository manages advertisements.
type AdRepository interface {
	ListAll(ctx context.Context, offset, limit int) ([]model.Advertisements, int, error)
	ListByStatus(ctx context.Context, status *uint8) ([]model.Advertisements, error)
	ListByScene(ctx context.Context, scene string, status *uint8) ([]model.Advertisements, error)
	Get(ctx context.Context, id uint32) (*model.Advertisements, error)
	GetByKey(ctx context.Context, adKey string) (*model.Advertisements, error)
	Create(ctx context.Context, a *model.Advertisements) error
	Update(ctx context.Context, a *model.Advertisements) error
	Delete(ctx context.Context, id uint32) error
}

type adRepo struct {
	db bun.IDB
}

// NewAdRepo creates an AdRepository.
func NewAdRepo(db *database.DB) AdRepository {
	return &adRepo{db: db}
}

func (r *adRepo) ListAll(ctx context.Context, offset, limit int) ([]model.Advertisements, int, error) {
	items := make([]model.Advertisements, 0, limit)
	q := r.db.NewSelect().Model(&items)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count advertisements: %w", err)
	}
	if err := q.Order("sort ASC, id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list advertisements: %w", err)
	}
	return items, total, nil
}

// ListByStatus returns advertisements optionally filtered by status.
func (r *adRepo) ListByStatus(ctx context.Context, status *uint8) ([]model.Advertisements, error) {
	return r.listAds(ctx, "", status)
}

// ListByScene returns advertisements for the given scene, optionally filtered by status.
func (r *adRepo) ListByScene(ctx context.Context, scene string, status *uint8) ([]model.Advertisements, error) {
	return r.listAds(ctx, scene, status)
}

func (r *adRepo) listAds(ctx context.Context, scene string, status *uint8) ([]model.Advertisements, error) {
	items := make([]model.Advertisements, 0, 20)
	q := r.db.NewSelect().Model(&items)
	if scene != "" {
		q = q.Where("scene = ?", scene)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Order("sort ASC, id DESC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("list advertisements: %w", err)
	}
	return items, nil
}

func (r *adRepo) Get(ctx context.Context, id uint32) (*model.Advertisements, error) {
	a := new(model.Advertisements)
	found, err := notFoundOrErr(r.db.NewSelect().Model(a).Where("id = ?", id).Scan(ctx), "get advertisement")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return a, nil
}

func (r *adRepo) GetByKey(ctx context.Context, adKey string) (*model.Advertisements, error) {
	a := new(model.Advertisements)
	found, err := notFoundOrErr(r.db.NewSelect().Model(a).Where("ad_key = ?", adKey).Scan(ctx), "get advertisement by key")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return a, nil
}

func (r *adRepo) Create(ctx context.Context, a *model.Advertisements) error {
	_, err := r.db.NewInsert().Model(a).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create advertisement: %w", err)
	}
	return nil
}

func (r *adRepo) Update(ctx context.Context, a *model.Advertisements) error {
	_, err := r.db.NewUpdate().Model(a).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("update advertisement: %w", err)
	}
	return nil
}

func (r *adRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.db.NewDelete().Model((*model.Advertisements)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete advertisement: %w", err)
	}
	return nil
}

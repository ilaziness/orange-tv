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

// LiveListFilter filters live channel queries.
type LiveListFilter struct {
	Category   string
	Keyword    string
	Status     *uint8
	OnlyOnline bool
	Offset     int
	Limit      int
}

// LiveRepository provides live channel persistence.
type LiveRepository interface {
	List(ctx context.Context, f LiveListFilter) ([]model.LiveChannels, int, error)
	GetByID(ctx context.Context, id int64) (*model.LiveChannels, error)
	Create(ctx context.Context, m *model.LiveChannels) error
	Update(ctx context.Context, m *model.LiveChannels) error
	SoftDelete(ctx context.Context, id int64) error
}

type liveRepo struct {
	db *database.DB
}

// NewLiveRepo creates a LiveRepository.
func NewLiveRepo(db *database.DB) LiveRepository {
	return &liveRepo{db: db}
}

func (r *liveRepo) List(ctx context.Context, f LiveListFilter) ([]model.LiveChannels, int, error) {
	var items []model.LiveChannels
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("(name LIKE ? OR description LIKE ?)", like, like)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.OnlyOnline {
		q = q.Where("status = ?", 1)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count live channels: %w", err)
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if err := q.OrderExpr("sort_order ASC, id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list live channels: %w", err)
	}
	return items, total, nil
}

func (r *liveRepo) GetByID(ctx context.Context, id int64) (*model.LiveChannels, error) {
	item := new(model.LiveChannels)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live channel: %w", err)
	}
	return item, nil
}

func (r *liveRepo) Create(ctx context.Context, m *model.LiveChannels) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create live channel: %w", err)
	}
	return nil
}

func (r *liveRepo) Update(ctx context.Context, m *model.LiveChannels) error {
	now := time.Now()
	m.UpdatedAt = &now
	_, err := r.db.NewUpdate().Model(m).
		Column("name", "category", "stream_url", "logo", "description", "sort_order", "status", "updated_at").
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update live channel: %w", err)
	}
	return nil
}

func (r *liveRepo) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.LiveChannels)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("soft delete live channel: %w", err)
	}
	return nil
}

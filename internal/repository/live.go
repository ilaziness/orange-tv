package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
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
	ListAll(ctx context.Context) ([]model.LiveChannels, error)
	GetByID(ctx context.Context, id uint32) (*model.LiveChannels, error)
	Create(ctx context.Context, m *model.LiveChannels) error
	BatchCreate(ctx context.Context, items []model.LiveChannels) error
	Update(ctx context.Context, m *model.LiveChannels) error
	Delete(ctx context.Context, id uint32) error
	DeleteByIDs(ctx context.Context, ids []uint32) error
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
	q := r.db.NewSelect().Model(&items)
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

func (r *liveRepo) GetByID(ctx context.Context, id uint32) (*model.LiveChannels, error) {
	item := new(model.LiveChannels)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
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
	_, err := r.db.NewUpdate().Model(m).
		Column("name", "category", "stream_url", "logo", "description", "sort_order", "status", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update live channel: %w", err)
	}
	return nil
}

func (r *liveRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.db.NewDelete().Model((*model.LiveChannels)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete live channel: %w", err)
	}
	return nil
}

func (r *liveRepo) ListAll(ctx context.Context) ([]model.LiveChannels, error) {
	var items []model.LiveChannels
	err := r.db.NewSelect().Model(&items).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all live channels: %w", err)
	}
	return items, nil
}

func (r *liveRepo) BatchCreate(ctx context.Context, items []model.LiveChannels) error {
	if len(items) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&items).Exec(ctx)
	if err != nil {
		return fmt.Errorf("batch create live channels: %w", err)
	}
	return nil
}

func (r *liveRepo) DeleteByIDs(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.NewDelete().Model((*model.LiveChannels)(nil)).
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete live channels by ids: %w", err)
	}
	return nil
}

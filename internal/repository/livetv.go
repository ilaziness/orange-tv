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

// LiveTVListFilter filters livetv channel queries.
type LiveTVListFilter struct {
	Category   string
	Keyword    string
	Status     *uint8
	OnlyOnline bool
	Offset     int
	Limit      int
}

// LiveTVRepository provides livetv channel persistence.
type LiveTVRepository interface {
	List(ctx context.Context, f LiveTVListFilter) ([]model.LivetvChannels, int, error)
	ListAll(ctx context.Context) ([]model.LivetvChannels, error)
	GetByID(ctx context.Context, id uint32) (*model.LivetvChannels, error)
	Create(ctx context.Context, m *model.LivetvChannels) error
	BatchCreate(ctx context.Context, items []model.LivetvChannels) error
	Update(ctx context.Context, m *model.LivetvChannels) error
	Delete(ctx context.Context, id uint32) error
	DeleteByIDs(ctx context.Context, ids []uint32) error
}

type liveTVRepo struct {
	db *database.DB
}

// NewLiveTVRepo creates a LiveTVRepository.
func NewLiveTVRepo(db *database.DB) LiveTVRepository {
	return &liveTVRepo{db: db}
}

func (r *liveTVRepo) List(ctx context.Context, f LiveTVListFilter) ([]model.LivetvChannels, int, error) {
	var items []model.LivetvChannels
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
		return nil, 0, fmt.Errorf("count livetv channels: %w", err)
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if err := q.OrderExpr("sort_order ASC, id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list livetv channels: %w", err)
	}
	return items, total, nil
}

func (r *liveTVRepo) GetByID(ctx context.Context, id uint32) (*model.LivetvChannels, error) {
	item := new(model.LivetvChannels)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get livetv channel: %w", err)
	}
	return item, nil
}

func (r *liveTVRepo) Create(ctx context.Context, m *model.LivetvChannels) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create livetv channel: %w", err)
	}
	return nil
}

func (r *liveTVRepo) Update(ctx context.Context, m *model.LivetvChannels) error {
	_, err := r.db.NewUpdate().Model(m).
		Column("name", "category", "stream_url", "logo", "description", "sort_order", "status", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update livetv channel: %w", err)
	}
	return nil
}

func (r *liveTVRepo) Delete(ctx context.Context, id uint32) error {
	_, err := r.db.NewDelete().Model((*model.LivetvChannels)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete livetv channel: %w", err)
	}
	return nil
}

func (r *liveTVRepo) ListAll(ctx context.Context) ([]model.LivetvChannels, error) {
	var items []model.LivetvChannels
	err := r.db.NewSelect().Model(&items).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all livetv channels: %w", err)
	}
	return items, nil
}

func (r *liveTVRepo) BatchCreate(ctx context.Context, items []model.LivetvChannels) error {
	if len(items) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&items).Exec(ctx)
	if err != nil {
		return fmt.Errorf("batch create livetv channels: %w", err)
	}
	return nil
}

func (r *liveTVRepo) DeleteByIDs(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.NewDelete().Model((*model.LivetvChannels)(nil)).
		Where("id IN (?)", bun.List(ids)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete livetv channels by ids: %w", err)
	}
	return nil
}

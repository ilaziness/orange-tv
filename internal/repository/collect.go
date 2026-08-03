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

// CollectSourceListFilter filters collect source queries.
type CollectSourceListFilter struct {
	Status *uint8
	Offset int
	Limit  int
}

// CollectLogListFilter filters collect log queries.
type CollectLogListFilter struct {
	SourceID uint32
	Offset   int
	Limit    int
}

// CollectRepository manages collect sources, category maps and logs.
type CollectRepository interface {
	ListSources(ctx context.Context, f CollectSourceListFilter) ([]model.CollectSources, int, error)
	ListEnabledCronSources(ctx context.Context) ([]model.CollectSources, error)
	GetSource(ctx context.Context, id uint32) (*model.CollectSources, error)
	CreateSource(ctx context.Context, m *model.CollectSources) error
	UpdateSource(ctx context.Context, m *model.CollectSources) error
	SoftDeleteSource(ctx context.Context, id uint32) error
	TouchLastCollect(ctx context.Context, id uint32, at time.Time) error

	ListCategories(ctx context.Context, sourceID uint32) ([]model.CollectSourceCategories, error)
	ReplaceCategories(ctx context.Context, sourceID uint32, items []model.CollectSourceCategories) error

	CreateLog(ctx context.Context, m *model.CollectLogs) error
	UpdateLog(ctx context.Context, m *model.CollectLogs) error
	IncrementLogCount(ctx context.Context, logID uint32, count int) error
	ListLogs(ctx context.Context, f CollectLogListFilter) ([]model.CollectLogs, int, error)

	// FindVideoByTitleYear finds undeleleted video by title+year (dedup).
	FindVideoByTitleYear(ctx context.Context, title string, year int32) (*model.Videos, error)
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error
}

type collectRepo struct {
	db bun.IDB
}

// NewCollectRepo creates a CollectRepository.
func NewCollectRepo(db *database.DB) CollectRepository {
	return &collectRepo{db: db}
}

func (r *collectRepo) ListSources(ctx context.Context, f CollectSourceListFilter) ([]model.CollectSources, int, error) {
	var items []model.CollectSources
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count collect sources: %w", err)
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if err := q.OrderExpr("id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list collect sources: %w", err)
	}
	return items, total, nil
}

func (r *collectRepo) ListEnabledCronSources(ctx context.Context) ([]model.CollectSources, error) {
	var items []model.CollectSources
	err := r.db.NewSelect().Model(&items).
		Where("deleted_at IS NULL").
		Where("schedule_enabled = ?", 1).
		Where("cron_expr <> ''").
		OrderExpr("id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cron collect sources: %w", err)
	}
	return items, nil
}

func (r *collectRepo) GetSource(ctx context.Context, id uint32) (*model.CollectSources, error) {
	item := new(model.CollectSources)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get collect source: %w", err)
	}
	return item, nil
}

func (r *collectRepo) CreateSource(ctx context.Context, m *model.CollectSources) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create collect source: %w", err)
	}
	return nil
}

func (r *collectRepo) UpdateSource(ctx context.Context, m *model.CollectSources) error {
	_, err := r.db.NewUpdate().Model(m).
		Column("name", "type", "collect_url", "api_key", "cron_expr", "play_source_id", "status", "schedule_enabled", "data_range", "updated_at").
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update collect source: %w", err)
	}
	return nil
}

func (r *collectRepo) SoftDeleteSource(ctx context.Context, id uint32) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.CollectSources)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("soft delete collect source: %w", err)
	}
	return nil
}

func (r *collectRepo) TouchLastCollect(ctx context.Context, id uint32, at time.Time) error {
	_, err := r.db.NewUpdate().Model((*model.CollectSources)(nil)).
		Set("last_collect_at = ?", at).
		Set("updated_at = ?", at).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("touch last collect: %w", err)
	}
	return nil
}

func (r *collectRepo) ListCategories(ctx context.Context, sourceID uint32) ([]model.CollectSourceCategories, error) {
	var items []model.CollectSourceCategories
	err := r.db.NewSelect().Model(&items).
		Where("source_id = ?", sourceID).
		OrderExpr("id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collect categories: %w", err)
	}
	return items, nil
}

func (r *collectRepo) ReplaceCategories(ctx context.Context, sourceID uint32, items []model.CollectSourceCategories) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().Model((*model.CollectSourceCategories)(nil)).
			Where("source_id = ?", sourceID).
			Exec(ctx); err != nil {
			return fmt.Errorf("clear collect categories: %w", err)
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].SourceID = sourceID
			items[i].ID = 0
		}
		if _, err := tx.NewInsert().Model(&items).Exec(ctx); err != nil {
			return fmt.Errorf("insert collect categories: %w", err)
		}
		return nil
	})
}

func (r *collectRepo) CreateLog(ctx context.Context, m *model.CollectLogs) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create collect log: %w", err)
	}
	return nil
}

func (r *collectRepo) UpdateLog(ctx context.Context, m *model.CollectLogs) error {
	_, err := r.db.NewUpdate().Model(m).
		Column("status", "duration_sec").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update collect log: %w", err)
	}
	return nil
}

func (r *collectRepo) IncrementLogCount(ctx context.Context, logID uint32, count int) error {
	_, err := r.db.NewUpdate().Model((*model.CollectLogs)(nil)).
		Set("collect_count = collect_count + ?", count).
		Where("id = ?", logID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("increment collect log count: %w", err)
	}
	return nil
}

func (r *collectRepo) ListLogs(ctx context.Context, f CollectLogListFilter) ([]model.CollectLogs, int, error) {
	var items []model.CollectLogs
	q := r.db.NewSelect().Model(&items)
	if f.SourceID > 0 {
		q = q.Where("source_id = ?", f.SourceID)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count collect logs: %w", err)
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if err := q.OrderExpr("id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list collect logs: %w", err)
	}
	return items, total, nil
}

func (r *collectRepo) FindVideoByTitleYear(ctx context.Context, title string, year int32) (*model.Videos, error) {
	item := new(model.Videos)
	err := r.db.NewSelect().Model(item).
		Where("title = ?", title).
		Where("year = ?", year).
		Where("deleted_at IS NULL").
		OrderExpr("id DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find video by title year: %w", err)
	}
	return item, nil
}

func (r *collectRepo) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return r.db.RunInTx(ctx, nil, fn)
}

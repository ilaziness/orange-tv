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

// ThemeRepository manages themes.
type ThemeRepository interface {
	List(ctx context.Context) ([]model.Themes, error)
	GetByID(ctx context.Context, id int64) (*model.Themes, error)
	GetActive(ctx context.Context) (*model.Themes, error)
	GetByIdentifier(ctx context.Context, identifier string) (*model.Themes, error)
	Create(ctx context.Context, m *model.Themes) error
	Update(ctx context.Context, m *model.Themes) error
	Activate(ctx context.Context, id int64) error
	EnsureDefault(ctx context.Context, m *model.Themes) error
}

type themeRepo struct {
	db *database.DB
}

// NewThemeRepo creates a ThemeRepository.
func NewThemeRepo(db *database.DB) ThemeRepository {
	return &themeRepo{db: db}
}

func (r *themeRepo) List(ctx context.Context) ([]model.Themes, error) {
	var items []model.Themes
	err := r.db.NewSelect().Model(&items).
		Where("deleted_at IS NULL").
		OrderExpr("is_active DESC, id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list themes: %w", err)
	}
	return items, nil
}

func (r *themeRepo) GetByID(ctx context.Context, id int64) (*model.Themes, error) {
	item := new(model.Themes)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get theme: %w", err)
	}
	return item, nil
}

func (r *themeRepo) GetActive(ctx context.Context) (*model.Themes, error) {
	item := new(model.Themes)
	err := r.db.NewSelect().Model(item).
		Where("is_active = ?", 1).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active theme: %w", err)
	}
	return item, nil
}

func (r *themeRepo) GetByIdentifier(ctx context.Context, identifier string) (*model.Themes, error) {
	item := new(model.Themes)
	err := r.db.NewSelect().Model(item).
		Where("identifier = ?", identifier).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get theme by identifier: %w", err)
	}
	return item, nil
}

func (r *themeRepo) Create(ctx context.Context, m *model.Themes) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create theme: %w", err)
	}
	return nil
}

func (r *themeRepo) Update(ctx context.Context, m *model.Themes) error {
	now := time.Now()
	m.UpdatedAt = &now
	_, err := r.db.NewUpdate().Model(m).
		Column("name", "version", "author", "description", "preview_image", "config", "custom_css", "custom_js", "updated_at").
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update theme: %w", err)
	}
	return nil
}

func (r *themeRepo) Activate(ctx context.Context, id int64) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		now := time.Now()
		if _, err := tx.NewUpdate().Model((*model.Themes)(nil)).
			Set("is_active = ?", 0).
			Set("updated_at = ?", now).
			Where("deleted_at IS NULL").
			Where("is_active = ?", 1).
			Exec(ctx); err != nil {
			return fmt.Errorf("deactivate themes: %w", err)
		}
		res, err := tx.NewUpdate().Model((*model.Themes)(nil)).
			Set("is_active = ?", 1).
			Set("updated_at = ?", now).
			Where("id = ?", id).
			Where("deleted_at IS NULL").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("activate theme: %w", err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (r *themeRepo) EnsureDefault(ctx context.Context, m *model.Themes) error {
	existing, err := r.GetByIdentifier(ctx, m.Identifier)
	if err != nil {
		return err
	}
	if existing != nil {
		// ensure at least one active
		active, err := r.GetActive(ctx)
		if err != nil {
			return err
		}
		if active == nil {
			return r.Activate(ctx, existing.ID)
		}
		return nil
	}
	// create default as the sole active theme
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		now := time.Now()
		if _, err := tx.NewUpdate().Model((*model.Themes)(nil)).
			Set("is_active = ?", 0).
			Set("updated_at = ?", now).
			Where("deleted_at IS NULL").
			Where("is_active = ?", 1).
			Exec(ctx); err != nil {
			return fmt.Errorf("deactivate themes: %w", err)
		}
		m.IsDefault = 1
		m.IsActive = 1
		if _, err := tx.NewInsert().Model(m).Exec(ctx); err != nil {
			return fmt.Errorf("create default theme: %w", err)
		}
		return nil
	})
}

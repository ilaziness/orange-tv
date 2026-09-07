package repository

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

// MediaRepository manages media_assets rows.
type MediaRepository interface {
	Create(ctx context.Context, m *model.MediaAssets) error
	GetByID(ctx context.Context, id uint64) (*model.MediaAssets, error)
	List(ctx context.Context, mediaType string, offset, limit int) ([]model.MediaAssets, int, error)
	Delete(ctx context.Context, id uint64) error
}

type mediaRepo struct {
	db bun.IDB
}

// NewMediaRepo creates a MediaRepository.
func NewMediaRepo(db *database.DB) MediaRepository {
	return &mediaRepo{db: db}
}

func (r *mediaRepo) Create(ctx context.Context, m *model.MediaAssets) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create media asset: %w", err)
	}
	return nil
}

func (r *mediaRepo) GetByID(ctx context.Context, id uint64) (*model.MediaAssets, error) {
	m := new(model.MediaAssets)
	found, err := notFoundOrErr(r.db.NewSelect().Model(m).Where("id = ?", id).Scan(ctx), "get media asset")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return m, nil
}

func (r *mediaRepo) List(ctx context.Context, mediaType string, offset, limit int) ([]model.MediaAssets, int, error) {
	items := make([]model.MediaAssets, 0, limit)
	q := r.db.NewSelect().Model(&items)
	if mediaType != "" {
		q = q.Where("media_type = ?", mediaType)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count media assets: %w", err)
	}
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list media assets: %w", err)
	}
	return items, total, nil
}

func (r *mediaRepo) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.NewDelete().Model((*model.MediaAssets)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete media asset: %w", err)
	}
	return nil
}

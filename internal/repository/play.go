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

// PlayRepository manages play sources and episodes.
type PlayRepository interface {
	ListSources(ctx context.Context) ([]model.PlaySources, error)
	GetSource(ctx context.Context, id int64) (*model.PlaySources, error)
	ExistsSourceName(ctx context.Context, name string, excludeID int64) (bool, error)
	CreateSource(ctx context.Context, m *model.PlaySources) error
	UpdateSource(ctx context.Context, m *model.PlaySources) error
	SoftDeleteSource(ctx context.Context, id int64) error
	CountEpisodesBySource(ctx context.Context, sourceID int64) (int, error)

	ListEpisodes(ctx context.Context, videoID, sourceID int64, offset, limit int) ([]model.PlayEpisodes, int, error)
	ListEpisodesByVideo(ctx context.Context, videoID int64, onlyEnabled bool) ([]model.PlayEpisodes, error)
	GetEpisode(ctx context.Context, id int64) (*model.PlayEpisodes, error)
	ExistsEpisode(ctx context.Context, videoID, sourceID int64, episodeNumber int32, excludeID int64) (bool, error)
	CreateEpisode(ctx context.Context, m *model.PlayEpisodes) error
	UpdateEpisode(ctx context.Context, m *model.PlayEpisodes) error
	SoftDeleteEpisode(ctx context.Context, id int64) error
}

type playRepo struct {
	db *database.DB
}

// NewPlayRepo creates a PlayRepository.
func NewPlayRepo(db *database.DB) PlayRepository {
	return &playRepo{db: db}
}

func (r *playRepo) ListSources(ctx context.Context) ([]model.PlaySources, error) {
	var items []model.PlaySources
	err := r.db.NewSelect().Model(&items).
		Where("deleted_at IS NULL").
		OrderExpr("sort_order ASC, id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list play sources: %w", err)
	}
	return items, nil
}

func (r *playRepo) GetSource(ctx context.Context, id int64) (*model.PlaySources, error) {
	item := new(model.PlaySources)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get play source: %w", err)
	}
	return item, nil
}

func (r *playRepo) ExistsSourceName(ctx context.Context, name string, excludeID int64) (bool, error) {
	q := r.db.NewSelect().Model((*model.PlaySources)(nil)).
		Where("name = ?", name).
		Where("deleted_at IS NULL")
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check play source name: %w", err)
	}
	return exists, nil
}

func (r *playRepo) CreateSource(ctx context.Context, m *model.PlaySources) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create play source: %w", err)
	}
	return nil
}

func (r *playRepo) UpdateSource(ctx context.Context, m *model.PlaySources) error {
	now := time.Now()
	m.UpdatedAt = &now
	_, err := r.db.NewUpdate().Model(m).WherePK().Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("update play source: %w", err)
	}
	return nil
}

func (r *playRepo) SoftDeleteSource(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.PlaySources)(nil)).
		Set("deleted_at = ?", now).Set("updated_at = ?", now).
		Where("id = ?", id).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete play source: %w", err)
	}
	return nil
}

func (r *playRepo) CountEpisodesBySource(ctx context.Context, sourceID int64) (int, error) {
	n, err := r.db.NewSelect().Model((*model.PlayEpisodes)(nil)).
		Where("source_id = ?", sourceID).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count episodes by source: %w", err)
	}
	return n, nil
}

func (r *playRepo) ListEpisodes(ctx context.Context, videoID, sourceID int64, offset, limit int) ([]model.PlayEpisodes, int, error) {
	var items []model.PlayEpisodes
	q := r.db.NewSelect().Model(&items).
		Where("video_id = ?", videoID).
		Where("source_id = ?", sourceID).
		Where("deleted_at IS NULL")
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count play episodes: %w", err)
	}
	if err := q.OrderExpr("sort_order ASC, episode_number ASC, id ASC").
		Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list play episodes: %w", err)
	}
	return items, total, nil
}

func (r *playRepo) ListEpisodesByVideo(ctx context.Context, videoID int64, onlyEnabled bool) ([]model.PlayEpisodes, error) {
	var items []model.PlayEpisodes
	q := r.db.NewSelect().Model(&items).
		Where("video_id = ?", videoID).
		Where("deleted_at IS NULL")
	if onlyEnabled {
		q = q.Where("status = ?", 1)
	}
	if err := q.OrderExpr("source_id ASC, sort_order ASC, episode_number ASC, id ASC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("list video play episodes: %w", err)
	}
	return items, nil
}

func (r *playRepo) GetEpisode(ctx context.Context, id int64) (*model.PlayEpisodes, error) {
	item := new(model.PlayEpisodes)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get play episode: %w", err)
	}
	return item, nil
}

func (r *playRepo) ExistsEpisode(ctx context.Context, videoID, sourceID int64, episodeNumber int32, excludeID int64) (bool, error) {
	q := r.db.NewSelect().Model((*model.PlayEpisodes)(nil)).
		Where("video_id = ?", videoID).
		Where("source_id = ?", sourceID).
		Where("episode_number = ?", episodeNumber)
	// Unique index covers all rows including soft-deleted. Check both for conflict messaging.
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check play episode: %w", err)
	}
	return exists, nil
}

func (r *playRepo) CreateEpisode(ctx context.Context, m *model.PlayEpisodes) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create play episode: %w", err)
	}
	return nil
}

func (r *playRepo) UpdateEpisode(ctx context.Context, m *model.PlayEpisodes) error {
	now := time.Now()
	m.UpdatedAt = &now
	_, err := r.db.NewUpdate().Model(m).WherePK().Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("update play episode: %w", err)
	}
	return nil
}

func (r *playRepo) SoftDeleteEpisode(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.PlayEpisodes)(nil)).
		Set("deleted_at = ?", now).Set("updated_at = ?", now).
		Where("id = ?", id).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete play episode: %w", err)
	}
	return nil
}

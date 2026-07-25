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

// MetadataRepository manages directors, actors and tags.
type MetadataRepository interface {
	// Directors
	ListDirectors(ctx context.Context, keyword string, offset, limit int) ([]model.Directors, int, error)
	GetDirector(ctx context.Context, id int64) (*model.Directors, error)
	GetDirectorByName(ctx context.Context, name string) (*model.Directors, error)
	GetDirectorsByIDs(ctx context.Context, ids []uint64) ([]model.Directors, error)
	ExistsDirectorName(ctx context.Context, name string, excludeID int64) (bool, error)
	CreateDirector(ctx context.Context, m *model.Directors) error
	UpdateDirector(ctx context.Context, m *model.Directors) error
	SoftDeleteDirector(ctx context.Context, id int64) error
	CountDirectorRefs(ctx context.Context, id int64) (int, error)

	// Actors
	ListActors(ctx context.Context, keyword string, offset, limit int) ([]model.Actors, int, error)
	GetActor(ctx context.Context, id int64) (*model.Actors, error)
	GetActorByName(ctx context.Context, name string) (*model.Actors, error)
	GetActorsByIDs(ctx context.Context, ids []uint64) ([]model.Actors, error)
	ExistsActorName(ctx context.Context, name string, excludeID int64) (bool, error)
	CreateActor(ctx context.Context, m *model.Actors) error
	UpdateActor(ctx context.Context, m *model.Actors) error
	SoftDeleteActor(ctx context.Context, id int64) error
	CountActorRefs(ctx context.Context, id int64) (int, error)

	// Tags
	ListTags(ctx context.Context, keyword string, offset, limit int) ([]model.Tags, int, error)
	GetTag(ctx context.Context, id int64) (*model.Tags, error)
	GetTagByName(ctx context.Context, name string) (*model.Tags, error)
	GetTagsByIDs(ctx context.Context, ids []uint64) ([]model.Tags, error)
	ExistsTagName(ctx context.Context, name string, excludeID int64) (bool, error)
	CreateTag(ctx context.Context, m *model.Tags) error
	UpdateTag(ctx context.Context, m *model.Tags) error
	SoftDeleteTag(ctx context.Context, id int64) error
	CountTagRefs(ctx context.Context, id int64) (int, error)
}

type metadataRepo struct {
	db *database.DB
}

// NewMetadataRepo creates a MetadataRepository.
func NewMetadataRepo(db *database.DB) MetadataRepository {
	return &metadataRepo{db: db}
}

func (r *metadataRepo) ListDirectors(ctx context.Context, keyword string, offset, limit int) ([]model.Directors, int, error) {
	var items []model.Directors
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count directors: %w", err)
	}
	if err := q.OrderExpr("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list directors: %w", err)
	}
	return items, total, nil
}

func (r *metadataRepo) GetDirector(ctx context.Context, id int64) (*model.Directors, error) {
	item := new(model.Directors)
	err := r.db.NewSelect().Model(item).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get director: %w", err)
	}
	return item, nil
}

func (r *metadataRepo) GetDirectorByName(ctx context.Context, name string) (*model.Directors, error) {
	item := new(model.Directors)
	err := r.db.NewSelect().Model(item).Where("name = ?", name).Where("deleted_at IS NULL").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get director by name: %w", err)
	}
	return item, nil
}

func (r *metadataRepo) GetDirectorsByIDs(ctx context.Context, ids []uint64) ([]model.Directors, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.Directors
	err := r.db.NewSelect().Model(&items).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get directors by ids: %w", err)
	}
	return items, nil
}

func (r *metadataRepo) ExistsDirectorName(ctx context.Context, name string, excludeID int64) (bool, error) {
	q := r.db.NewSelect().Model((*model.Directors)(nil)).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check director name: %w", err)
	}
	return exists, nil
}

func (r *metadataRepo) CreateDirector(ctx context.Context, m *model.Directors) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create director: %w", err)
	}
	return nil
}

func (r *metadataRepo) UpdateDirector(ctx context.Context, m *model.Directors) error {
	_, err := r.db.NewUpdate().Model(m).WherePK().Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("update director: %w", err)
	}
	return nil
}

func (r *metadataRepo) SoftDeleteDirector(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Directors)(nil)).
		Set("deleted_at = ?", now).Set("updated_at = ?", now).
		Where("id = ?", id).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete director: %w", err)
	}
	return nil
}

func (r *metadataRepo) CountDirectorRefs(ctx context.Context, id int64) (int, error) {
	// Only count associations to non-soft-deleted videos.
	n, err := r.db.NewSelect().
		TableExpr("video_directors AS vd").
		Join("JOIN videos AS v ON v.id = vd.video_id AND v.deleted_at IS NULL").
		Where("vd.director_id = ?", id).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count director refs: %w", err)
	}
	return n, nil
}

func (r *metadataRepo) ListActors(ctx context.Context, keyword string, offset, limit int) ([]model.Actors, int, error) {
	var items []model.Actors
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count actors: %w", err)
	}
	if err := q.OrderExpr("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list actors: %w", err)
	}
	return items, total, nil
}

func (r *metadataRepo) GetActor(ctx context.Context, id int64) (*model.Actors, error) {
	item := new(model.Actors)
	err := r.db.NewSelect().Model(item).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get actor: %w", err)
	}
	return item, nil
}

func (r *metadataRepo) GetActorByName(ctx context.Context, name string) (*model.Actors, error) {
	item := new(model.Actors)
	err := r.db.NewSelect().Model(item).Where("name = ?", name).Where("deleted_at IS NULL").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get actor by name: %w", err)
	}
	return item, nil
}

func (r *metadataRepo) GetActorsByIDs(ctx context.Context, ids []uint64) ([]model.Actors, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.Actors
	err := r.db.NewSelect().Model(&items).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get actors by ids: %w", err)
	}
	return items, nil
}

func (r *metadataRepo) ExistsActorName(ctx context.Context, name string, excludeID int64) (bool, error) {
	q := r.db.NewSelect().Model((*model.Actors)(nil)).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check actor name: %w", err)
	}
	return exists, nil
}

func (r *metadataRepo) CreateActor(ctx context.Context, m *model.Actors) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create actor: %w", err)
	}
	return nil
}

func (r *metadataRepo) UpdateActor(ctx context.Context, m *model.Actors) error {
	_, err := r.db.NewUpdate().Model(m).WherePK().Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("update actor: %w", err)
	}
	return nil
}

func (r *metadataRepo) SoftDeleteActor(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Actors)(nil)).
		Set("deleted_at = ?", now).Set("updated_at = ?", now).
		Where("id = ?", id).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete actor: %w", err)
	}
	return nil
}

func (r *metadataRepo) CountActorRefs(ctx context.Context, id int64) (int, error) {
	n, err := r.db.NewSelect().
		TableExpr("video_actors AS va").
		Join("JOIN videos AS v ON v.id = va.video_id AND v.deleted_at IS NULL").
		Where("va.actor_id = ?", id).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count actor refs: %w", err)
	}
	return n, nil
}

func (r *metadataRepo) ListTags(ctx context.Context, keyword string, offset, limit int) ([]model.Tags, int, error) {
	var items []model.Tags
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count tags: %w", err)
	}
	if err := q.OrderExpr("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}
	return items, total, nil
}

func (r *metadataRepo) GetTag(ctx context.Context, id int64) (*model.Tags, error) {
	item := new(model.Tags)
	err := r.db.NewSelect().Model(item).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}
	return item, nil
}

func (r *metadataRepo) GetTagByName(ctx context.Context, name string) (*model.Tags, error) {
	item := new(model.Tags)
	err := r.db.NewSelect().Model(item).Where("name = ?", name).Where("deleted_at IS NULL").Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tag by name: %w", err)
	}
	return item, nil
}

func (r *metadataRepo) GetTagsByIDs(ctx context.Context, ids []uint64) ([]model.Tags, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.Tags
	err := r.db.NewSelect().Model(&items).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tags by ids: %w", err)
	}
	return items, nil
}

func (r *metadataRepo) ExistsTagName(ctx context.Context, name string, excludeID int64) (bool, error) {
	q := r.db.NewSelect().Model((*model.Tags)(nil)).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check tag name: %w", err)
	}
	return exists, nil
}

func (r *metadataRepo) CreateTag(ctx context.Context, m *model.Tags) error {
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create tag: %w", err)
	}
	return nil
}

func (r *metadataRepo) UpdateTag(ctx context.Context, m *model.Tags) error {
	_, err := r.db.NewUpdate().Model(m).WherePK().Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}
	return nil
}

func (r *metadataRepo) SoftDeleteTag(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Tags)(nil)).
		Set("deleted_at = ?", now).Set("updated_at = ?", now).
		Where("id = ?", id).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}

func (r *metadataRepo) CountTagRefs(ctx context.Context, id int64) (int, error) {
	n, err := r.db.NewSelect().
		TableExpr("video_tags AS vt").
		Join("JOIN videos AS v ON v.id = vt.video_id AND v.deleted_at IS NULL").
		Where("vt.tag_id = ?", id).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count tag refs: %w", err)
	}
	return n, nil
}

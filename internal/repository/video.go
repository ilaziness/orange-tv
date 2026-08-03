package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

// VideoTagRow is one row of a batch video→tag query.
type VideoTagRow struct {
	VideoID uint32 `bun:"video_id"`
	TagID   uint32 `bun:"tag_id"`
	Name    string `bun:"name"`
}

// VideoListFilter filters video queries.
type VideoListFilter struct {
	Keyword          string
	CategoryID       uint32
	ParentCategoryID uint32
	CategoryIDs      []uint32
	PublishStatus    *uint8
	Year             uint32 // exact year match (admin)
	YearStart        uint32 // year range start (client, inclusive)
	YearEnd          uint32 // year range end (client, inclusive)
	Region           string
	Language         string
	Sort             string
	DirectorID       uint32
	ActorID          uint32
	TagID            uint32
	OnlyOnline       bool
	CreatedAfter     *time.Time // open API: created_at >= cutoff (data_range filter)
	SourceName       string     // open API: play source name exact match
	Offset           int
	Limit            int
}

// VideoRepository provides video and association persistence.
type VideoRepository interface {
	List(ctx context.Context, f VideoListFilter) ([]model.Videos, int, error)
	GetByID(ctx context.Context, id uint32) (*model.Videos, error)
	GetByIDs(ctx context.Context, ids []uint32) ([]model.Videos, error)
	Create(ctx context.Context, v *model.Videos) error
	BatchCreate(ctx context.Context, videos []*model.Videos) error
	Update(ctx context.Context, v *model.Videos) error
	SoftDelete(ctx context.Context, id uint32) error
	ReplaceDirectors(ctx context.Context, videoID uint32, directorIDs []uint32) error
	ReplaceActors(ctx context.Context, videoID uint32, actors []model.VideoActors) error
	ReplaceTags(ctx context.Context, videoID uint32, tagIDs []uint32) error
	ListDirectorIDs(ctx context.Context, videoID uint32) ([]uint32, error)
	ListActorRels(ctx context.Context, videoID uint32) ([]model.VideoActors, error)
	ListTagIDs(ctx context.Context, videoID uint32) ([]uint32, error)
	ListTagsByVideoIDs(ctx context.Context, videoIDs []uint32) ([]VideoTagRow, error)
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error
	WithTx(tx bun.Tx) VideoRepository
	UpdateRatingStats(ctx context.Context, videoID uint32, rating float64, count uint32) error

	// Batch operations (A2)
	BatchUpdatePublishStatus(ctx context.Context, ids []uint32, status uint8) (int, error)
	BatchSoftDelete(ctx context.Context, ids []uint32) (int, error)

	// Stats (A1)
	CountVideos(ctx context.Context) (int, error)
	CountVideosToday(ctx context.Context, since time.Time) (int, error)
	CountVideosByStatus(ctx context.Context, status uint8) (int, error)
	CountCategories(ctx context.Context) (int, error)

	// UpdateCoverDomainByCollectSource batch-replaces the host portion of cover_image
	// for all videos collected by the given collect source. Used when a remote source
	// changes its image CDN domain; collecting one item triggers migration of all
	// historical covers. oldHost/newHost are authority strings (e.g. "img.old.com").
	UpdateCoverDomainByCollectSource(ctx context.Context, collectSourceID uint32, oldHost, newHost string) (int, error)
}

type videoRepo struct {
	db bun.IDB
}

// NewVideoRepo creates a VideoRepository.
func NewVideoRepo(db *database.DB) VideoRepository {
	return &videoRepo{db: db}
}

func (r *videoRepo) WithTx(tx bun.Tx) VideoRepository {
	return &videoRepo{db: tx}
}

func (r *videoRepo) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return r.db.RunInTx(ctx, nil, fn)
}

// UpdateRatingStats updates the rating and rating_count columns for a video.
// Uses explicit Set to persist zero values (Bun's model Update skips zero fields).
func (r *videoRepo) UpdateRatingStats(ctx context.Context, videoID uint32, rating float64, count uint32) error {
	_, err := r.db.NewUpdate().Model((*model.Videos)(nil)).
		Set("rating = ?", rating).
		Set("rating_count = ?", count).
		Where("id = ?", videoID).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update rating stats: %w", err)
	}
	return nil
}

func (r *videoRepo) List(ctx context.Context, f VideoListFilter) ([]model.Videos, int, error) {
	var items []model.Videos
	q := r.db.NewSelect().Model(&items).Where("deleted_at IS NULL")
	if f.OnlyOnline {
		q = q.Where("publish_status = ?", constant.PublishStatusOnline)
	}
	if f.PublishStatus != nil {
		q = q.Where("publish_status = ?", *f.PublishStatus)
	}
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.ParentCategoryID > 0 {
		q = q.Where("parent_category_id = ?", f.ParentCategoryID)
	}
	if len(f.CategoryIDs) > 0 {
		q = q.Where("category_id IN (?)", bun.In(f.CategoryIDs))
	}
	if f.Year > 0 {
		q = q.Where("year = ?", f.Year)
	}
	if f.YearStart > 0 && f.YearEnd > 0 {
		q = q.Where("year BETWEEN ? AND ?", f.YearStart, f.YearEnd)
	} else if f.YearStart > 0 {
		q = q.Where("year >= ?", f.YearStart)
	} else if f.YearEnd > 0 {
		q = q.Where("year <= ?", f.YearEnd)
	}
	if f.Region != "" {
		q = q.Where("(region = ? OR region LIKE ?)", f.Region, "%"+f.Region+"%")
	}
	if f.Language != "" {
		q = q.Where("language = ?", f.Language)
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.WhereOr("title LIKE ?", like).
				WhereOr("subtitle LIKE ?", like).
				WhereOr("description LIKE ?", like)
		})
	}
	if f.DirectorID > 0 {
		q = q.Where("id IN (SELECT video_id FROM video_directors WHERE director_id = ?)", f.DirectorID)
	}
	if f.ActorID > 0 {
		q = q.Where("id IN (SELECT video_id FROM video_actors WHERE actor_id = ?)", f.ActorID)
	}
	if f.TagID > 0 {
		q = q.Where("id IN (SELECT video_id FROM video_tags WHERE tag_id = ?)", f.TagID)
	}
	if f.CreatedAfter != nil {
		q = q.Where("created_at >= ?", *f.CreatedAfter)
	}
	if name := strings.TrimSpace(f.SourceName); name != "" {
		q = q.Where("id IN (SELECT pe.video_id FROM play_episodes pe JOIN play_sources ps ON ps.id = pe.source_id WHERE ps.name = ? AND ps.deleted_at IS NULL)", name)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count videos: %w", err)
	}

	order := videoSortExpr(f.Sort)
	if err := q.OrderExpr(order).Offset(f.Offset).Limit(f.Limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list videos: %w", err)
	}
	return items, total, nil
}

func videoSortExpr(sort string) string {
	switch sort {
	case "year", "year_asc":
		return "year ASC, id DESC"
	case "year_desc":
		return "year DESC, id DESC"
	case "rating", "rating_desc":
		return "rating DESC, id DESC"
	case "rating_asc":
		return "rating ASC, id DESC"
	case "view_count", "view_count_desc":
		return "view_count DESC, id DESC"
	case "created_at_asc":
		return "created_at ASC, id ASC"
	case "id_desc":
		return "id DESC"
	default:
		return "year DESC, id DESC"
	}
}

func (r *videoRepo) GetByID(ctx context.Context, id uint32) (*model.Videos, error) {
	item := new(model.Videos)
	err := r.db.NewSelect().Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get video: %w", err)
	}
	return item, nil
}

func (r *videoRepo) GetByIDs(ctx context.Context, ids []uint32) ([]model.Videos, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []model.Videos
	err := r.db.NewSelect().Model(&items).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Where("publish_status = ?", constant.PublishStatusOnline).
		OrderExpr("id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("get videos by ids: %w", err)
	}
	return items, nil
}

func (r *videoRepo) Create(ctx context.Context, v *model.Videos) error {
	_, err := r.db.NewInsert().Model(v).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create video: %w", err)
	}
	return nil
}

func (r *videoRepo) BatchCreate(ctx context.Context, videos []*model.Videos) error {
	if len(videos) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&videos).Exec(ctx)
	if err != nil {
		return fmt.Errorf("batch create videos: %w", err)
	}
	return nil
}

func (r *videoRepo) Update(ctx context.Context, v *model.Videos) error {
	_, err := r.db.NewUpdate().Model(v).WherePK().Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return fmt.Errorf("update video: %w", err)
	}
	return nil
}

func (r *videoRepo) SoftDelete(ctx context.Context, id uint32) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.Videos)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete video: %w", err)
	}
	return nil
}

func (r *videoRepo) ReplaceDirectors(ctx context.Context, videoID uint32, directorIDs []uint32) error {
	if _, err := r.db.NewDelete().Model((*model.VideoDirectors)(nil)).
		Where("video_id = ?", videoID).Exec(ctx); err != nil {
		return fmt.Errorf("clear video directors: %w", err)
	}
	if len(directorIDs) == 0 {
		return nil
	}
	rows := make([]model.VideoDirectors, 0, len(directorIDs))
	for _, id := range directorIDs {
		rows = append(rows, model.VideoDirectors{VideoID: videoID, DirectorID: id})
	}
	if _, err := r.db.NewInsert().Model(&rows).Exec(ctx); err != nil {
		return fmt.Errorf("insert video directors: %w", err)
	}
	return nil
}

func (r *videoRepo) ReplaceActors(ctx context.Context, videoID uint32, actors []model.VideoActors) error {
	if _, err := r.db.NewDelete().Model((*model.VideoActors)(nil)).
		Where("video_id = ?", videoID).Exec(ctx); err != nil {
		return fmt.Errorf("clear video actors: %w", err)
	}
	if len(actors) == 0 {
		return nil
	}
	for i := range actors {
		actors[i].VideoID = videoID
	}
	if _, err := r.db.NewInsert().Model(&actors).Exec(ctx); err != nil {
		return fmt.Errorf("insert video actors: %w", err)
	}
	return nil
}

func (r *videoRepo) ReplaceTags(ctx context.Context, videoID uint32, tagIDs []uint32) error {
	if _, err := r.db.NewDelete().Model((*model.VideoTags)(nil)).
		Where("video_id = ?", videoID).Exec(ctx); err != nil {
		return fmt.Errorf("clear video tags: %w", err)
	}
	if len(tagIDs) == 0 {
		return nil
	}
	rows := make([]model.VideoTags, 0, len(tagIDs))
	for _, id := range tagIDs {
		rows = append(rows, model.VideoTags{VideoID: videoID, TagID: id})
	}
	if _, err := r.db.NewInsert().Model(&rows).Exec(ctx); err != nil {
		return fmt.Errorf("insert video tags: %w", err)
	}
	return nil
}

func (r *videoRepo) ListDirectorIDs(ctx context.Context, videoID uint32) ([]uint32, error) {
	var rows []model.VideoDirectors
	if err := r.db.NewSelect().Model(&rows).Where("video_id = ?", videoID).Scan(ctx); err != nil {
		return nil, fmt.Errorf("list video directors: %w", err)
	}
	ids := make([]uint32, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.DirectorID)
	}
	return ids, nil
}

func (r *videoRepo) ListActorRels(ctx context.Context, videoID uint32) ([]model.VideoActors, error) {
	var rows []model.VideoActors
	if err := r.db.NewSelect().Model(&rows).Where("video_id = ?", videoID).Scan(ctx); err != nil {
		return nil, fmt.Errorf("list video actors: %w", err)
	}
	return rows, nil
}

func (r *videoRepo) ListTagIDs(ctx context.Context, videoID uint32) ([]uint32, error) {
	var rows []model.VideoTags
	if err := r.db.NewSelect().Model(&rows).Where("video_id = ?", videoID).Scan(ctx); err != nil {
		return nil, fmt.Errorf("list video tags: %w", err)
	}
	ids := make([]uint32, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.TagID)
	}
	return ids, nil
}

func (r *videoRepo) ListTagsByVideoIDs(ctx context.Context, videoIDs []uint32) ([]VideoTagRow, error) {
	if len(videoIDs) == 0 {
		return nil, nil
	}
	var rows []VideoTagRow
	err := r.db.NewSelect().
		Model((*model.VideoTags)(nil)).
		ColumnExpr("vt.video_id, vt.tag_id, ta.name").
		Join("JOIN tags AS ta ON ta.id = vt.tag_id AND ta.deleted_at IS NULL").
		Where("vt.video_id IN (?)", bun.In(videoIDs)).
		OrderExpr("vt.id ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("list tags by video ids: %w", err)
	}
	return rows, nil
}

// ===== Batch operations (A2) =====

func (r *videoRepo) BatchUpdatePublishStatus(ctx context.Context, ids []uint32, status uint8) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.Videos)(nil)).
		Set("publish_status = ?", status).
		Set("updated_at = ?", now).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("batch update publish status: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *videoRepo) BatchSoftDelete(ctx context.Context, ids []uint32) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.Videos)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("batch delete videos: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// ===== Stats (A1) =====

func (r *videoRepo) CountVideos(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().Model((*model.Videos)(nil)).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count videos: %w", err)
	}
	return count, nil
}

func (r *videoRepo) CountVideosToday(ctx context.Context, since time.Time) (int, error) {
	count, err := r.db.NewSelect().Model((*model.Videos)(nil)).
		Where("deleted_at IS NULL").
		Where("created_at >= ?", since).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count videos today: %w", err)
	}
	return count, nil
}

func (r *videoRepo) CountVideosByStatus(ctx context.Context, status uint8) (int, error) {
	count, err := r.db.NewSelect().Model((*model.Videos)(nil)).
		Where("deleted_at IS NULL").
		Where("publish_status = ?", status).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count videos by status: %w", err)
	}
	return count, nil
}

func (r *videoRepo) CountCategories(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().Model((*model.Categories)(nil)).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return count, nil
}

// UpdateCoverDomainByCollectSource batch-replaces the host portion of cover_image
// for all videos of a collect source. The match uses "://oldHost" prefix to ensure
// only the URL authority is replaced, not coincidental substrings in the path.
func (r *videoRepo) UpdateCoverDomainByCollectSource(ctx context.Context, collectSourceID uint32, oldHost, newHost string) (int, error) {
	if collectSourceID == 0 || oldHost == "" || newHost == "" || oldHost == newHost {
		return 0, nil
	}
	oldSeg := "://" + oldHost
	newSeg := "://" + newHost
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.Videos)(nil)).
		Set("cover_image = REPLACE(cover_image, ?, ?)", oldSeg, newSeg).
		Set("updated_at = ?", now).
		Where("collect_source_id = ?", collectSourceID).
		Where("cover_image LIKE ?", "%"+oldSeg+"%").
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("update cover domain by collect source: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

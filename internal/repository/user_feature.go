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

// UserFeatureRepository manages user favorites, play history, comments, banners and site stats.
type UserFeatureRepository interface {
	// Favorites
	ListFavorites(ctx context.Context, userID int64, offset, limit int) ([]model.UserFavorites, int, error)
	GetFavorite(ctx context.Context, userID, videoID int64) (*model.UserFavorites, error)
	AddFavorite(ctx context.Context, f *model.UserFavorites) error
	RemoveFavorite(ctx context.Context, userID, videoID int64) error

	// Play history
	ListHistory(ctx context.Context, userID int64, offset, limit int) ([]model.UserPlayHistory, int, error)
	GetHistory(ctx context.Context, userID, videoID int64) (*model.UserPlayHistory, error)
	UpsertHistory(ctx context.Context, h *model.UserPlayHistory) error
	DeleteHistory(ctx context.Context, userID, videoID int64) error
	ClearHistory(ctx context.Context, userID int64) error

	// Comments
	ListComments(ctx context.Context, videoID int64, offset, limit int) ([]model.VideoComments, int, error)
	ListCommentsByUser(ctx context.Context, userID int64, offset, limit int) ([]model.VideoComments, int, error)
	GetComment(ctx context.Context, id int64) (*model.VideoComments, error)
	CreateComment(ctx context.Context, c *model.VideoComments) error
	UpdateCommentStatus(ctx context.Context, id int64, status int8) error
	DeleteComment(ctx context.Context, id int64) error

	// User login logs
	CreateUserLoginLog(ctx context.Context, l *model.UserLoginLogs) error
	ListUserLoginLogs(ctx context.Context, userID int64, offset, limit int) ([]model.UserLoginLogs, int, error)

	// Banners
	ListBanners(ctx context.Context, status *uint8) ([]model.Banners, error)
	ListAllBanners(ctx context.Context, offset, limit int) ([]model.Banners, int, error)
	GetBanner(ctx context.Context, id int64) (*model.Banners, error)
	CreateBanner(ctx context.Context, b *model.Banners) error
	UpdateBanner(ctx context.Context, b *model.Banners) error
	DeleteBanner(ctx context.Context, id int64) error

	// Site stats
	IncrDailyStats(ctx context.Context, date time.Time, pv, uv int) error
	GetDailyStats(ctx context.Context, date time.Time) (*model.SiteStatsDaily, error)
	UpsertOnlineSession(ctx context.Context, key, ip string) error
	CountOnlineSessions(ctx context.Context, since time.Time) (int, error)
	CleanupOnlineSessions(ctx context.Context, before time.Time) error
}

type userFeatureRepo struct {
	db *database.DB
}

// NewUserFeatureRepo creates a UserFeatureRepository.
func NewUserFeatureRepo(db *database.DB) UserFeatureRepository {
	return &userFeatureRepo{db: db}
}

func notFoundOrErr(err error, wrap string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", wrap, err)
	}
	return nil
}

// ===== Favorites =====

func (r *userFeatureRepo) ListFavorites(ctx context.Context, userID int64, offset, limit int) ([]model.UserFavorites, int, error) {
	items := make([]model.UserFavorites, 0, limit)
	q := r.db.NewSelect().Model(&items).Where("user_id = ?", userID)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	return items, total, nil
}

func (r *userFeatureRepo) GetFavorite(ctx context.Context, userID, videoID int64) (*model.UserFavorites, error) {
	f := new(model.UserFavorites)
	err := r.db.NewSelect().Model(f).
		Where("user_id = ?", userID).
		Where("video_id = ?", videoID).
		Scan(ctx)
	if err := notFoundOrErr(err, "get favorite"); err != nil {
		return nil, err
	}
	return f, nil
}

func (r *userFeatureRepo) AddFavorite(ctx context.Context, f *model.UserFavorites) error {
	_, err := r.db.NewInsert().Model(f).Exec(ctx)
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) RemoveFavorite(ctx context.Context, userID, videoID int64) error {
	_, err := r.db.NewDelete().Model((*model.UserFavorites)(nil)).
		Where("user_id = ?", userID).
		Where("video_id = ?", videoID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

// ===== Play history =====

func (r *userFeatureRepo) ListHistory(ctx context.Context, userID int64, offset, limit int) ([]model.UserPlayHistory, int, error) {
	items := make([]model.UserPlayHistory, 0, limit)
	q := r.db.NewSelect().Model(&items).Where("user_id = ?", userID)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count history: %w", err)
	}
	if err := q.Order("last_played_at DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list history: %w", err)
	}
	return items, total, nil
}

func (r *userFeatureRepo) GetHistory(ctx context.Context, userID, videoID int64) (*model.UserPlayHistory, error) {
	h := new(model.UserPlayHistory)
	err := r.db.NewSelect().Model(h).
		Where("user_id = ?", userID).
		Where("video_id = ?", videoID).
		Scan(ctx)
	if err := notFoundOrErr(err, "get history"); err != nil {
		return nil, err
	}
	return h, nil
}

func (r *userFeatureRepo) UpsertHistory(ctx context.Context, h *model.UserPlayHistory) error {
	_, err := r.db.NewInsert().Model(h).
		On("DUPLICATE KEY UPDATE").
		Set("play_source_id = VALUES(play_source_id)").
		Set("episode_id = VALUES(episode_id)").
		Set("progress = VALUES(progress)").
		Set("duration = VALUES(duration)").
		Set("last_played_at = VALUES(last_played_at)").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("upsert history: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) DeleteHistory(ctx context.Context, userID, videoID int64) error {
	_, err := r.db.NewDelete().Model((*model.UserPlayHistory)(nil)).
		Where("user_id = ?", userID).
		Where("video_id = ?", videoID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete history: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) ClearHistory(ctx context.Context, userID int64) error {
	_, err := r.db.NewDelete().Model((*model.UserPlayHistory)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("clear history: %w", err)
	}
	return nil
}

// ===== Comments =====

func (r *userFeatureRepo) ListComments(ctx context.Context, videoID int64, offset, limit int) ([]model.VideoComments, int, error) {
	items := make([]model.VideoComments, 0, limit)
	q := r.db.NewSelect().Model(&items).
		Relation("User").
		Where("vc.video_id = ?", videoID).
		Where("vc.status = ?", 1)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}
	if err := q.Order("vc.id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	return items, total, nil
}

func (r *userFeatureRepo) ListCommentsByUser(ctx context.Context, userID int64, offset, limit int) ([]model.VideoComments, int, error) {
	items := make([]model.VideoComments, 0, limit)
	q := r.db.NewSelect().Model(&items).Where("user_id = ?", userID)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count user comments: %w", err)
	}
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list user comments: %w", err)
	}
	return items, total, nil
}

func (r *userFeatureRepo) GetComment(ctx context.Context, id int64) (*model.VideoComments, error) {
	c := new(model.VideoComments)
	err := r.db.NewSelect().Model(c).Where("id = ?", id).Scan(ctx)
	if err := notFoundOrErr(err, "get comment"); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *userFeatureRepo) CreateComment(ctx context.Context, c *model.VideoComments) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) UpdateCommentStatus(ctx context.Context, id int64, status int8) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.VideoComments)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update comment status: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) DeleteComment(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*model.VideoComments)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}

// ===== User login logs =====

func (r *userFeatureRepo) CreateUserLoginLog(ctx context.Context, l *model.UserLoginLogs) error {
	_, err := r.db.NewInsert().Model(l).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create user login log: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) ListUserLoginLogs(ctx context.Context, userID int64, offset, limit int) ([]model.UserLoginLogs, int, error) {
	items := make([]model.UserLoginLogs, 0, limit)
	q := r.db.NewSelect().Model(&items).Where("user_id = ?", userID)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count user login logs: %w", err)
	}
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list user login logs: %w", err)
	}
	return items, total, nil
}

// ===== Banners =====

func (r *userFeatureRepo) ListBanners(ctx context.Context, status *uint8) ([]model.Banners, error) {
	items := make([]model.Banners, 0, 20)
	q := r.db.NewSelect().Model(&items)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Order("sort ASC, id DESC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("list banners: %w", err)
	}
	return items, nil
}

func (r *userFeatureRepo) ListAllBanners(ctx context.Context, offset, limit int) ([]model.Banners, int, error) {
	items := make([]model.Banners, 0, limit)
	q := r.db.NewSelect().Model(&items)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count banners: %w", err)
	}
	if err := q.Order("sort ASC, id DESC").Offset(offset).Limit(limit).Scan(ctx); err != nil {
		return nil, 0, fmt.Errorf("list banners: %w", err)
	}
	return items, total, nil
}

func (r *userFeatureRepo) GetBanner(ctx context.Context, id int64) (*model.Banners, error) {
	b := new(model.Banners)
	err := r.db.NewSelect().Model(b).Where("id = ?", id).Scan(ctx)
	if err := notFoundOrErr(err, "get banner"); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *userFeatureRepo) CreateBanner(ctx context.Context, b *model.Banners) error {
	_, err := r.db.NewInsert().Model(b).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create banner: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) UpdateBanner(ctx context.Context, b *model.Banners) error {
	_, err := r.db.NewUpdate().Model(b).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("update banner: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) DeleteBanner(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*model.Banners)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete banner: %w", err)
	}
	return nil
}

// ===== Site stats =====

func (r *userFeatureRepo) IncrDailyStats(ctx context.Context, date time.Time, pv, uv int) error {
	// Normalize to midnight in the same timezone as the caller's date.
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	_, err := r.db.NewInsert().Model(&model.SiteStatsDaily{
		StatDate: d,
		PV:       uint64(pv),
		UV:       uint64(uv),
	}).
		On("DUPLICATE KEY UPDATE").
		Set("pv = pv + VALUES(pv)").
		Set("uv = uv + VALUES(uv)").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("incr daily stats: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) GetDailyStats(ctx context.Context, date time.Time) (*model.SiteStatsDaily, error) {
	// Normalize to midnight in the same timezone as the caller's date.
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	s := new(model.SiteStatsDaily)
	err := r.db.NewSelect().Model(s).Where("stat_date = ?", d).Scan(ctx)
	if err := notFoundOrErr(err, "get daily stats"); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *userFeatureRepo) UpsertOnlineSession(ctx context.Context, key, ip string) error {
	now := time.Now()
	_, err := r.db.NewInsert().Model(&model.OnlineSessions{
		SessionKey:   key,
		IP:           ip,
		LastActiveAt: now,
	}).
		On("DUPLICATE KEY UPDATE").
		Set("ip = VALUES(ip)").
		Set("last_active_at = VALUES(last_active_at)").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("upsert online session: %w", err)
	}
	return nil
}

func (r *userFeatureRepo) CountOnlineSessions(ctx context.Context, since time.Time) (int, error) {
	count, err := r.db.NewSelect().Model((*model.OnlineSessions)(nil)).
		Where("last_active_at >= ?", since).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count online sessions: %w", err)
	}
	return count, nil
}

func (r *userFeatureRepo) CleanupOnlineSessions(ctx context.Context, before time.Time) error {
	_, err := r.db.NewDelete().Model((*model.OnlineSessions)(nil)).
		Where("last_active_at < ?", before).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("cleanup online sessions: %w", err)
	}
	return nil
}

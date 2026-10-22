package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/uptrace/bun"
)

// LiveCategoryListFilter filters live categories.
type LiveCategoryListFilter struct {
	Status   *uint8
	ParentID *uint32
	Keyword  string
	Offset   int
	Limit    int
}

// LiveApplyListFilter filters streamer applies.
type LiveApplyListFilter struct {
	Status  *uint8
	UserID  *uint32
	Keyword string
	Offset  int
	Limit   int
}

// LiveStreamerListFilter filters streamers (users.is_streamer=1).
type LiveStreamerListFilter struct {
	Keyword    string
	RoomStatus *uint8
	Offset     int
	Limit      int
}

// LiveRoomListFilter filters live rooms.
type LiveRoomListFilter struct {
	Status     *uint8
	CategoryID *uint32
	UserID     *uint32
	Keyword    string
	LivingOnly bool
	Offset     int
	Limit      int
}

// LiveSessionListFilter filters live sessions.
type LiveSessionListFilter struct {
	RoomID *uint32
	UserID *uint32
	Offset int
	Limit  int
}

// LiveCategoryRow is a category joined with optional parent name.
type LiveCategoryRow struct {
	model.LiveCategories
	ParentName string `bun:"parent_name"`
}

// LiveStreamerRow is a streamer joined with optional room/category.
type LiveStreamerRow struct {
	UserID          uint32     `bun:"user_id"`
	UserStrID       string     `bun:"user_str_id"`
	Nickname        string     `bun:"nickname"`
	Email           string     `bun:"email"`
	Avatar          string     `bun:"avatar"`
	StreamerAt      *time.Time `bun:"streamer_at"`
	RoomID          uint32     `bun:"room_id"`
	RoomTitle       string     `bun:"room_title"`
	RoomStatus      uint8      `bun:"room_status"`
	CategoryID      uint32     `bun:"category_id"`
	CategoryName    string     `bun:"category_name"`
	SubCategoryID   uint32     `bun:"sub_category_id"`
	SubCategoryName string     `bun:"sub_category_name"`
}

// LiveRoomRow is a room joined with user/category.
type LiveRoomRow struct {
	model.LiveRooms
	UserStrID       string `bun:"user_str_id"`
	Nickname        string `bun:"nickname"`
	Avatar          string `bun:"avatar"`
	CategoryName    string `bun:"category_name"`
	SubCategoryName string `bun:"sub_category_name"`
}

// LiveApplyRow is an apply joined with user.
type LiveApplyRow struct {
	model.StreamerApplies
	UserStrID string `bun:"user_str_id"`
	Nickname  string `bun:"nickname"`
	Email     string `bun:"email"`
	Avatar    string `bun:"avatar"`
}

// LiveSessionRow is a session joined with category names.
type LiveSessionRow struct {
	model.LiveSessions
	CategoryName    string `bun:"category_name"`
	SubCategoryName string `bun:"sub_category_name"`
}

// LiveRepository manages user-live tables.
type LiveRepository interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error

	// Categories
	ListCategories(ctx context.Context, f LiveCategoryListFilter) ([]LiveCategoryRow, int, error)
	ListEnabledCategories(ctx context.Context) ([]model.LiveCategories, error)
	GetCategory(ctx context.Context, id uint32) (*model.LiveCategories, error)
	ExistsCategoryName(ctx context.Context, name string, parentID, excludeID uint32) (bool, error)
	CountCategoryChildren(ctx context.Context, parentID uint32) (int, error)
	CountEnabledCategoryChildren(ctx context.Context, parentID uint32) (int, error)
	CountRoomsUsingCategory(ctx context.Context, categoryID uint32) (int, error)
	CreateCategory(ctx context.Context, c *model.LiveCategories) error
	UpdateCategory(ctx context.Context, c *model.LiveCategories) error
	SoftDeleteCategory(ctx context.Context, id uint32) error

	// Applies
	ListApplies(ctx context.Context, f LiveApplyListFilter) ([]LiveApplyRow, int, error)
	GetApply(ctx context.Context, id uint32) (*model.StreamerApplies, error)
	GetLatestApplyByUser(ctx context.Context, userID uint32) (*model.StreamerApplies, error)
	HasPendingApply(ctx context.Context, userID uint32) (bool, error)
	CreateApply(ctx context.Context, a *model.StreamerApplies) error
	ApproveApply(ctx context.Context, id, adminID uint32, remark string, reviewedAt time.Time) (bool, error)
	RejectApply(ctx context.Context, id, adminID uint32, remark string, reviewedAt time.Time) (bool, error)

	// Users / streamers
	GetUser(ctx context.Context, id uint32) (*model.Users, error)
	ListStreamers(ctx context.Context, f LiveStreamerListFilter) ([]LiveStreamerRow, int, error)
	MarkStreamer(ctx context.Context, userID uint32, at time.Time) error
	RevokeStreamer(ctx context.Context, userID uint32) (bool, error)

	// Rooms
	ListRooms(ctx context.Context, f LiveRoomListFilter) ([]LiveRoomRow, int, error)
	GetRoom(ctx context.Context, id uint32) (*model.LiveRooms, error)
	GetRoomByUser(ctx context.Context, userID uint32) (*model.LiveRooms, error)
	GetRoomByStreamKey(ctx context.Context, streamKey string) (*model.LiveRooms, error)
	CountLivingRooms(ctx context.Context) (int, error)
	GetRoomRow(ctx context.Context, id uint32) (*LiveRoomRow, error)
	CreateRoom(ctx context.Context, room *model.LiveRooms) error
	UpdateRoom(ctx context.Context, room *model.LiveRooms) error
	UpdateRoomStreamKey(ctx context.Context, roomID uint32, streamKey string) (bool, error)
	ConditionalStartRoom(ctx context.Context, roomID, sessionID uint32, fromStatus uint8, startedAt time.Time) (bool, error)
	ConditionalStopRoom(ctx context.Context, roomID uint32, toStatus uint8, endedAt time.Time) (bool, error)
	SetRoomBanned(ctx context.Context, roomID uint32) error
	SetRoomIdle(ctx context.Context, roomID uint32) (bool, error)

	// Sessions
	ListSessions(ctx context.Context, f LiveSessionListFilter) ([]LiveSessionRow, int, error)
	CreateSession(ctx context.Context, s *model.LiveSessions) error
	EndSession(ctx context.Context, sessionID uint32, endedAt time.Time, endReason string) (bool, error)
}

type liveRepo struct {
	db bun.IDB
}

// NewLiveRepo creates a LiveRepository.
func NewLiveRepo(db *database.DB) LiveRepository {
	return &liveRepo{db: db}
}

// NewLiveRepoFromIDB creates a LiveRepository bound to an existing bun.IDB (e.g. transaction).
func NewLiveRepoFromIDB(db bun.IDB) LiveRepository {
	return &liveRepo{db: db}
}

func (r *liveRepo) RunInTx(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return r.db.RunInTx(ctx, nil, fn)
}

// ===== Categories =====

func (r *liveRepo) ListCategories(ctx context.Context, f LiveCategoryListFilter) ([]LiveCategoryRow, int, error) {
	items := make([]LiveCategoryRow, 0)
	q := r.db.NewSelect().
		TableExpr("live_categories AS lc").
		ColumnExpr("lc.*").
		ColumnExpr("COALESCE(p.name, '') AS parent_name").
		Join("LEFT JOIN live_categories AS p ON p.id = lc.parent_id AND p.deleted_at IS NULL").
		Where("lc.deleted_at IS NULL")
	if f.Status != nil {
		q = q.Where("lc.status = ?", *f.Status)
	}
	if f.ParentID != nil {
		q = q.Where("lc.parent_id = ?", *f.ParentID)
	}
	if f.Keyword != "" {
		q = q.Where("lc.name LIKE ?", "%"+f.Keyword+"%")
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count live categories: %w", err)
	}
	if err := q.OrderExpr("lc.sort ASC, lc.id ASC").Offset(f.Offset).Limit(f.Limit).Scan(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("list live categories: %w", err)
	}
	return items, total, nil
}

func (r *liveRepo) ListEnabledCategories(ctx context.Context) ([]model.LiveCategories, error) {
	items := make([]model.LiveCategories, 0)
	err := r.db.NewSelect().Model(&items).
		Where("deleted_at IS NULL").
		Where("status = ?", constant.StatusEnabled).
		Order("sort ASC, id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled live categories: %w", err)
	}
	return items, nil
}

func (r *liveRepo) GetCategory(ctx context.Context, id uint32) (*model.LiveCategories, error) {
	c := new(model.LiveCategories)
	err := r.db.NewSelect().Model(c).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live category: %w", err)
	}
	return c, nil
}

func (r *liveRepo) ExistsCategoryName(ctx context.Context, name string, parentID, excludeID uint32) (bool, error) {
	q := r.db.NewSelect().Model((*model.LiveCategories)(nil)).
		Where("name = ?", name).
		Where("parent_id = ?", parentID).
		Where("deleted_at IS NULL")
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check live category name: %w", err)
	}
	return exists, nil
}

func (r *liveRepo) CountCategoryChildren(ctx context.Context, parentID uint32) (int, error) {
	n, err := r.db.NewSelect().Model((*model.LiveCategories)(nil)).
		Where("parent_id = ?", parentID).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count live category children: %w", err)
	}
	return n, nil
}

func (r *liveRepo) CountEnabledCategoryChildren(ctx context.Context, parentID uint32) (int, error) {
	n, err := r.db.NewSelect().Model((*model.LiveCategories)(nil)).
		Where("parent_id = ?", parentID).
		Where("status = ?", constant.StatusEnabled).
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count enabled live category children: %w", err)
	}
	return n, nil
}

func (r *liveRepo) CountRoomsUsingCategory(ctx context.Context, categoryID uint32) (int, error) {
	n, err := r.db.NewSelect().Model((*model.LiveRooms)(nil)).
		Where("category_id = ? OR sub_category_id = ?", categoryID, categoryID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count rooms using live category: %w", err)
	}
	return n, nil
}

func (r *liveRepo) CreateCategory(ctx context.Context, c *model.LiveCategories) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create live category: %w", err)
	}
	return nil
}

func (r *liveRepo) UpdateCategory(ctx context.Context, c *model.LiveCategories) error {
	_, err := r.db.NewUpdate().Model(c).
		WherePK().
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update live category: %w", err)
	}
	return nil
}

func (r *liveRepo) SoftDeleteCategory(ctx context.Context, id uint32) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.LiveCategories)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("soft delete live category: %w", err)
	}
	return nil
}

// ===== Applies =====

func (r *liveRepo) ListApplies(ctx context.Context, f LiveApplyListFilter) ([]LiveApplyRow, int, error) {
	items := make([]LiveApplyRow, 0)
	q := r.db.NewSelect().
		TableExpr("streamer_applies AS sa").
		ColumnExpr("sa.*").
		ColumnExpr("us.str_id AS user_str_id").
		ColumnExpr("us.nickname AS nickname").
		ColumnExpr("us.email AS email").
		ColumnExpr("us.avatar AS avatar").
		Join("LEFT JOIN users AS us ON us.id = sa.user_id AND us.deleted_at IS NULL")
	if f.Status != nil {
		q = q.Where("sa.status = ?", *f.Status)
	}
	if f.UserID != nil {
		q = q.Where("sa.user_id = ?", *f.UserID)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("(us.nickname LIKE ? OR us.email LIKE ? OR us.str_id LIKE ?)", like, like, like)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count streamer applies: %w", err)
	}
	if err := q.OrderExpr("sa.id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("list streamer applies: %w", err)
	}
	return items, total, nil
}

func (r *liveRepo) GetApply(ctx context.Context, id uint32) (*model.StreamerApplies, error) {
	a := new(model.StreamerApplies)
	err := r.db.NewSelect().Model(a).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get streamer apply: %w", err)
	}
	return a, nil
}

func (r *liveRepo) GetLatestApplyByUser(ctx context.Context, userID uint32) (*model.StreamerApplies, error) {
	a := new(model.StreamerApplies)
	err := r.db.NewSelect().Model(a).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest streamer apply: %w", err)
	}
	return a, nil
}

func (r *liveRepo) HasPendingApply(ctx context.Context, userID uint32) (bool, error) {
	exists, err := r.db.NewSelect().Model((*model.StreamerApplies)(nil)).
		Where("user_id = ?", userID).
		Where("status = ?", constant.StreamerApplyPending).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("check pending streamer apply: %w", err)
	}
	return exists, nil
}

func (r *liveRepo) CreateApply(ctx context.Context, a *model.StreamerApplies) error {
	_, err := r.db.NewInsert().Model(a).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create streamer apply: %w", err)
	}
	return nil
}

func (r *liveRepo) ApproveApply(ctx context.Context, id, adminID uint32, remark string, reviewedAt time.Time) (bool, error) {
	res, err := r.db.NewUpdate().Model((*model.StreamerApplies)(nil)).
		Set("status = ?", constant.StreamerApplyApproved).
		Set("review_admin_id = ?", adminID).
		Set("review_remark = ?", remark).
		Set("reviewed_at = ?", reviewedAt).
		Set("updated_at = ?", reviewedAt).
		Where("id = ?", id).
		Where("status = ?", constant.StreamerApplyPending).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("approve streamer apply: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *liveRepo) RejectApply(ctx context.Context, id, adminID uint32, remark string, reviewedAt time.Time) (bool, error) {
	res, err := r.db.NewUpdate().Model((*model.StreamerApplies)(nil)).
		Set("status = ?", constant.StreamerApplyRejected).
		Set("review_admin_id = ?", adminID).
		Set("review_remark = ?", remark).
		Set("reviewed_at = ?", reviewedAt).
		Set("updated_at = ?", reviewedAt).
		Where("id = ?", id).
		Where("status = ?", constant.StreamerApplyPending).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("reject streamer apply: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ===== Users / streamers =====

func (r *liveRepo) GetUser(ctx context.Context, id uint32) (*model.Users, error) {
	u := new(model.Users)
	err := r.db.NewSelect().Model(u).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live user: %w", err)
	}
	return u, nil
}

func (r *liveRepo) ListStreamers(ctx context.Context, f LiveStreamerListFilter) ([]LiveStreamerRow, int, error) {
	items := make([]LiveStreamerRow, 0)
	q := r.db.NewSelect().
		TableExpr("users AS us").
		ColumnExpr("us.id AS user_id").
		ColumnExpr("us.str_id AS user_str_id").
		ColumnExpr("us.nickname AS nickname").
		ColumnExpr("us.email AS email").
		ColumnExpr("us.avatar AS avatar").
		ColumnExpr("us.streamer_at AS streamer_at").
		ColumnExpr("COALESCE(lr.id, 0) AS room_id").
		ColumnExpr("COALESCE(lr.title, '') AS room_title").
		ColumnExpr("COALESCE(lr.status, 0) AS room_status").
		ColumnExpr("COALESCE(lr.category_id, 0) AS category_id").
		ColumnExpr("COALESCE(lc.name, '') AS category_name").
		ColumnExpr("COALESCE(lr.sub_category_id, 0) AS sub_category_id").
		ColumnExpr("COALESCE(lsc.name, '') AS sub_category_name").
		Join("LEFT JOIN live_rooms AS lr ON lr.user_id = us.id").
		Join("LEFT JOIN live_categories AS lc ON lc.id = lr.category_id AND lc.deleted_at IS NULL").
		Join("LEFT JOIN live_categories AS lsc ON lsc.id = lr.sub_category_id AND lsc.deleted_at IS NULL").
		Where("us.is_streamer = ?", 1).
		Where("us.deleted_at IS NULL")
	if f.RoomStatus != nil {
		q = q.Where("lr.status = ?", *f.RoomStatus)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("(us.nickname LIKE ? OR us.email LIKE ? OR us.str_id LIKE ?)", like, like, like)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count streamers: %w", err)
	}
	if err := q.OrderExpr("us.streamer_at DESC, us.id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("list streamers: %w", err)
	}
	return items, total, nil
}

func (r *liveRepo) MarkStreamer(ctx context.Context, userID uint32, at time.Time) error {
	_, err := r.db.NewUpdate().Model((*model.Users)(nil)).
		Set("is_streamer = ?", 1).
		Set("streamer_at = ?", at).
		Set("updated_at = ?", at).
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("mark streamer: %w", err)
	}
	return nil
}

func (r *liveRepo) RevokeStreamer(ctx context.Context, userID uint32) (bool, error) {
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.Users)(nil)).
		Set("is_streamer = ?", 0).
		Set("streamer_at = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", userID).
		Where("is_streamer = ?", 1).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("revoke streamer: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ===== Rooms =====

func (r *liveRepo) ListRooms(ctx context.Context, f LiveRoomListFilter) ([]LiveRoomRow, int, error) {
	items := make([]LiveRoomRow, 0)
	q := r.db.NewSelect().
		TableExpr("live_rooms AS lr").
		ColumnExpr("lr.*").
		ColumnExpr("COALESCE(us.str_id, '') AS user_str_id").
		ColumnExpr("COALESCE(us.nickname, '') AS nickname").
		ColumnExpr("COALESCE(us.avatar, '') AS avatar").
		ColumnExpr("COALESCE(lc.name, '') AS category_name").
		ColumnExpr("COALESCE(lsc.name, '') AS sub_category_name").
		Join("LEFT JOIN users AS us ON us.id = lr.user_id AND us.deleted_at IS NULL").
		Join("LEFT JOIN live_categories AS lc ON lc.id = lr.category_id AND lc.deleted_at IS NULL").
		Join("LEFT JOIN live_categories AS lsc ON lsc.id = lr.sub_category_id AND lsc.deleted_at IS NULL")
	if f.LivingOnly {
		q = q.Where("lr.status = ?", constant.LiveRoomLiving)
	}
	if f.Status != nil {
		q = q.Where("lr.status = ?", *f.Status)
	}
	if f.CategoryID != nil {
		cat, err := r.GetCategory(ctx, *f.CategoryID)
		if err != nil {
			return nil, 0, err
		}
		if cat == nil {
			return items, 0, nil
		}
		if cat.ParentID == 0 {
			q = q.Where("lr.category_id = ?", cat.ID)
		} else {
			q = q.Where("lr.sub_category_id = ?", cat.ID).Where("lr.category_id = ?", cat.ParentID)
		}
	}
	if f.UserID != nil {
		q = q.Where("lr.user_id = ?", *f.UserID)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("(lr.title LIKE ? OR us.nickname LIKE ? OR us.str_id LIKE ?)", like, like, like)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count live rooms: %w", err)
	}
	order := "lr.id DESC"
	if f.LivingOnly {
		order = "lr.started_at DESC, lr.id DESC"
	}
	if err := q.OrderExpr(order).Offset(f.Offset).Limit(f.Limit).Scan(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("list live rooms: %w", err)
	}
	return items, total, nil
}

func (r *liveRepo) GetRoom(ctx context.Context, id uint32) (*model.LiveRooms, error) {
	room := new(model.LiveRooms)
	err := r.db.NewSelect().Model(room).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live room: %w", err)
	}
	return room, nil
}

func (r *liveRepo) GetRoomByStreamKey(ctx context.Context, streamKey string) (*model.LiveRooms, error) {
	room := new(model.LiveRooms)
	err := r.db.NewSelect().Model(room).Where("stream_key = ?", streamKey).Limit(1).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live room by stream key: %w", err)
	}
	return room, nil
}

func (r *liveRepo) CountLivingRooms(ctx context.Context) (int, error) {
	n, err := r.db.NewSelect().Model((*model.LiveRooms)(nil)).
		Where("status = ?", constant.LiveRoomLiving).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count living rooms: %w", err)
	}
	return n, nil
}

func (r *liveRepo) GetRoomByUser(ctx context.Context, userID uint32) (*model.LiveRooms, error) {
	room := new(model.LiveRooms)
	err := r.db.NewSelect().Model(room).Where("user_id = ?", userID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live room by user: %w", err)
	}
	return room, nil
}

func (r *liveRepo) GetRoomRow(ctx context.Context, id uint32) (*LiveRoomRow, error) {
	row := new(LiveRoomRow)
	err := r.db.NewSelect().
		TableExpr("live_rooms AS lr").
		ColumnExpr("lr.*").
		ColumnExpr("COALESCE(us.str_id, '') AS user_str_id").
		ColumnExpr("COALESCE(us.nickname, '') AS nickname").
		ColumnExpr("COALESCE(us.avatar, '') AS avatar").
		ColumnExpr("COALESCE(lc.name, '') AS category_name").
		ColumnExpr("COALESCE(lsc.name, '') AS sub_category_name").
		Join("LEFT JOIN users AS us ON us.id = lr.user_id AND us.deleted_at IS NULL").
		Join("LEFT JOIN live_categories AS lc ON lc.id = lr.category_id AND lc.deleted_at IS NULL").
		Join("LEFT JOIN live_categories AS lsc ON lsc.id = lr.sub_category_id AND lsc.deleted_at IS NULL").
		Where("lr.id = ?", id).
		Scan(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get live room row: %w", err)
	}
	return row, nil
}

func (r *liveRepo) CreateRoom(ctx context.Context, room *model.LiveRooms) error {
	_, err := r.db.NewInsert().Model(room).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create live room: %w", err)
	}
	return nil
}

func (r *liveRepo) UpdateRoom(ctx context.Context, room *model.LiveRooms) error {
	_, err := r.db.NewUpdate().Model(room).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("update live room: %w", err)
	}
	return nil
}

func (r *liveRepo) UpdateRoomStreamKey(ctx context.Context, roomID uint32, streamKey string) (bool, error) {
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.LiveRooms)(nil)).
		Set("stream_key = ?", streamKey).
		Set("updated_at = ?", now).
		Where("id = ?", roomID).
		Where("status <> ?", constant.LiveRoomLiving).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("update live room stream key: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *liveRepo) ConditionalStartRoom(ctx context.Context, roomID, sessionID uint32, fromStatus uint8, startedAt time.Time) (bool, error) {
	res, err := r.db.NewUpdate().Model((*model.LiveRooms)(nil)).
		Set("status = ?", constant.LiveRoomLiving).
		Set("current_session_id = ?", sessionID).
		Set("started_at = ?", startedAt).
		Set("ended_at = NULL").
		Set("updated_at = ?", startedAt).
		Where("id = ?", roomID).
		Where("status = ?", fromStatus).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("conditional start live room: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *liveRepo) ConditionalStopRoom(ctx context.Context, roomID uint32, toStatus uint8, endedAt time.Time) (bool, error) {
	res, err := r.db.NewUpdate().Model((*model.LiveRooms)(nil)).
		Set("status = ?", toStatus).
		Set("current_session_id = NULL").
		Set("ended_at = ?", endedAt).
		Set("updated_at = ?", endedAt).
		Where("id = ?", roomID).
		Where("status = ?", constant.LiveRoomLiving).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("conditional stop live room: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *liveRepo) SetRoomBanned(ctx context.Context, roomID uint32) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*model.LiveRooms)(nil)).
		Set("status = ?", constant.LiveRoomBanned).
		Set("updated_at = ?", now).
		Where("id = ?", roomID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("ban live room: %w", err)
	}
	return nil
}

func (r *liveRepo) SetRoomIdle(ctx context.Context, roomID uint32) (bool, error) {
	now := time.Now()
	res, err := r.db.NewUpdate().Model((*model.LiveRooms)(nil)).
		Set("status = ?", constant.LiveRoomIdle).
		Set("updated_at = ?", now).
		Where("id = ?", roomID).
		Where("status = ?", constant.LiveRoomBanned).
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("unban live room: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ===== Sessions =====

func (r *liveRepo) ListSessions(ctx context.Context, f LiveSessionListFilter) ([]LiveSessionRow, int, error) {
	items := make([]LiveSessionRow, 0)
	q := r.db.NewSelect().
		TableExpr("live_sessions AS ls").
		ColumnExpr("ls.*").
		ColumnExpr("COALESCE(lc.name, '') AS category_name").
		ColumnExpr("COALESCE(lsc.name, '') AS sub_category_name").
		Join("LEFT JOIN live_categories AS lc ON lc.id = ls.category_id AND lc.deleted_at IS NULL").
		Join("LEFT JOIN live_categories AS lsc ON lsc.id = ls.sub_category_id AND lsc.deleted_at IS NULL")
	if f.RoomID != nil {
		q = q.Where("ls.room_id = ?", *f.RoomID)
	}
	if f.UserID != nil {
		q = q.Where("ls.user_id = ?", *f.UserID)
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count live sessions: %w", err)
	}
	if err := q.OrderExpr("ls.started_at DESC, ls.id DESC").Offset(f.Offset).Limit(f.Limit).Scan(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("list live sessions: %w", err)
	}
	return items, total, nil
}

func (r *liveRepo) CreateSession(ctx context.Context, s *model.LiveSessions) error {
	_, err := r.db.NewInsert().Model(s).Exec(ctx)
	if err != nil {
		return fmt.Errorf("create live session: %w", err)
	}
	return nil
}

func (r *liveRepo) EndSession(ctx context.Context, sessionID uint32, endedAt time.Time, endReason string) (bool, error) {
	res, err := r.db.NewUpdate().Model((*model.LiveSessions)(nil)).
		Set("ended_at = ?", endedAt).
		Set("end_reason = ?", endReason).
		Set("updated_at = ?", endedAt).
		Where("id = ?", sessionID).
		Where("ended_at IS NULL").
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("end live session: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

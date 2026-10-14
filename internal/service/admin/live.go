package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	internallock "github.com/ilaziness/orange-tv/internal/lock"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/utils"
	pkglock "github.com/ilaziness/orange-tv/pkg/lock"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

const liveWriteLockTTL = 10 * time.Second

// LiveService manages admin user-live APIs.
type LiveService interface {
	ListCategories(ctx context.Context, req *admindto.LiveCategoryListRequest) ([]admindto.LiveCategoryItem, int, error)
	CreateCategory(ctx context.Context, req *admindto.CreateLiveCategoryRequest) (*admindto.LiveCategoryItem, error)
	UpdateCategory(ctx context.Context, id uint32, req *admindto.UpdateLiveCategoryRequest) (*admindto.LiveCategoryItem, error)
	DeleteCategory(ctx context.Context, id uint32) error

	ListApplies(ctx context.Context, req *admindto.LiveApplyListRequest) ([]admindto.LiveApplyItem, int, error)
	ApproveApply(ctx context.Context, id, adminID uint32, req *admindto.ReviewLiveApplyRequest) error
	RejectApply(ctx context.Context, id, adminID uint32, req *admindto.ReviewLiveApplyRequest) error

	ListStreamers(ctx context.Context, req *admindto.LiveStreamerListRequest) ([]admindto.LiveStreamerItem, int, error)
	RevokeStreamer(ctx context.Context, userID uint32) error

	ListRooms(ctx context.Context, req *admindto.LiveRoomListRequest) ([]admindto.LiveRoomItem, int, error)
	ForceStopRoom(ctx context.Context, roomID uint32) error
	BanRoom(ctx context.Context, roomID uint32) error
	UnbanRoom(ctx context.Context, roomID uint32) error

	ListSessions(ctx context.Context, req *admindto.LiveSessionListRequest) ([]admindto.LiveSessionItem, int, error)
}

type liveService struct {
	repo   repository.LiveRepository
	locker pkglock.Locker
	log    *zap.Logger
}

// NewLiveService creates an admin LiveService.
func NewLiveService(repo repository.LiveRepository, locker pkglock.Locker, log *zap.Logger) LiveService {
	return &liveService{repo: repo, locker: locker, log: log}
}

// ===== Categories =====

func (s *liveService) ListCategories(ctx context.Context, req *admindto.LiveCategoryListRequest) ([]admindto.LiveCategoryItem, int, error) {
	items, total, err := s.repo.ListCategories(ctx, repository.LiveCategoryListFilter{
		Status:  req.Status,
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  req.GetOffset(),
		Limit:   req.GetLimit(),
	})
	if err != nil {
		s.log.Error("admin live: list categories failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LiveCategoryItem, 0, len(items))
	for i := range items {
		out = append(out, toAdminLiveCategory(&items[i]))
	}
	return out, total, nil
}

func (s *liveService) CreateCategory(ctx context.Context, req *admindto.CreateLiveCategoryRequest) (*admindto.LiveCategoryItem, error) {
	name := strings.TrimSpace(req.Name)
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	exists, err := s.repo.ExistsCategoryName(ctx, name, status, 0)
	if err != nil {
		s.log.Error("admin live: check category name failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.LiveCategoryNameExists
	}
	c := &model.LiveCategories{Name: name, Sort: req.Sort, Status: status}
	if err := s.repo.CreateCategory(ctx, c); err != nil {
		s.log.Error("admin live: create category failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	item := toAdminLiveCategory(c)
	return &item, nil
}

func (s *liveService) UpdateCategory(ctx context.Context, id uint32, req *admindto.UpdateLiveCategoryRequest) (*admindto.LiveCategoryItem, error) {
	c, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		s.log.Error("admin live: get category failed", zap.Uint32("id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return nil, errcode.LiveCategoryNotFound
	}
	if req.Name != nil {
		c.Name = strings.TrimSpace(*req.Name)
	}
	if req.Sort != nil {
		c.Sort = *req.Sort
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	exists, err := s.repo.ExistsCategoryName(ctx, c.Name, c.Status, c.ID)
	if err != nil {
		s.log.Error("admin live: check category name failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.LiveCategoryNameExists
	}
	if err := s.repo.UpdateCategory(ctx, c); err != nil {
		s.log.Error("admin live: update category failed", zap.Uint32("id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	item := toAdminLiveCategory(c)
	return &item, nil
}

func (s *liveService) DeleteCategory(ctx context.Context, id uint32) error {
	c, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		s.log.Error("admin live: get category for delete failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if c == nil {
		return errcode.LiveCategoryNotFound
	}
	if err := s.repo.SoftDeleteCategory(ctx, id); err != nil {
		s.log.Error("admin live: delete category failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// ===== Applies =====

func (s *liveService) ListApplies(ctx context.Context, req *admindto.LiveApplyListRequest) ([]admindto.LiveApplyItem, int, error) {
	items, total, err := s.repo.ListApplies(ctx, repository.LiveApplyListFilter{
		Status:  req.Status,
		UserID:  req.UserID,
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  req.GetOffset(),
		Limit:   req.GetLimit(),
	})
	if err != nil {
		s.log.Error("admin live: list applies failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LiveApplyItem, 0, len(items))
	for i := range items {
		out = append(out, toAdminLiveApply(&items[i]))
	}
	return out, total, nil
}

func (s *liveService) ApproveApply(ctx context.Context, id, adminID uint32, req *admindto.ReviewLiveApplyRequest) error {
	apply, err := s.repo.GetApply(ctx, id)
	if err != nil {
		s.log.Error("admin live: get apply failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if apply == nil {
		return errcode.LiveApplyNotFound
	}
	if apply.Status != constant.StreamerApplyPending {
		return errcode.LiveApplyInvalidStatus
	}

	lock, err := s.locker.Lock(ctx, internallock.LiveApplyUserKey(apply.UserID), pkglock.WithTTL(liveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return errcode.TooManyRequests
	}
	if err != nil {
		s.log.Error("admin live: acquire apply lock failed", zap.Uint32("user_id", apply.UserID), zap.Error(err))
		return errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	remark := ""
	if req != nil {
		remark = strings.TrimSpace(req.Remark)
	}
	now := time.Now()

	err = s.repo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := repository.NewLiveRepoFromIDB(tx)
		ok, err := txRepo.ApproveApply(ctx, id, adminID, remark, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.LiveApplyInvalidStatus
		}
		if err := txRepo.MarkStreamer(ctx, apply.UserID, now); err != nil {
			return err
		}
		room, err := txRepo.GetRoomByUser(ctx, apply.UserID)
		if err != nil {
			return err
		}
		if room != nil {
			return nil
		}
		key, err := service.GenerateLiveStreamKey()
		if err != nil {
			return err
		}
		return txRepo.CreateRoom(ctx, &model.LiveRooms{
			UserID:    apply.UserID,
			StreamKey: key,
			Status:    constant.LiveRoomIdle,
			Title:     "我的直播间",
		})
	})
	if err != nil {
		if code, ok := errcode.As(err); ok {
			return code
		}
		s.log.Error("admin live: approve apply failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *liveService) RejectApply(ctx context.Context, id, adminID uint32, req *admindto.ReviewLiveApplyRequest) error {
	apply, err := s.repo.GetApply(ctx, id)
	if err != nil {
		s.log.Error("admin live: get apply failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if apply == nil {
		return errcode.LiveApplyNotFound
	}
	if apply.Status != constant.StreamerApplyPending {
		return errcode.LiveApplyInvalidStatus
	}

	lock, err := s.locker.Lock(ctx, internallock.LiveApplyUserKey(apply.UserID), pkglock.WithTTL(liveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return errcode.TooManyRequests
	}
	if err != nil {
		s.log.Error("admin live: acquire apply lock failed", zap.Uint32("user_id", apply.UserID), zap.Error(err))
		return errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	remark := ""
	if req != nil {
		remark = strings.TrimSpace(req.Remark)
	}
	ok, err := s.repo.RejectApply(ctx, id, adminID, remark, time.Now())
	if err != nil {
		s.log.Error("admin live: reject apply failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if !ok {
		return errcode.LiveApplyInvalidStatus
	}
	return nil
}

// ===== Streamers =====

func (s *liveService) ListStreamers(ctx context.Context, req *admindto.LiveStreamerListRequest) ([]admindto.LiveStreamerItem, int, error) {
	items, total, err := s.repo.ListStreamers(ctx, repository.LiveStreamerListFilter{
		Keyword:    strings.TrimSpace(req.Keyword),
		RoomStatus: req.RoomStatus,
		Offset:     req.GetOffset(),
		Limit:      req.GetLimit(),
	})
	if err != nil {
		s.log.Error("admin live: list streamers failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LiveStreamerItem, 0, len(items))
	for i := range items {
		out = append(out, toAdminLiveStreamer(&items[i]))
	}
	return out, total, nil
}

func (s *liveService) RevokeStreamer(ctx context.Context, userID uint32) error {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		s.log.Error("admin live: get user for revoke failed", zap.Uint32("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if user == nil || user.IsStreamer == 0 {
		return errcode.LiveStreamerNotFound
	}

	room, err := s.repo.GetRoomByUser(ctx, userID)
	if err != nil {
		s.log.Error("admin live: get room for revoke failed", zap.Uint32("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}

	if room != nil {
		lock, err := s.locker.Lock(ctx, internallock.LiveRoomKey(room.ID), pkglock.WithTTL(liveWriteLockTTL))
		if errors.Is(err, pkglock.ErrLockNotHeld) {
			return errcode.TooManyRequests
		}
		if err != nil {
			s.log.Error("admin live: acquire room lock failed", zap.Uint32("room_id", room.ID), zap.Error(err))
			return errcode.Wrap(errcode.TooManyRequests, err)
		}
		defer func() { _ = lock.Release(ctx) }()

		room, err = s.repo.GetRoom(ctx, room.ID)
		if err != nil {
			s.log.Error("admin live: re-get room for revoke failed", zap.Uint32("user_id", userID), zap.Error(err))
			return errcode.Wrap(errcode.DatabaseError, err)
		}
	}

	err = s.repo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := repository.NewLiveRepoFromIDB(tx)
		if room != nil && room.Status == constant.LiveRoomLiving {
			if err := stopLivingRoom(ctx, txRepo, room, constant.LiveEndReasonRevoke, false); err != nil {
				return err
			}
		}
		ok, err := txRepo.RevokeStreamer(ctx, userID)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.LiveStreamerNotFound
		}
		return nil
	})
	if err != nil {
		if code, ok := errcode.As(err); ok {
			return code
		}
		s.log.Error("admin live: revoke streamer failed", zap.Uint32("user_id", userID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// ===== Rooms =====

func (s *liveService) ListRooms(ctx context.Context, req *admindto.LiveRoomListRequest) ([]admindto.LiveRoomItem, int, error) {
	items, total, err := s.repo.ListRooms(ctx, repository.LiveRoomListFilter{
		Status:     req.Status,
		CategoryID: req.CategoryID,
		UserID:     req.UserID,
		Keyword:    strings.TrimSpace(req.Keyword),
		Offset:     req.GetOffset(),
		Limit:      req.GetLimit(),
	})
	if err != nil {
		s.log.Error("admin live: list rooms failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LiveRoomItem, 0, len(items))
	for i := range items {
		out = append(out, toAdminLiveRoom(&items[i]))
	}
	return out, total, nil
}

func (s *liveService) ForceStopRoom(ctx context.Context, roomID uint32) error {
	return s.withRoomLock(ctx, roomID, func(room *model.LiveRooms) error {
		if room.Status != constant.LiveRoomLiving {
			return errcode.LiveRoomNotLiving
		}
		return s.repo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
			return stopLivingRoom(ctx, repository.NewLiveRepoFromIDB(tx), room, constant.LiveEndReasonAdmin, false)
		})
	})
}

func (s *liveService) BanRoom(ctx context.Context, roomID uint32) error {
	return s.withRoomLock(ctx, roomID, func(room *model.LiveRooms) error {
		if room.Status == constant.LiveRoomBanned {
			return nil
		}
		return s.repo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
			txRepo := repository.NewLiveRepoFromIDB(tx)
			if room.Status == constant.LiveRoomLiving {
				if err := stopLivingRoom(ctx, txRepo, room, constant.LiveEndReasonBanned, true); err != nil {
					return err
				}
				return nil
			}
			return txRepo.SetRoomBanned(ctx, room.ID)
		})
	})
}

func (s *liveService) UnbanRoom(ctx context.Context, roomID uint32) error {
	return s.withRoomLock(ctx, roomID, func(room *model.LiveRooms) error {
		if room.Status != constant.LiveRoomBanned {
			return errcode.WithMessage(errcode.ParamError, "直播间未处于禁播状态")
		}
		ok, err := s.repo.SetRoomIdle(ctx, room.ID)
		if err != nil {
			return errcode.Wrap(errcode.DatabaseError, err)
		}
		if !ok {
			return errcode.WithMessage(errcode.ParamError, "直播间未处于禁播状态")
		}
		return nil
	})
}

func (s *liveService) withRoomLock(ctx context.Context, roomID uint32, fn func(room *model.LiveRooms) error) error {
	room, err := s.repo.GetRoom(ctx, roomID)
	if err != nil {
		s.log.Error("admin live: get room failed", zap.Uint32("room_id", roomID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return errcode.LiveRoomNotFound
	}
	lock, err := s.locker.Lock(ctx, internallock.LiveRoomKey(roomID), pkglock.WithTTL(liveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return errcode.TooManyRequests
	}
	if err != nil {
		s.log.Error("admin live: acquire room lock failed", zap.Uint32("room_id", roomID), zap.Error(err))
		return errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	room, err = s.repo.GetRoom(ctx, roomID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return errcode.LiveRoomNotFound
	}
	if err := fn(room); err != nil {
		if code, ok := errcode.As(err); ok {
			return code
		}
		s.log.Error("admin live: room write failed", zap.Uint32("room_id", roomID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// ===== Sessions =====

func (s *liveService) ListSessions(ctx context.Context, req *admindto.LiveSessionListRequest) ([]admindto.LiveSessionItem, int, error) {
	items, total, err := s.repo.ListSessions(ctx, repository.LiveSessionListFilter{
		RoomID: req.RoomID,
		UserID: req.UserID,
		Offset: req.GetOffset(),
		Limit:  req.GetLimit(),
	})
	if err != nil {
		s.log.Error("admin live: list sessions failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.LiveSessionItem, 0, len(items))
	for i := range items {
		out = append(out, toAdminLiveSession(&items[i]))
	}
	return out, total, nil
}

func stopLivingRoom(ctx context.Context, repo repository.LiveRepository, room *model.LiveRooms, endReason string, ban bool) error {
	now := time.Now()
	toStatus := constant.LiveRoomIdle
	if ban {
		toStatus = constant.LiveRoomBanned
	}
	if room.CurrentSessionID != nil && *room.CurrentSessionID > 0 {
		if _, err := repo.EndSession(ctx, *room.CurrentSessionID, now, endReason); err != nil {
			return err
		}
	}
	ok, err := repo.ConditionalStopRoom(ctx, room.ID, toStatus, now)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.LiveRoomNotLiving
	}
	return nil
}

func toAdminLiveCategory(c *model.LiveCategories) admindto.LiveCategoryItem {
	return admindto.LiveCategoryItem{
		ID:        c.ID,
		Name:      c.Name,
		Sort:      c.Sort,
		Status:    c.Status,
		CreatedAt: utils.FormatTimeStr(c.CreatedAt),
		UpdatedAt: utils.FormatTimeStr(c.UpdatedAt),
	}
}

func toAdminLiveApply(a *repository.LiveApplyRow) admindto.LiveApplyItem {
	return admindto.LiveApplyItem{
		ID:            a.ID,
		UserID:        a.UserID,
		UserStrID:     a.UserStrID,
		Nickname:      a.Nickname,
		Email:         a.Email,
		Avatar:        a.Avatar,
		Reason:        a.Reason,
		Contact:       a.Contact,
		Status:        a.Status,
		ReviewAdminID: a.ReviewAdminID,
		ReviewRemark:  a.ReviewRemark,
		ReviewedAt:    utils.FormatTimePtrStr(a.ReviewedAt),
		CreatedAt:     utils.FormatTimeStr(a.CreatedAt),
	}
}

func toAdminLiveStreamer(r *repository.LiveStreamerRow) admindto.LiveStreamerItem {
	return admindto.LiveStreamerItem{
		UserID:       r.UserID,
		UserStrID:    r.UserStrID,
		Nickname:     r.Nickname,
		Email:        r.Email,
		Avatar:       r.Avatar,
		StreamerAt:   utils.FormatTimePtrStr(r.StreamerAt),
		RoomID:       r.RoomID,
		RoomTitle:    r.RoomTitle,
		RoomStatus:   r.RoomStatus,
		CategoryID:   r.CategoryID,
		CategoryName: r.CategoryName,
	}
}

func toAdminLiveRoom(r *repository.LiveRoomRow) admindto.LiveRoomItem {
	var sessionID uint32
	if r.CurrentSessionID != nil {
		sessionID = *r.CurrentSessionID
	}
	return admindto.LiveRoomItem{
		ID:               r.ID,
		UserID:           r.UserID,
		UserStrID:        r.UserStrID,
		Nickname:         r.Nickname,
		Avatar:           r.Avatar,
		CategoryID:       r.CategoryID,
		CategoryName:     r.CategoryName,
		Title:            r.Title,
		Cover:            r.Cover,
		Notice:           r.Notice,
		StreamKey:        r.StreamKey,
		Status:           r.Status,
		CurrentSessionID: sessionID,
		StartedAt:        utils.FormatTimePtrStr(r.StartedAt),
		EndedAt:          utils.FormatTimePtrStr(r.EndedAt),
		CreatedAt:        utils.FormatTimeStr(r.CreatedAt),
		UpdatedAt:        utils.FormatTimeStr(r.UpdatedAt),
	}
}

func toAdminLiveSession(s *repository.LiveSessionRow) admindto.LiveSessionItem {
	return admindto.LiveSessionItem{
		ID:               s.ID,
		RoomID:           s.RoomID,
		UserID:           s.UserID,
		Title:            s.Title,
		CategoryID:       s.CategoryID,
		CategoryName:     s.CategoryName,
		StreamerNickname: s.StreamerNickname,
		StreamerAvatar:   s.StreamerAvatar,
		StartedAt:        utils.FormatTimeStr(s.StartedAt),
		EndedAt:          utils.FormatTimePtrStr(s.EndedAt),
		EndReason:        s.EndReason,
	}
}

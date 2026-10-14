package client

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
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

const clientLiveWriteLockTTL = 10 * time.Second

// LiveService manages client user-live APIs.
type LiveService interface {
	CreateApply(ctx context.Context, userID uint32, req *clientdto.CreateLiveApplyRequest) (*clientdto.LiveApplyItem, error)
	GetMyApply(ctx context.Context, userID uint32) (*clientdto.LiveApplyItem, error)

	ListCategories(ctx context.Context) ([]clientdto.LiveCategoryItem, error)

	ListLivingRooms(ctx context.Context, req *clientdto.LiveRoomListRequest) ([]clientdto.LiveRoomListItem, int, error)
	GetRoom(ctx context.Context, roomID uint32) (*clientdto.LiveRoomDetail, error)

	GetMyRoom(ctx context.Context, userID uint32) (*clientdto.MyLiveRoom, error)
	UpdateMyRoom(ctx context.Context, userID uint32, req *clientdto.UpdateMyLiveRoomRequest) (*clientdto.MyLiveRoom, error)
	RegenerateKey(ctx context.Context, userID uint32) (*clientdto.MyLiveRoom, error)
	Start(ctx context.Context, userID uint32, req *clientdto.StartLiveRequest) (*clientdto.StartLiveResponse, error)
	Stop(ctx context.Context, userID uint32) error
}

type liveService struct {
	repo     repository.LiveRepository
	settings service.SettingsService
	locker   pkglock.Locker
	log      *zap.Logger
}

// NewLiveService creates a client LiveService.
func NewLiveService(repo repository.LiveRepository, settings service.SettingsService, locker pkglock.Locker, log *zap.Logger) LiveService {
	return &liveService{repo: repo, settings: settings, locker: locker, log: log}
}

func (s *liveService) CreateApply(ctx context.Context, userID uint32, req *clientdto.CreateLiveApplyRequest) (*clientdto.LiveApplyItem, error) {
	lock, err := s.locker.Lock(ctx, internallock.LiveApplyUserKey(userID), pkglock.WithTTL(clientLiveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return nil, errcode.TooManyRequests
	}
	if err != nil {
		s.log.Error("client live: acquire apply lock failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		s.log.Error("client live: get user failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if user == nil {
		return nil, errcode.UserNotFound
	}
	if user.IsStreamer == 1 {
		return nil, errcode.LiveApplyAlreadyStreamer
	}
	pending, err := s.repo.HasPendingApply(ctx, userID)
	if err != nil {
		s.log.Error("client live: check pending apply failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if pending {
		return nil, errcode.LiveApplyPendingExists
	}

	a := &model.StreamerApplies{
		UserID:  userID,
		Reason:  strings.TrimSpace(req.Reason),
		Contact: strings.TrimSpace(req.Contact),
		Status:  constant.StreamerApplyPending,
	}
	if err := s.repo.CreateApply(ctx, a); err != nil {
		s.log.Error("client live: create apply failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	item := toClientLiveApply(a)
	return &item, nil
}

func (s *liveService) GetMyApply(ctx context.Context, userID uint32) (*clientdto.LiveApplyItem, error) {
	a, err := s.repo.GetLatestApplyByUser(ctx, userID)
	if err != nil {
		s.log.Error("client live: get my apply failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if a == nil {
		return nil, nil
	}
	item := toClientLiveApply(a)
	return &item, nil
}

func (s *liveService) ListCategories(ctx context.Context) ([]clientdto.LiveCategoryItem, error) {
	items, err := s.repo.ListEnabledCategories(ctx)
	if err != nil {
		s.log.Error("client live: list categories failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.LiveCategoryItem, 0, len(items))
	for _, c := range items {
		out = append(out, clientdto.LiveCategoryItem{ID: c.ID, Name: c.Name, Sort: c.Sort})
	}
	return out, nil
}

func (s *liveService) ListLivingRooms(ctx context.Context, req *clientdto.LiveRoomListRequest) ([]clientdto.LiveRoomListItem, int, error) {
	items, total, err := s.repo.ListRooms(ctx, repository.LiveRoomListFilter{
		CategoryID: req.CategoryID,
		LivingOnly: true,
		Offset:     req.GetOffset(),
		Limit:      req.GetLimit(),
	})
	if err != nil {
		s.log.Error("client live: list rooms failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.LiveRoomListItem, 0, len(items))
	for i := range items {
		r := &items[i]
		out = append(out, clientdto.LiveRoomListItem{
			ID:               r.ID,
			Title:            r.Title,
			Cover:            r.Cover,
			CategoryID:       r.CategoryID,
			CategoryName:     r.CategoryName,
			StreamerUserID:   r.UserID,
			StreamerStrID:    r.UserStrID,
			StreamerNickname: r.Nickname,
			StreamerAvatar:   r.Avatar,
			StartedAt:        utils.FormatTimePtrStr(r.StartedAt),
		})
	}
	return out, total, nil
}

func (s *liveService) GetRoom(ctx context.Context, roomID uint32) (*clientdto.LiveRoomDetail, error) {
	row, err := s.repo.GetRoomRow(ctx, roomID)
	if err != nil {
		s.log.Error("client live: get room failed", zap.Uint32("room_id", roomID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if row == nil {
		return nil, errcode.LiveRoomNotFound
	}

	cfg, err := s.loadStreamConfig(ctx)
	if err != nil {
		return nil, err
	}
	hls, flv := cfg.ResolvePublicPlayURLs(row.StreamKey, row.Status)
	return &clientdto.LiveRoomDetail{
		ID:               row.ID,
		Title:            row.Title,
		Cover:            row.Cover,
		Notice:           row.Notice,
		CategoryID:       row.CategoryID,
		CategoryName:     row.CategoryName,
		Status:           row.Status,
		StreamerUserID:   row.UserID,
		StreamerStrID:    row.UserStrID,
		StreamerNickname: row.Nickname,
		StreamerAvatar:   row.Avatar,
		PlayHLSURL:       hls,
		PlayFLVURL:       flv,
		StartedAt:        utils.FormatTimePtrStr(row.StartedAt),
	}, nil
}

func (s *liveService) GetMyRoom(ctx context.Context, userID uint32) (*clientdto.MyLiveRoom, error) {
	if err := s.requireStreamer(ctx, userID); err != nil {
		return nil, err
	}
	room, err := s.repo.GetRoomByUser(ctx, userID)
	if err != nil {
		s.log.Error("client live: get my room failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return nil, errcode.LiveRoomNotFound
	}
	return s.toMyRoom(ctx, room)
}

func (s *liveService) UpdateMyRoom(ctx context.Context, userID uint32, req *clientdto.UpdateMyLiveRoomRequest) (*clientdto.MyLiveRoom, error) {
	if err := s.requireStreamer(ctx, userID); err != nil {
		return nil, err
	}
	room, err := s.repo.GetRoomByUser(ctx, userID)
	if err != nil {
		s.log.Error("client live: get my room failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return nil, errcode.LiveRoomNotFound
	}
	if req.Title != nil {
		room.Title = strings.TrimSpace(*req.Title)
	}
	if req.Cover != nil {
		room.Cover = strings.TrimSpace(*req.Cover)
	}
	if req.Notice != nil {
		room.Notice = strings.TrimSpace(*req.Notice)
	}
	if req.CategoryID != nil {
		cat, err := s.repo.GetCategory(ctx, *req.CategoryID)
		if err != nil {
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if cat == nil || cat.Status != constant.StatusEnabled {
			return nil, errcode.LiveCategoryDisabled
		}
		room.CategoryID = *req.CategoryID
	}
	if err := s.repo.UpdateRoom(ctx, room); err != nil {
		s.log.Error("client live: update my room failed", zap.Uint32("user_id", userID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.toMyRoom(ctx, room)
}

func (s *liveService) RegenerateKey(ctx context.Context, userID uint32) (*clientdto.MyLiveRoom, error) {
	if err := s.requireStreamer(ctx, userID); err != nil {
		return nil, err
	}
	room, err := s.repo.GetRoomByUser(ctx, userID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return nil, errcode.LiveRoomNotFound
	}
	if room.Status == constant.LiveRoomLiving {
		return nil, errcode.LiveRoomLiving
	}

	lock, err := s.locker.Lock(ctx, internallock.LiveRoomKey(room.ID), pkglock.WithTTL(clientLiveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return nil, errcode.TooManyRequests
	}
	if err != nil {
		return nil, errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	room, err = s.repo.GetRoom(ctx, room.ID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return nil, errcode.LiveRoomNotFound
	}
	if room.Status == constant.LiveRoomLiving {
		return nil, errcode.LiveRoomLiving
	}

	key, err := service.GenerateLiveStreamKey()
	if err != nil {
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	ok, err := s.repo.UpdateRoomStreamKey(ctx, room.ID, key)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if !ok {
		return nil, errcode.LiveRoomLiving
	}
	room.StreamKey = key
	return s.toMyRoom(ctx, room)
}

func (s *liveService) Start(ctx context.Context, userID uint32, req *clientdto.StartLiveRequest) (*clientdto.StartLiveResponse, error) {
	if err := s.requireStreamer(ctx, userID); err != nil {
		return nil, err
	}
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if user == nil {
		return nil, errcode.UserNotFound
	}
	room, err := s.repo.GetRoomByUser(ctx, userID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return nil, errcode.LiveRoomNotFound
	}
	if room.Status == constant.LiveRoomBanned {
		return nil, errcode.LiveRoomBanned
	}

	cfg, err := s.loadStreamConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.ReadyForBroadcast() {
		return nil, errcode.LiveConfigIncomplete
	}

	lock, err := s.locker.Lock(ctx, internallock.LiveRoomKey(room.ID), pkglock.WithTTL(clientLiveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return nil, errcode.TooManyRequests
	}
	if err != nil {
		return nil, errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	room, err = s.repo.GetRoom(ctx, room.ID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return nil, errcode.LiveRoomNotFound
	}
	if room.Status == constant.LiveRoomBanned {
		return nil, errcode.LiveRoomBanned
	}
	// Idempotent: already living → return push info.
	if room.Status == constant.LiveRoomLiving {
		var sessionID uint32
		if room.CurrentSessionID != nil {
			sessionID = *room.CurrentSessionID
		}
		return &clientdto.StartLiveResponse{
			RoomID:     room.ID,
			SessionID:  sessionID,
			Status:     room.Status,
			PushServer: cfg.PushServer,
			StreamKey:  room.StreamKey,
			StartedAt:  utils.FormatTimePtrStr(room.StartedAt),
		}, nil
	}

	if req != nil {
		if req.Title != nil {
			room.Title = strings.TrimSpace(*req.Title)
		}
		if req.CategoryID != nil {
			room.CategoryID = *req.CategoryID
		}
	}
	if room.CategoryID == 0 {
		return nil, errcode.WithMessage(errcode.ParamError, "请选择直播分类")
	}
	cat, err := s.repo.GetCategory(ctx, room.CategoryID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if cat == nil || cat.Status != constant.StatusEnabled {
		return nil, errcode.LiveCategoryDisabled
	}
	if strings.TrimSpace(room.Title) == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "请填写直播标题")
	}

	now := time.Now()
	var sessionID uint32
	err = s.repo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := repository.NewLiveRepoFromIDB(tx)
		if err := txRepo.UpdateRoom(ctx, room); err != nil {
			return err
		}
		session := service.NewLiveSessionSnapshot(room, user, now)
		if err := txRepo.CreateSession(ctx, session); err != nil {
			return err
		}
		sessionID = session.ID
		ok, err := txRepo.ConditionalStartRoom(ctx, room.ID, session.ID, constant.LiveRoomIdle, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.LiveRoomLiving
		}
		return nil
	})
	if err != nil {
		if code, ok := errcode.As(err); ok {
			return nil, code
		}
		s.log.Error("client live: start failed", zap.Uint32("room_id", room.ID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	return &clientdto.StartLiveResponse{
		RoomID:     room.ID,
		SessionID:  sessionID,
		Status:     constant.LiveRoomLiving,
		PushServer: cfg.PushServer,
		StreamKey:  room.StreamKey,
		StartedAt:  utils.FormatTimeStr(now),
	}, nil
}

func (s *liveService) Stop(ctx context.Context, userID uint32) error {
	if err := s.requireStreamer(ctx, userID); err != nil {
		return err
	}
	room, err := s.repo.GetRoomByUser(ctx, userID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return errcode.LiveRoomNotFound
	}
	if room.Status != constant.LiveRoomLiving {
		return errcode.LiveRoomNotLiving
	}

	lock, err := s.locker.Lock(ctx, internallock.LiveRoomKey(room.ID), pkglock.WithTTL(clientLiveWriteLockTTL))
	if errors.Is(err, pkglock.ErrLockNotHeld) {
		return errcode.TooManyRequests
	}
	if err != nil {
		return errcode.Wrap(errcode.TooManyRequests, err)
	}
	defer func() { _ = lock.Release(ctx) }()

	room, err = s.repo.GetRoom(ctx, room.ID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if room == nil {
		return errcode.LiveRoomNotFound
	}
	if room.Status != constant.LiveRoomLiving {
		return errcode.LiveRoomNotLiving
	}

	now := time.Now()
	err = s.repo.RunInTx(ctx, func(ctx context.Context, tx bun.Tx) error {
		txRepo := repository.NewLiveRepoFromIDB(tx)
		if room.CurrentSessionID != nil && *room.CurrentSessionID > 0 {
			if _, err := txRepo.EndSession(ctx, *room.CurrentSessionID, now, constant.LiveEndReasonHost); err != nil {
				return err
			}
		}
		ok, err := txRepo.ConditionalStopRoom(ctx, room.ID, constant.LiveRoomIdle, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.LiveRoomNotLiving
		}
		return nil
	})
	if err != nil {
		if code, ok := errcode.As(err); ok {
			return code
		}
		s.log.Error("client live: stop failed", zap.Uint32("room_id", room.ID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *liveService) requireStreamer(ctx context.Context, userID uint32) error {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if user == nil {
		return errcode.UserNotFound
	}
	if user.IsStreamer == 0 {
		return errcode.LiveNotStreamer
	}
	return nil
}

func (s *liveService) loadStreamConfig(ctx context.Context) (service.LiveStreamConfig, error) {
	m, err := s.settings.LoadMapByGroup(ctx, constant.SettingGroupLive)
	if err != nil {
		s.log.Error("client live: load live settings failed", zap.Error(err))
		return service.LiveStreamConfig{}, errcode.Wrap(errcode.DatabaseError, err)
	}
	return service.ParseLiveStreamConfig(m), nil
}

func (s *liveService) toMyRoom(ctx context.Context, room *model.LiveRooms) (*clientdto.MyLiveRoom, error) {
	cfg, err := s.loadStreamConfig(ctx)
	if err != nil {
		return nil, err
	}
	catName := ""
	if room.CategoryID > 0 {
		cat, err := s.repo.GetCategory(ctx, room.CategoryID)
		if err != nil {
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if cat != nil {
			catName = cat.Name
		}
	}
	var sessionID uint32
	if room.CurrentSessionID != nil {
		sessionID = *room.CurrentSessionID
	}
	return &clientdto.MyLiveRoom{
		ID:               room.ID,
		CategoryID:       room.CategoryID,
		CategoryName:     catName,
		Title:            room.Title,
		Cover:            room.Cover,
		Notice:           room.Notice,
		Status:           room.Status,
		StreamKey:        room.StreamKey,
		PushServer:       cfg.PushServer,
		CurrentSessionID: sessionID,
		StartedAt:        utils.FormatTimePtrStr(room.StartedAt),
		EndedAt:          utils.FormatTimePtrStr(room.EndedAt),
	}, nil
}

func toClientLiveApply(a *model.StreamerApplies) clientdto.LiveApplyItem {
	return clientdto.LiveApplyItem{
		ID:           a.ID,
		Reason:       a.Reason,
		Contact:      a.Contact,
		Status:       a.Status,
		ReviewRemark: a.ReviewRemark,
		ReviewedAt:   utils.FormatTimePtrStr(a.ReviewedAt),
		CreatedAt:    utils.FormatTimeStr(a.CreatedAt),
	}
}

package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// LiveService manages live channels for admin.
type LiveService interface {
	List(ctx context.Context, req *admindto.LiveListRequest) ([]admindto.LiveChannelItem, int, error)
	Create(ctx context.Context, req *admindto.CreateLiveRequest) (*admindto.LiveChannelItem, error)
	Update(ctx context.Context, id uint32, req *admindto.UpdateLiveRequest) (*admindto.LiveChannelItem, error)
	Delete(ctx context.Context, id uint32) error
	SyncFromSource(ctx context.Context, sourceURL string) (*admindto.LiveSyncResult, error)
}

type liveService struct {
	repo  repository.LiveRepository
	cache *cache.Manager
	log   *zap.Logger
}

// NewLiveService creates a LiveService.
func NewLiveService(repo repository.LiveRepository, c *cache.Manager, log *zap.Logger) LiveService {
	if log == nil {
		log = zap.NewNop()
	}
	return &liveService{repo: repo, cache: c, log: log}
}

func (s *liveService) List(ctx context.Context, req *admindto.LiveListRequest) ([]admindto.LiveChannelItem, int, error) {
	items, total, err := s.repo.List(ctx, repository.LiveListFilter{
		Category: strings.TrimSpace(req.Category),
		Keyword:  strings.TrimSpace(req.Keyword),
		Status:   req.Status,
		Offset:   req.GetOffset(),
		Limit:    req.GetLimit(),
	})
	if err != nil {
		s.log.Error("live: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapLiveItems(items, true), total, nil
}

func (s *liveService) Create(ctx context.Context, req *admindto.CreateLiveRequest) (*admindto.LiveChannelItem, error) {
	name := strings.TrimSpace(req.Name)
	streamURL := strings.TrimSpace(req.StreamURL)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "频道名称不能为空")
	}
	if streamURL == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "直播流地址不能为空")
	}
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	desc := strings.TrimSpace(req.Description)
	item := &model.LiveChannels{
		Name:        name,
		Category:    strings.TrimSpace(req.Category),
		StreamURL:   streamURL,
		Logo:        strings.TrimSpace(req.Logo),
		Description: desc,
		SortOrder:   req.SortOrder,
		Status:      status,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		s.log.Error("live: create failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateLive(ctx)
	out := toLiveItem(item, true)
	return &out, nil
}

func (s *liveService) Update(ctx context.Context, id uint32, req *admindto.UpdateLiveRequest) (*admindto.LiveChannelItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("live: get by id failed", zap.Uint32("live_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return nil, errcode.LiveChannelNotFound
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "频道名称不能为空")
		}
		item.Name = name
	}
	if req.Category != nil {
		item.Category = strings.TrimSpace(*req.Category)
	}
	if req.StreamURL != nil {
		url := strings.TrimSpace(*req.StreamURL)
		if url == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "直播流地址不能为空")
		}
		item.StreamURL = url
	}
	if req.Logo != nil {
		item.Logo = strings.TrimSpace(*req.Logo)
	}
	if req.Description != nil {
		item.Description = strings.TrimSpace(*req.Description)
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Update(ctx, item); err != nil {
		s.log.Error("live: update failed", zap.Uint32("live_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateLive(ctx)
	out := toLiveItem(item, true)
	return &out, nil
}

func (s *liveService) Delete(ctx context.Context, id uint32) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("live: get by id for delete failed", zap.Uint32("live_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.LiveChannelNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("live: delete failed", zap.Uint32("live_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateLive(ctx)
	return nil
}

func mapLiveItems(items []model.LiveChannels, withStatus bool) []admindto.LiveChannelItem {
	out := make([]admindto.LiveChannelItem, 0, len(items))
	for i := range items {
		out = append(out, toLiveItem(&items[i], withStatus))
	}
	return out
}

func toLiveItem(m *model.LiveChannels, withStatus bool) admindto.LiveChannelItem {
	item := admindto.LiveChannelItem{
		ID:          m.ID,
		Name:        m.Name,
		Category:    m.Category,
		StreamURL:   m.StreamURL,
		Logo:        m.Logo,
		Description: m.Description,
		SortOrder:   m.SortOrder,
	}
	if withStatus {
		item.Status = m.Status
	}
	return item
}

func (s *liveService) SyncFromSource(ctx context.Context, sourceURL string) (*admindto.LiveSyncResult, error) {
	parser, err := selectParser(sourceURL)
	if err != nil {
		s.log.Error("live: select parser failed", zap.String("url", sourceURL), zap.Error(err))
		return nil, errcode.WithMessage(errcode.LiveSyncFailed, err.Error())
	}

	raw, err := fetchLiveSource(ctx, sourceURL)
	if err != nil {
		s.log.Error("live: fetch from source failed", zap.String("url", sourceURL), zap.Error(err))
		return nil, errcode.WithMessage(errcode.LiveSyncFailed, err.Error())
	}

	entries := parser.Parse(raw)
	if len(entries) == 0 {
		s.log.Warn("live: no entries parsed from source", zap.String("url", sourceURL))
		return nil, errcode.WithMessage(errcode.LiveSyncFailed, "no valid entries parsed from source")
	}

	existing, err := s.repo.ListAll(ctx)
	if err != nil {
		s.log.Error("live: list all for sync failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	existingMap := make(map[string]*model.LiveChannels, len(existing))
	for i := range existing {
		existingMap[existing[i].Name] = &existing[i]
	}

	seenNames := make(map[string]bool, len(entries))
	var toCreate []model.LiveChannels
	var toDeleteIDs []uint32

	for _, entry := range entries {
		seenNames[entry.Name] = true
		if item, ok := existingMap[entry.Name]; ok {
			item.Category = entry.Category
			item.StreamURL = entry.StreamURL
			item.Logo = entry.Logo
			item.SortOrder = entry.SortOrder
			if err := s.repo.Update(ctx, item); err != nil {
				s.log.Error("live: sync update failed", zap.String("name", entry.Name), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
		} else {
			toCreate = append(toCreate, model.LiveChannels{
				Name:      entry.Name,
				Category:  entry.Category,
				StreamURL: entry.StreamURL,
				Logo:      entry.Logo,
				SortOrder: entry.SortOrder,
				Status:    constant.StatusEnabled,
			})
		}
	}

	for i := range existing {
		if !seenNames[existing[i].Name] {
			toDeleteIDs = append(toDeleteIDs, existing[i].ID)
		}
	}

	if err := s.repo.BatchCreate(ctx, toCreate); err != nil {
		s.log.Error("live: sync batch create failed", zap.Int("count", len(toCreate)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	if err := s.repo.DeleteByIDs(ctx, toDeleteIDs); err != nil {
		s.log.Error("live: sync delete by ids failed", zap.Int("count", len(toDeleteIDs)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	result := &admindto.LiveSyncResult{
		Total:   len(entries),
		Created: len(toCreate),
		Updated: len(entries) - len(toCreate),
		Deleted: len(toDeleteIDs),
	}
	// 同步会批量增删改，按分类精确失效成本高，直接清空直播列表缓存；
	// 缓存会在下次访问时按需重建。
	s.cache.InvalidateLive(ctx)
	return result, nil
}

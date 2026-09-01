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
	"github.com/ilaziness/orange-tv/internal/service"
	"go.uber.org/zap"
)

// LiveTVService manages livetv channels for admin.
type LiveTVService interface {
	List(ctx context.Context, req *admindto.LiveTVListRequest) ([]admindto.LiveTVChannelItem, int, error)
	Create(ctx context.Context, req *admindto.CreateLiveTVRequest) (*admindto.LiveTVChannelItem, error)
	Update(ctx context.Context, id uint32, req *admindto.UpdateLiveTVRequest) (*admindto.LiveTVChannelItem, error)
	Delete(ctx context.Context, id uint32) error
	SyncFromSource(ctx context.Context, sourceURL string) (*admindto.LiveTVSyncResult, error)
	GetSyncSourceURL(ctx context.Context) (string, error)
	SaveSyncSourceURL(ctx context.Context, sourceURL string) error
}

type liveTVService struct {
	repo     repository.LiveTVRepository
	cache    *cache.Manager
	settings service.SettingsService
	log      *zap.Logger
}

// NewLiveTVService creates a LiveTVService.
func NewLiveTVService(repo repository.LiveTVRepository, c *cache.Manager, settings service.SettingsService, log *zap.Logger) LiveTVService {
	if log == nil {
		log = zap.NewNop()
	}
	return &liveTVService{repo: repo, cache: c, settings: settings, log: log}
}

func (s *liveTVService) List(ctx context.Context, req *admindto.LiveTVListRequest) ([]admindto.LiveTVChannelItem, int, error) {
	items, total, err := s.repo.List(ctx, repository.LiveTVListFilter{
		Category: strings.TrimSpace(req.Category),
		Keyword:  strings.TrimSpace(req.Keyword),
		Status:   req.Status,
		Offset:   req.GetOffset(),
		Limit:    req.GetLimit(),
	})
	if err != nil {
		s.log.Error("livetv: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapLiveTVItems(items, true), total, nil
}

func (s *liveTVService) Create(ctx context.Context, req *admindto.CreateLiveTVRequest) (*admindto.LiveTVChannelItem, error) {
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
	item := &model.LivetvChannels{
		Name:        name,
		Category:    strings.TrimSpace(req.Category),
		StreamURL:   streamURL,
		Logo:        strings.TrimSpace(req.Logo),
		Description: desc,
		SortOrder:   req.SortOrder,
		Status:      status,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		s.log.Error("livetv: create failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateLiveTV(ctx)
	out := toLiveTVItem(item, true)
	return &out, nil
}

func (s *liveTVService) Update(ctx context.Context, id uint32, req *admindto.UpdateLiveTVRequest) (*admindto.LiveTVChannelItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("livetv: get by id failed", zap.Uint32("livetv_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return nil, errcode.LiveTVChannelNotFound
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
		s.log.Error("livetv: update failed", zap.Uint32("livetv_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateLiveTV(ctx)
	out := toLiveTVItem(item, true)
	return &out, nil
}

func (s *liveTVService) Delete(ctx context.Context, id uint32) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("livetv: get by id for delete failed", zap.Uint32("livetv_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.LiveTVChannelNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("livetv: delete failed", zap.Uint32("livetv_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateLiveTV(ctx)
	return nil
}

func mapLiveTVItems(items []model.LivetvChannels, withStatus bool) []admindto.LiveTVChannelItem {
	out := make([]admindto.LiveTVChannelItem, 0, len(items))
	for i := range items {
		out = append(out, toLiveTVItem(&items[i], withStatus))
	}
	return out
}

func toLiveTVItem(m *model.LivetvChannels, withStatus bool) admindto.LiveTVChannelItem {
	item := admindto.LiveTVChannelItem{
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

func (s *liveTVService) GetSyncSourceURL(ctx context.Context) (string, error) {
	m, err := s.settings.LoadMapByGroup(ctx, constant.SettingGroupLiveTV)
	if err != nil {
		s.log.Error("livetv: load sync source url failed", zap.Error(err))
		return "", errcode.Wrap(errcode.DatabaseError, err)
	}
	return service.StrVal(m, constant.SettingLiveTVSyncSourceURL), nil
}

func (s *liveTVService) SaveSyncSourceURL(ctx context.Context, sourceURL string) error {
	url := strings.TrimSpace(sourceURL)
	if url == "" {
		return errcode.WithMessage(errcode.ParamError, "直播源地址不能为空")
	}
	if err := s.settings.UpsertMany(ctx, constant.SettingGroupLiveTV, []repository.SettingUpsert{{
		Key:         constant.SettingLiveTVSyncSourceURL,
		Group:       constant.SettingGroupLiveTV,
		Value:       url,
		SettingType: constant.SettingTypeString,
		Description: "直播源同步地址",
	}}); err != nil {
		s.log.Error("livetv: save sync source url failed", zap.String("url", url), zap.Error(err))
		return err
	}
	return nil
}

func (s *liveTVService) SyncFromSource(ctx context.Context, sourceURL string) (*admindto.LiveTVSyncResult, error) {
	sourceURL = strings.TrimSpace(sourceURL)
	parser, err := selectParser(sourceURL)
	if err != nil {
		s.log.Error("livetv: select parser failed", zap.String("url", sourceURL), zap.Error(err))
		return nil, errcode.WithMessage(errcode.LiveTVSyncFailed, err.Error())
	}

	raw, err := fetchLiveTVSource(ctx, sourceURL)
	if err != nil {
		s.log.Error("livetv: fetch from source failed", zap.String("url", sourceURL), zap.Error(err))
		return nil, errcode.WithMessage(errcode.LiveTVSyncFailed, err.Error())
	}

	entries := parser.Parse(raw)
	if len(entries) == 0 {
		s.log.Warn("livetv: no entries parsed from source", zap.String("url", sourceURL))
		return nil, errcode.WithMessage(errcode.LiveTVSyncFailed, "no valid entries parsed from source")
	}

	existing, err := s.repo.ListAll(ctx)
	if err != nil {
		s.log.Error("livetv: list all for sync failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	existingMap := make(map[string]*model.LivetvChannels, len(existing))
	for i := range existing {
		existingMap[existing[i].Name] = &existing[i]
	}

	seenNames := make(map[string]bool, len(entries))
	var toCreate []model.LivetvChannels
	var toDeleteIDs []uint32

	for _, entry := range entries {
		name := strings.TrimSpace(entry.Name)
		if name == "" {
			continue
		}
		seenNames[name] = true
		if item, ok := existingMap[name]; ok {
			item.Category = strings.TrimSpace(entry.Category)
			item.StreamURL = strings.TrimSpace(entry.StreamURL)
			item.Logo = strings.TrimSpace(entry.Logo)
			item.SortOrder = entry.SortOrder
			if err := s.repo.Update(ctx, item); err != nil {
				s.log.Error("livetv: sync update failed", zap.String("name", name), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
		} else {
			toCreate = append(toCreate, model.LivetvChannels{
				Name:      name,
				Category:  strings.TrimSpace(entry.Category),
				StreamURL: strings.TrimSpace(entry.StreamURL),
				Logo:      strings.TrimSpace(entry.Logo),
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
		s.log.Error("livetv: sync batch create failed", zap.Int("count", len(toCreate)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	if err := s.repo.DeleteByIDs(ctx, toDeleteIDs); err != nil {
		s.log.Error("livetv: sync delete by ids failed", zap.Int("count", len(toDeleteIDs)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}

	result := &admindto.LiveTVSyncResult{
		Total:   len(entries),
		Created: len(toCreate),
		Updated: len(entries) - len(toCreate),
		Deleted: len(toDeleteIDs),
	}
	// 同步会批量增删改，按分类精确失效成本高，直接清空直播列表缓存；
	// 缓存会在下次访问时按需重建。
	s.cache.InvalidateLiveTV(ctx)
	return result, nil
}

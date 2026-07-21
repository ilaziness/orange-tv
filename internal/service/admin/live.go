package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// LiveService manages live channels for admin.
type LiveService interface {
	List(ctx context.Context, req *admindto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error)
	Create(ctx context.Context, req *admindto.CreateLiveRequest) (*shareddto.LiveChannelItem, error)
	Update(ctx context.Context, id int64, req *admindto.UpdateLiveRequest) (*shareddto.LiveChannelItem, error)
	Delete(ctx context.Context, id int64) error
}

type liveService struct {
	repo repository.LiveRepository
}

// NewLiveService creates a LiveService.
func NewLiveService(repo repository.LiveRepository) LiveService {
	return &liveService{repo: repo}
}

func (s *liveService) List(ctx context.Context, req *admindto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error) {
	items, total, err := s.repo.List(ctx, repository.LiveListFilter{
		Category: strings.TrimSpace(req.Category),
		Keyword:  strings.TrimSpace(req.Keyword),
		Status:   req.Status,
		Offset:   req.GetOffset(),
		Limit:    req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapLiveItems(items, true), total, nil
}

func (s *liveService) Create(ctx context.Context, req *admindto.CreateLiveRequest) (*shareddto.LiveChannelItem, error) {
	name := strings.TrimSpace(req.Name)
	streamURL := strings.TrimSpace(req.StreamURL)
	if name == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "频道名称不能为空")
	}
	if streamURL == "" {
		return nil, errcode.WithMessage(errcode.ParamError, "直播流地址不能为空")
	}
	status := uint8(constant.StatusEnabled)
	if req.Status != nil {
		status = *req.Status
	}
	desc := strings.TrimSpace(req.Description)
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}
	item := &model.LiveChannels{
		Name:        name,
		Category:    strings.TrimSpace(req.Category),
		StreamURL:   streamURL,
		Logo:        strings.TrimSpace(req.Logo),
		Description: descPtr,
		SortOrder:   req.SortOrder,
		Status:      status,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := toLiveItem(item, true)
	return &out, nil
}

func (s *liveService) Update(ctx context.Context, id int64, req *admindto.UpdateLiveRequest) (*shareddto.LiveChannelItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
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
		desc := strings.TrimSpace(*req.Description)
		if desc == "" {
			item.Description = nil
		} else {
			item.Description = &desc
		}
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := toLiveItem(item, true)
	return &out, nil
}

func (s *liveService) Delete(ctx context.Context, id int64) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.LiveChannelNotFound
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func mapLiveItems(items []model.LiveChannels, withStatus bool) []shareddto.LiveChannelItem {
	out := make([]shareddto.LiveChannelItem, 0, len(items))
	for i := range items {
		out = append(out, toLiveItem(&items[i], withStatus))
	}
	return out
}

func toLiveItem(m *model.LiveChannels, withStatus bool) shareddto.LiveChannelItem {
	desc := ""
	if m.Description != nil {
		desc = *m.Description
	}
	item := shareddto.LiveChannelItem{
		ID:          m.ID,
		Name:        m.Name,
		Category:    m.Category,
		StreamURL:   m.StreamURL,
		Logo:        m.Logo,
		Description: desc,
		SortOrder:   m.SortOrder,
	}
	if withStatus {
		item.Status = m.Status
	}
	return item
}

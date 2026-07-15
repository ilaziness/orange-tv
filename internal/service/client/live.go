package client

import (
	"context"
	"strings"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// LiveService provides public live channel queries.
type LiveService interface {
	List(ctx context.Context, req *clientdto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error)
}

type liveService struct {
	repo repository.LiveRepository
}

// NewLiveService creates a client LiveService.
func NewLiveService(repo repository.LiveRepository) LiveService {
	return &liveService{repo: repo}
}

func (s *liveService) List(ctx context.Context, req *clientdto.LiveListRequest) ([]shareddto.LiveChannelItem, int, error) {
	items, total, err := s.repo.List(ctx, repository.LiveListFilter{
		Category:   strings.TrimSpace(req.Category),
		OnlyOnline: true,
		Offset:     req.GetOffset(),
		Limit:      req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapLiveItems(items), total, nil
}

func mapLiveItems(items []model.LiveChannels) []shareddto.LiveChannelItem {
	out := make([]shareddto.LiveChannelItem, 0, len(items))
	for i := range items {
		m := &items[i]
		desc := ""
		if m.Description != nil {
			desc = *m.Description
		}
		out = append(out, shareddto.LiveChannelItem{
			ID:          m.ID,
			Name:        m.Name,
			Category:    m.Category,
			StreamURL:   m.StreamURL,
			Logo:        m.Logo,
			Description: desc,
			SortOrder:   m.SortOrder,
		})
	}
	return out
}

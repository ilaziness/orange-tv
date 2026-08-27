package client

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// AdService manages advertisements for the client API.
type AdService interface {
	List(ctx context.Context, scene string) ([]clientdto.AdItem, error)
}

type adService struct {
	repo repository.AdRepository
	log  *zap.Logger
}

// NewAdService creates a client AdService.
func NewAdService(repo repository.AdRepository, log *zap.Logger) AdService {
	return &adService{repo: repo, log: log}
}

func (s *adService) List(ctx context.Context, scene string) ([]clientdto.AdItem, error) {
	status := constant.StatusEnabled
	var (
		items []model.Advertisements
		err   error
	)
	if scene != "" {
		items, err = s.repo.ListByScene(ctx, scene, &status)
	} else {
		items, err = s.repo.ListByStatus(ctx, &status)
	}
	if err != nil {
		s.log.Error("client ad: list failed", zap.String("scene", scene), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toAdItems(items), nil
}

func toAdItems(items []model.Advertisements) []clientdto.AdItem {
	out := make([]clientdto.AdItem, 0, len(items))
	for _, a := range items {
		out = append(out, clientdto.AdItem{
			ID:          a.ID,
			AdKey:       a.AdKey,
			Type:        a.Type,
			ContentURL:  a.ContentURL,
			ContentCode: a.ContentCode,
			LinkURL:     a.LinkURL,
			Duration:    a.Duration,
		})
	}
	return out
}

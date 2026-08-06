package client

import (
	"context"

	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// AdService manages advertisements for the client API.
type AdService interface {
	ListByScene(ctx context.Context, scene string) ([]clientdto.AdItem, error)
}

type adService struct {
	repo repository.AdRepository
	log  *zap.Logger
}

// NewAdService creates a client AdService.
func NewAdService(repo repository.AdRepository, log *zap.Logger) AdService {
	if log == nil {
		log = zap.NewNop()
	}
	return &adService{repo: repo, log: log}
}

func (s *adService) ListByScene(ctx context.Context, scene string) ([]clientdto.AdItem, error) {
	enabled := uint8(1)
	items, err := s.repo.ListByScene(ctx, scene, &enabled)
	if err != nil {
		s.log.Error("client ad: list by scene failed", zap.String("scene", scene), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
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
	return out, nil
}

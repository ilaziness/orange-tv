package client

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// BannerService provides client banner queries.
type BannerService interface {
	ListBanners(ctx context.Context) ([]clientdto.BannerItem, error)
}

type bannerService struct {
	userRepo repository.UserFeatureRepository
	log      *zap.Logger
}

// NewBannerService creates a client BannerService.
func NewBannerService(userRepo repository.UserFeatureRepository, log *zap.Logger) BannerService {
	if log == nil {
		log = zap.NewNop()
	}
	return &bannerService{userRepo: userRepo, log: log}
}

func (s *bannerService) ListBanners(ctx context.Context) ([]clientdto.BannerItem, error) {
	status := constant.StatusEnabled
	items, err := s.userRepo.ListBanners(ctx, &status)
	if err != nil {
		s.log.Error("client banner: list failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]clientdto.BannerItem, 0, len(items))
	for _, b := range items {
		out = append(out, clientdto.BannerItem{
			ID:      b.ID,
			Title:   strings.TrimSpace(b.Title),
			Cover:   b.Cover,
			Link:    b.Link,
			VideoID: b.VideoID,
		})
	}
	return out, nil
}

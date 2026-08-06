package admin

import (
	"context"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// AdService manages advertisements for the admin API.
type AdService interface {
	List(ctx context.Context, offset, limit int) ([]admindto.AdItem, int, error)
	Create(ctx context.Context, req *admindto.CreateAdRequest) (*admindto.AdItem, error)
	Update(ctx context.Context, id uint32, req *admindto.UpdateAdRequest) (*admindto.AdItem, error)
	Delete(ctx context.Context, id uint32) error
}

type adService struct {
	repo repository.AdRepository
	log  *zap.Logger
}

// NewAdService creates an AdService.
func NewAdService(repo repository.AdRepository, log *zap.Logger) AdService {
	if log == nil {
		log = zap.NewNop()
	}
	return &adService{repo: repo, log: log}
}

func (s *adService) List(ctx context.Context, offset, limit int) ([]admindto.AdItem, int, error) {
	items, total, err := s.repo.ListAll(ctx, offset, limit)
	if err != nil {
		s.log.Error("ad: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.AdItem, 0, len(items))
	for _, a := range items {
		out = append(out, *toAdItem(&a))
	}
	return out, total, nil
}

func (s *adService) Create(ctx context.Context, req *admindto.CreateAdRequest) (*admindto.AdItem, error) {
	adKey := strings.TrimSpace(req.AdKey)
	// Check ad_key uniqueness.
	existing, err := s.repo.GetByKey(ctx, adKey)
	if err != nil {
		s.log.Error("ad: check key uniqueness failed", zap.String("ad_key", adKey), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if existing != nil {
		return nil, errcode.AdKeyExists
	}
	// Business validation: type=code requires content_code; otherwise requires content_url.
	if req.Type == constant.AdTypeCode {
		if req.ContentCode == nil || strings.TrimSpace(*req.ContentCode) == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "code类型广告必须提供广告代码")
		}
	} else {
		if strings.TrimSpace(req.ContentURL) == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "非code类型广告必须提供素材URL")
		}
	}

	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	duration := req.Duration
	if duration == 0 {
		duration = 5
	}
	a := &model.Advertisements{
		AdKey:       adKey,
		Title:       strings.TrimSpace(req.Title),
		Scene:       req.Scene,
		Type:        req.Type,
		ContentURL:  strings.TrimSpace(req.ContentURL),
		ContentCode: req.ContentCode,
		LinkURL:     strings.TrimSpace(req.LinkURL),
		Duration:    duration,
		Sort:        req.Sort,
		Status:      status,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		s.log.Error("ad: create failed", zap.String("ad_key", adKey), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toAdItem(a), nil
}

func (s *adService) Update(ctx context.Context, id uint32, req *admindto.UpdateAdRequest) (*admindto.AdItem, error) {
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		s.log.Error("ad: get for update failed", zap.Uint32("id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if a == nil {
		return nil, errcode.AdNotFound
	}
	// Check ad_key uniqueness if changing.
	if req.AdKey != nil {
		newKey := strings.TrimSpace(*req.AdKey)
		if newKey != a.AdKey {
			existing, err := s.repo.GetByKey(ctx, newKey)
			if err != nil {
				s.log.Error("ad: check key uniqueness failed", zap.String("ad_key", newKey), zap.Error(err))
				return nil, errcode.Wrap(errcode.DatabaseError, err)
			}
			if existing != nil {
				return nil, errcode.AdKeyExists
			}
			a.AdKey = newKey
		}
	}
	if req.Title != nil {
		a.Title = strings.TrimSpace(*req.Title)
	}
	if req.Scene != nil {
		a.Scene = *req.Scene
	}
	if req.Type != nil {
		a.Type = *req.Type
	}
	if req.ContentURL != nil {
		a.ContentURL = strings.TrimSpace(*req.ContentURL)
	}
	if req.ContentCode != nil {
		a.ContentCode = req.ContentCode
	}
	if req.LinkURL != nil {
		a.LinkURL = strings.TrimSpace(*req.LinkURL)
	}
	if req.Duration != nil {
		a.Duration = *req.Duration
	}
	if req.Sort != nil {
		a.Sort = *req.Sort
	}
	if req.Status != nil {
		a.Status = *req.Status
	}
	// Business validation: type=code requires content_code; otherwise requires content_url.
	if a.Type == constant.AdTypeCode {
		if a.ContentCode == nil || strings.TrimSpace(*a.ContentCode) == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "code类型广告必须提供广告代码")
		}
	} else {
		if strings.TrimSpace(a.ContentURL) == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "非code类型广告必须提供素材URL")
		}
	}
	if err := s.repo.Update(ctx, a); err != nil {
		s.log.Error("ad: update failed", zap.Uint32("id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toAdItem(a), nil
}

func (s *adService) Delete(ctx context.Context, id uint32) error {
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		s.log.Error("ad: get for delete failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if a == nil {
		return errcode.AdNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("ad: delete failed", zap.Uint32("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func toAdItem(a *model.Advertisements) *admindto.AdItem {
	return &admindto.AdItem{
		ID:          a.ID,
		AdKey:       a.AdKey,
		Title:       a.Title,
		Scene:       a.Scene,
		Type:        a.Type,
		ContentURL:  a.ContentURL,
		ContentCode: a.ContentCode,
		LinkURL:     a.LinkURL,
		Duration:    a.Duration,
		Sort:        a.Sort,
		Status:      a.Status,
	}
}

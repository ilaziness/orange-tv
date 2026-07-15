package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

const themeCurrentCacheKey = "theme:current"
const themeCurrentCacheTTL = 10 * time.Minute

// ThemeService provides public theme API.
type ThemeService interface {
	Current(ctx context.Context) (*shareddto.ThemeCurrentResponse, error)
}

type themeService struct {
	repo       repository.ThemeRepository
	adminTheme adminsvc.ThemeService
	cache      cache.Cache
}

// NewThemeService creates a client ThemeService.
func NewThemeService(repo repository.ThemeRepository, adminTheme adminsvc.ThemeService, c cache.Cache) ThemeService {
	if c == nil {
		c = cache.NewNopCache()
	}
	return &themeService{repo: repo, adminTheme: adminTheme, cache: c}
}

func (s *themeService) Current(ctx context.Context) (*shareddto.ThemeCurrentResponse, error) {
	if v, err := s.cache.Get(ctx, themeCurrentCacheKey); err == nil {
		if t, ok := v.(*shareddto.ThemeCurrentResponse); ok && t != nil {
			return t, nil
		}
	}
	if s.adminTheme != nil {
		_ = s.adminTheme.EnsureDefaultTheme(ctx)
	}
	item, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return nil, errcode.ThemeNoActive
	}
	cfg := map[string]any{}
	if item.Config != nil && *item.Config != "" {
		_ = json.Unmarshal([]byte(*item.Config), &cfg)
	}
	css, js := "", ""
	if item.CustomCss != nil {
		css = *item.CustomCss
	}
	if item.CustomJs != nil {
		js = *item.CustomJs
	}
	out := &shareddto.ThemeCurrentResponse{
		Name:       item.Name,
		Identifier: item.Identifier,
		Version:    item.Version,
		Config:     cfg,
		CustomCSS:  css,
		CustomJS:   js,
	}
	_ = s.cache.Set(ctx, themeCurrentCacheKey, out, themeCurrentCacheTTL)
	return out, nil
}

// InvalidateThemeCache clears the public theme cache.
func InvalidateThemeCache(ctx context.Context, c cache.Cache) {
	if c == nil {
		return
	}
	_ = c.Delete(ctx, themeCurrentCacheKey)
}

package client

import (
	"context"
	"encoding/json"

	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
)

// ThemeService provides public theme API.
type ThemeService interface {
	Current(ctx context.Context) (*shareddto.ThemeCurrentResponse, error)
}

type themeService struct {
	repo       repository.ThemeRepository
	adminTheme adminsvc.ThemeService
}

// NewThemeService creates a client ThemeService.
func NewThemeService(repo repository.ThemeRepository, adminTheme adminsvc.ThemeService) ThemeService {
	return &themeService{repo: repo, adminTheme: adminTheme}
}

func (s *themeService) Current(ctx context.Context) (*shareddto.ThemeCurrentResponse, error) {
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
	return &shareddto.ThemeCurrentResponse{
		Name:       item.Name,
		Identifier: item.Identifier,
		Version:    item.Version,
		Config:     cfg,
		CustomCSS:  css,
		CustomJS:   js,
	}, nil
}

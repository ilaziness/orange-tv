package admin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

// ThemeService manages themes for admin.
type ThemeService interface {
	List(ctx context.Context) ([]shareddto.ThemeListItem, error)
	Update(ctx context.Context, id int64, req *admindto.UpdateThemeRequest) (*shareddto.ThemeListItem, error)
	Activate(ctx context.Context, id int64) error
	EnsureDefaultTheme(ctx context.Context) error
}

type themeService struct {
	repo  repository.ThemeRepository
	cache cache.Cache
}

// NewThemeService creates a ThemeService.
func NewThemeService(repo repository.ThemeRepository, c cache.Cache) ThemeService {
	if c == nil {
		c = cache.NewNopCache()
	}
	return &themeService{repo: repo, cache: c}
}

func (s *themeService) List(ctx context.Context) ([]shareddto.ThemeListItem, error) {
	if err := s.EnsureDefaultTheme(ctx); err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]shareddto.ThemeListItem, 0, len(items))
	for i := range items {
		out = append(out, toThemeListItem(&items[i]))
	}
	return out, nil
}

func (s *themeService) Update(ctx context.Context, id int64, req *admindto.UpdateThemeRequest) (*shareddto.ThemeListItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return nil, errcode.ThemeNotFound
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "主题名称不能为空")
		}
		item.Name = name
	}
	if req.Config != nil {
		b, err := json.Marshal(req.Config)
		if err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "主题 config 无效")
		}
		cfg := string(b)
		item.Config = &cfg
	}
	if req.CustomCSS != nil {
		css := *req.CustomCSS
		item.CustomCss = &css
	}
	if req.CustomJS != nil {
		js := *req.CustomJS
		item.CustomJs = &js
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.cache.Delete(ctx, "theme:current")
	out := toThemeListItem(item)
	return &out, nil
}

func (s *themeService) Activate(ctx context.Context, id int64) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if item == nil {
		return errcode.ThemeNotFound
	}
	if err := s.repo.Activate(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.cache.Delete(ctx, "theme:current")
	return nil
}

func (s *themeService) EnsureDefaultTheme(ctx context.Context) error {
	cfg := defaultThemeConfigJSON()
	desc := "系统默认主题"
	m := &model.Themes{
		Name:        "默认主题",
		Identifier:  constant.ThemeIdentifierDefault,
		Version:     "1.0.0",
		Author:      "系统",
		Description: &desc,
		Config:      &cfg,
		IsDefault:   1,
		IsActive:    1,
	}
	if err := s.repo.EnsureDefault(ctx, m); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func defaultThemeConfigJSON() string {
	return `{
  "primary_color": "#1890ff",
  "secondary_color": "#52c41a",
  "background_color": "#f0f2f5",
  "text_color": "#262626",
  "header_height": "64px",
  "sidebar_width": "240px",
  "enable_dark_mode": false,
  "custom_fonts": []
}`
}

func toThemeListItem(m *model.Themes) shareddto.ThemeListItem {
	cfg := map[string]any{}
	if m.Config != nil && *m.Config != "" {
		_ = json.Unmarshal([]byte(*m.Config), &cfg)
	}
	desc, css, js := "", "", ""
	if m.Description != nil {
		desc = *m.Description
	}
	if m.CustomCss != nil {
		css = *m.CustomCss
	}
	if m.CustomJs != nil {
		js = *m.CustomJs
	}
	return shareddto.ThemeListItem{
		ID:           m.ID,
		Name:         m.Name,
		Identifier:   m.Identifier,
		Version:      m.Version,
		Author:       m.Author,
		Description:  desc,
		PreviewImage: m.PreviewImage,
		Config:       cfg,
		CustomCSS:    css,
		CustomJS:     js,
		IsDefault:    m.IsDefault,
		IsActive:     m.IsActive,
	}
}

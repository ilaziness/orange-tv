package client

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// ClientSettingsService manages system settings for the client API surface.
// Only whitelisted groups are exposed; non-whitelisted groups return empty data.
type ClientSettingsService interface {
	// GetByGroup returns settings for a single group.
	// If the group is not in the client whitelist, an empty response is returned.
	GetByGroup(ctx context.Context, group string) (*clientdto.SettingsResponse, error)
	// GetAll returns settings for all whitelisted groups.
	GetAll(ctx context.Context) (*clientdto.SettingsResponse, error)
	// GetPublic returns the public site response (site + ad groups combined).
	GetPublic(ctx context.Context) (*clientdto.PublicSiteResponse, error)
}

type clientSettingsService struct {
	repo  repository.SettingsRepository
	cache *cache.Manager
	log   *zap.Logger
}

// NewClientSettingsService creates a ClientSettingsService.
func NewClientSettingsService(repo repository.SettingsRepository, c *cache.Manager, log *zap.Logger) ClientSettingsService {
	if log == nil {
		log = zap.NewNop()
	}
	return &clientSettingsService{repo: repo, cache: c, log: log}
}

func (s *clientSettingsService) GetByGroup(ctx context.Context, group string) (*clientdto.SettingsResponse, error) {
	if !constant.IsValidSettingGroup(group) {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置分组")
	}
	resp := &clientdto.SettingsResponse{Group: group}
	if !constant.IsClientAllowedGroup(group) {
		return resp, nil
	}
	m, err := s.loadMapByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	switch group {
	case constant.SettingGroupSite:
		resp.Site = mapToSiteSettings(m)
	case constant.SettingGroupAd:
		resp.Ad = mapToAdSettings(m)
	}
	return resp, nil
}

func (s *clientSettingsService) GetAll(ctx context.Context) (*clientdto.SettingsResponse, error) {
	m, err := s.loadMapByGroups(ctx, constant.ClientAllowedGroups())
	if err != nil {
		return nil, err
	}
	return &clientdto.SettingsResponse{
		Site: mapToSiteSettings(filterByGroup(m, constant.SettingGroupSite)),
		Ad:   mapToAdSettings(filterByGroup(m, constant.SettingGroupAd)),
	}, nil
}

func (s *clientSettingsService) GetPublic(ctx context.Context) (*clientdto.PublicSiteResponse, error) {
	if p, err := s.cache.GetSettingsPublic(ctx); err == nil && p != nil {
		return p, nil
	}
	m, err := s.loadMapByGroups(ctx, []string{constant.SettingGroupSite, constant.SettingGroupAd})
	if err != nil {
		return nil, err
	}
	siteM := filterByGroup(m, constant.SettingGroupSite)
	adM := filterByGroup(m, constant.SettingGroupAd)
	pub := &clientdto.PublicSiteResponse{
		Name:        utils.DefaultStr(service.StrVal(siteM, constant.SettingSiteName), "Orange TV"),
		Logo:        service.StrVal(siteM, constant.SettingSiteLogo),
		Copyright:   service.StrVal(siteM, constant.SettingSiteCopyright),
		ICP:         service.StrVal(siteM, constant.SettingSiteICP),
		SEOKeywords: service.StrVal(siteM, constant.SettingSiteSEOKeywords),
		Description: service.StrVal(siteM, constant.SettingSiteDescription),
		Ad:          mapToAdSettings(adM),
	}
	_ = s.cache.SetSettingsPublic(ctx, pub)
	return pub, nil
}

func (s *clientSettingsService) loadMapByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	if m, err := s.cache.GetSettingsByGroup(ctx, group); err == nil && m != nil {
		return m, nil
	}
	m, err := s.repo.GetByGroup(ctx, group)
	if err != nil {
		s.log.Error("client settings: list by group failed", zap.String("group", group), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.cache.SetSettingsByGroup(ctx, group, m)
	return m, nil
}

func (s *clientSettingsService) loadMapByGroups(ctx context.Context, groups []string) (map[string]model.SystemSettings, error) {
	result := make(map[string]model.SystemSettings)
	missing := make([]string, 0, len(groups))
	for _, g := range groups {
		if m, err := s.cache.GetSettingsByGroup(ctx, g); err == nil && m != nil {
			for k, v := range m {
				result[k] = v
			}
		} else {
			missing = append(missing, g)
		}
	}
	if len(missing) > 0 {
		items, err := s.repo.GetByGroups(ctx, missing)
		if err != nil {
			s.log.Error("client settings: list by groups failed", zap.Strings("groups", missing), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		perGroup := make(map[string]map[string]model.SystemSettings)
		for k, v := range items {
			g := v.SettingGroup
			if perGroup[g] == nil {
				perGroup[g] = make(map[string]model.SystemSettings)
			}
			perGroup[g][k] = v
			result[k] = v
		}
		for g, m := range perGroup {
			_ = s.cache.SetSettingsByGroup(ctx, g, m)
		}
	}
	return result, nil
}

func filterByGroup(m map[string]model.SystemSettings, group string) map[string]model.SystemSettings {
	out := make(map[string]model.SystemSettings, len(m))
	for k, v := range m {
		if v.SettingGroup == group {
			out[k] = v
		}
	}
	return out
}

func mapToSiteSettings(m map[string]model.SystemSettings) clientdto.SiteSettings {
	return clientdto.SiteSettings{
		Name:        utils.DefaultStr(service.StrVal(m, constant.SettingSiteName), "Orange TV"),
		Logo:        service.StrVal(m, constant.SettingSiteLogo),
		Copyright:   service.StrVal(m, constant.SettingSiteCopyright),
		ICP:         service.StrVal(m, constant.SettingSiteICP),
		SEOKeywords: service.StrVal(m, constant.SettingSiteSEOKeywords),
		Description: service.StrVal(m, constant.SettingSiteDescription),
	}
}

func mapToAdSettings(m map[string]model.SystemSettings) clientdto.AdSettings {
	return clientdto.AdSettings{
		Enabled:  service.BoolVal(m, constant.SettingVideoAdEnabled, false),
		Type:     service.StrVal(m, constant.SettingVideoAdType),
		URL:      service.StrVal(m, constant.SettingVideoAdUrl),
		Link:     service.StrVal(m, constant.SettingVideoAdLink),
		Duration: service.IntVal(m, constant.SettingVideoAdDuration, 5),
		Skipable: service.BoolVal(m, constant.SettingVideoAdSkipable, true),
	}
}

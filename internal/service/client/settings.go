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
// Only whitelisted groups are exposed; non-whitelisted groups return an error.
type ClientSettingsService interface {
	// GetByGroup returns settings for a single group.
	// If the group is not in the client whitelist, an error is returned.
	GetByGroup(ctx context.Context, group string) (any, error)
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

func (s *clientSettingsService) GetByGroup(ctx context.Context, group string) (any, error) {
	if !constant.IsValidSettingGroup(group) {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置分组")
	}
	if !constant.IsClientAllowedGroup(group) {
		return nil, errcode.WithMessage(errcode.ParamError, "无权限访问该设置分组")
	}
	m, err := s.loadMapByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	switch group {
	case constant.SettingGroupSite:
		return mapToSiteSettings(m), nil
	case constant.SettingGroupAd:
		return mapToAdSettings(m), nil
	case constant.SettingGroupFeature:
		return mapToFeatureSettings(m), nil
	default:
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置分组")
	}
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

func mapToFeatureSettings(m map[string]model.SystemSettings) clientdto.FeatureSettings {
	return clientdto.FeatureSettings{
		LiveEnabled:    service.BoolVal(m, constant.SettingFeatureLiveEnabled, false),
		CommentEnabled: service.BoolVal(m, constant.SettingFeatureCommentEnabled, true),
		CommentReview:  service.BoolVal(m, constant.SettingFeatureCommentReview, true),
	}
}

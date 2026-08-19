package service

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// SettingsService is the shared service for reading and writing system settings.
// It provides cached reads, DTO mapping for client-visible groups (site/feature),
// and upsert execution with cache invalidation for admin writes.
type SettingsService interface {
	// LoadMapByGroup loads settings for a single group as a key→model map (with cache).
	LoadMapByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error)
	// LoadGroupMaps loads multiple groups, returning group→key→model maps.
	LoadGroupMaps(ctx context.Context, groups []string) (map[string]map[string]model.SystemSettings, error)
	// MapGroupToResponse maps a single group's settings to its shared DTO (site/feature only).
	MapGroupToResponse(group string, m map[string]model.SystemSettings) (any, error)
	// MapGroupsToResponse maps multiple groups: single group → flat DTO, multiple → map[string]any.
	MapGroupsToResponse(groups []string, maps map[string]map[string]model.SystemSettings) (any, error)
	// UpsertMany executes upserts and invalidates settings cache.
	UpsertMany(ctx context.Context, group string, upserts []repository.SettingUpsert) error
}

type settingsService struct {
	repo  repository.SettingsRepository
	cache *cache.Manager
	log   *zap.Logger
}

// NewSettingsService creates a shared SettingsService.
func NewSettingsService(repo repository.SettingsRepository, c *cache.Manager, log *zap.Logger) SettingsService {
	if log == nil {
		log = zap.NewNop()
	}
	return &settingsService{repo: repo, cache: c, log: log}
}

func (s *settingsService) LoadMapByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	if m, err := s.cache.GetSettingsByGroup(ctx, group); err == nil && m != nil {
		return m, nil
	}
	m, err := s.repo.GetByGroup(ctx, group)
	if err != nil {
		s.log.Error("settings: list by group failed", zap.String("group", group), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.cache.SetSettingsByGroup(ctx, group, m)
	return m, nil
}

func (s *settingsService) LoadGroupMaps(ctx context.Context, groups []string) (map[string]map[string]model.SystemSettings, error) {
	out := make(map[string]map[string]model.SystemSettings, len(groups))
	for _, g := range groups {
		m, err := s.LoadMapByGroup(ctx, g)
		if err != nil {
			return nil, err
		}
		out[g] = m
	}
	return out, nil
}

func (s *settingsService) MapGroupToResponse(group string, m map[string]model.SystemSettings) (any, error) {
	switch group {
	case constant.SettingGroupSite:
		return mapToSiteSettings(m), nil
	case constant.SettingGroupFeature:
		return mapToFeatureSettings(m), nil
	default:
		return nil, fmt.Errorf("unsupported group for shared mapping: %s", group)
	}
}

func (s *settingsService) MapGroupsToResponse(groups []string, maps map[string]map[string]model.SystemSettings) (any, error) {
	if len(groups) == 1 {
		return s.MapGroupToResponse(groups[0], maps[groups[0]])
	}
	result := make(map[string]any, len(groups))
	for _, g := range groups {
		resp, err := s.MapGroupToResponse(g, maps[g])
		if err != nil {
			return nil, err
		}
		result[g] = resp
	}
	return result, nil
}

func (s *settingsService) UpsertMany(ctx context.Context, group string, upserts []repository.SettingUpsert) error {
	if err := s.repo.UpsertMany(ctx, upserts); err != nil {
		s.log.Error("settings: upsert many failed", zap.String("group", group), zap.Int("count", len(upserts)), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	s.cache.InvalidateSettings(ctx)
	return nil
}

// ===== mapping helpers (shared) =====

func mapToSiteSettings(m map[string]model.SystemSettings) dto.SiteSettings {
	return dto.SiteSettings{
		Name:          utils.DefaultStr(StrVal(m, constant.SettingSiteName), "小橘TV"),
		Logo:          StrVal(m, constant.SettingSiteLogo),
		Copyright:     StrVal(m, constant.SettingSiteCopyright),
		ICP:           StrVal(m, constant.SettingSiteICP),
		SEOKeywords:   StrVal(m, constant.SettingSiteSEOKeywords),
		Description:   StrVal(m, constant.SettingSiteDescription),
		AnalyticsCode: StrVal(m, constant.SettingSiteAnalyticsCode),
	}
}

func mapToFeatureSettings(m map[string]model.SystemSettings) dto.FeatureSettings {
	return dto.FeatureSettings{
		LiveEnabled:    BoolVal(m, constant.SettingFeatureLiveEnabled, false),
		CommentEnabled: BoolVal(m, constant.SettingFeatureCommentEnabled, true),
		CommentReview:  BoolVal(m, constant.SettingFeatureCommentReview, true),
		RatingEnabled:  BoolVal(m, constant.SettingFeatureRatingEnabled, true),
	}
}

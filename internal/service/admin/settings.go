package admin

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
)

const (
	cacheKeySettingsAll    = "settings:all"
	cacheKeySettingsPublic = "settings:public"
	settingsCacheTTL       = 5 * time.Minute
)

// SettingsService manages system settings.
type SettingsService interface {
	Get(ctx context.Context) (*admindto.SettingsResponse, error)
	Update(ctx context.Context, req *admindto.UpdateSettingsRequest) (*admindto.SettingsResponse, error)
	GetPublic(ctx context.Context) (*admindto.PublicSiteResponse, error)
	// ResourceConfig is used by open API for access control / format.
	ResourceConfig(ctx context.Context) (*ResourceConfig, error)
	InvalidateCache(ctx context.Context)
}

// ResourceConfig is runtime resource-station config.
type ResourceConfig struct {
	SiteMode                string
	APIOutputFormat         string
	EnableThirdPartyCollect bool
	APIKey                  string
}

type settingsService struct {
	repo  repository.SettingsRepository
	cache cache.Cache
}

// NewSettingsService creates a SettingsService.
func NewSettingsService(repo repository.SettingsRepository, c cache.Cache) SettingsService {
	if c == nil {
		c = cache.NewNopCache()
	}
	return &settingsService{repo: repo, cache: c}
}

func (s *settingsService) Get(ctx context.Context) (*admindto.SettingsResponse, error) {
	m, err := s.loadMap(ctx)
	if err != nil {
		return nil, err
	}
	return mapToAdminSettings(m), nil
}

func (s *settingsService) GetPublic(ctx context.Context) (*admindto.PublicSiteResponse, error) {
	if v, err := s.cache.Get(ctx, cacheKeySettingsPublic); err == nil {
		if p, ok := v.(*admindto.PublicSiteResponse); ok && p != nil {
			return p, nil
		}
	}
	m, err := s.loadMap(ctx)
	if err != nil {
		return nil, err
	}
	pub := &admindto.PublicSiteResponse{
		Name:        strVal(m, constant.SettingSiteName),
		Logo:        strVal(m, constant.SettingSiteLogo),
		Copyright:   strVal(m, constant.SettingSiteCopyright),
		ICP:         strVal(m, constant.SettingSiteICP),
		SEOKeywords: strVal(m, constant.SettingSiteSEOKeywords),
		Description: strVal(m, constant.SettingSiteDescription),
	}
	if pub.Name == "" {
		pub.Name = "Orange TV"
	}
	_ = s.cache.Set(ctx, cacheKeySettingsPublic, pub, settingsCacheTTL)
	return pub, nil
}

func (s *settingsService) ResourceConfig(ctx context.Context) (*ResourceConfig, error) {
	m, err := s.loadMap(ctx)
	if err != nil {
		return nil, err
	}
	return &ResourceConfig{
		SiteMode:                defaultStr(strVal(m, constant.SettingSiteMode), constant.SiteModeVideoSite),
		APIOutputFormat:         normalizeAPIOutputFormat(strVal(m, constant.SettingAPIOutputFormat)),
		EnableThirdPartyCollect: boolVal(m, constant.SettingEnableThirdPartyCollect, true),
		APIKey:                  strVal(m, constant.SettingResourceAPIKey),
	}, nil
}

func (s *settingsService) Update(ctx context.Context, req *admindto.UpdateSettingsRequest) (*admindto.SettingsResponse, error) {
	if req == nil || (req.Site == nil && req.API == nil) {
		return nil, errcode.WithMessage(errcode.ParamError, "无更新内容")
	}
	upserts := make([]repository.SettingUpsert, 0, 12)
	if req.Site != nil {
		site := req.Site
		if site.Name != nil {
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteName, Value: strings.TrimSpace(*site.Name),
				SettingType: constant.SettingTypeString, Description: "站点名称",
			})
		}
		if site.Logo != nil {
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteLogo, Value: strings.TrimSpace(*site.Logo),
				SettingType: constant.SettingTypeString, Description: "站点 Logo URL",
			})
		}
		if site.Copyright != nil {
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteCopyright, Value: strings.TrimSpace(*site.Copyright),
				SettingType: constant.SettingTypeString, Description: "站点版权信息",
			})
		}
		if site.ICP != nil {
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteICP, Value: strings.TrimSpace(*site.ICP),
				SettingType: constant.SettingTypeString, Description: "备案号",
			})
		}
		if site.SEOKeywords != nil {
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteSEOKeywords, Value: strings.TrimSpace(*site.SEOKeywords),
				SettingType: constant.SettingTypeString, Description: "SEO 关键词",
			})
		}
		if site.Description != nil {
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteDescription, Value: strings.TrimSpace(*site.Description),
				SettingType: constant.SettingTypeString, Description: "站点描述",
			})
		}
	}
	if req.API != nil {
		api := req.API
		if api.SiteMode != nil {
			mode := strings.TrimSpace(*api.SiteMode)
			if mode != constant.SiteModeVideoSite && mode != constant.SiteModeResourceSite {
				return nil, errcode.WithMessage(errcode.SettingInvalid, "站点模式无效")
			}
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingSiteMode, Value: mode,
				SettingType: constant.SettingTypeString, Description: "站点模式",
			})
		}
		if api.APIOutputFormat != nil {
			fmtv := strings.ToLower(strings.TrimSpace(*api.APIOutputFormat))
			if fmtv != constant.APIOutputDefault && fmtv != constant.APIOutputAppleCMS {
				return nil, errcode.WithMessage(errcode.SettingInvalid, "API 输出格式无效，仅支持 default / apple_cms")
			}
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingAPIOutputFormat, Value: fmtv,
				SettingType: constant.SettingTypeString, Description: "API输出格式：default(系统默认) apple_cms(苹果CMS)",
			})
		}
		if api.EnableThirdPartyCollect != nil {
			v := "0"
			if *api.EnableThirdPartyCollect {
				v = "1"
			}
			upserts = append(upserts, repository.SettingUpsert{
				Key: constant.SettingEnableThirdPartyCollect, Value: v,
				SettingType: constant.SettingTypeBoolean, Description: "是否允许第三方采集",
			})
		}
		if api.ResourceAPIKey != nil {
			key := strings.TrimSpace(*api.ResourceAPIKey)
			// empty / omit-after-trim means leave unchanged (avoid wiping key via masked UI)
			if key != "" {
				upserts = append(upserts, repository.SettingUpsert{
					Key: constant.SettingResourceAPIKey, Value: key,
					SettingType: constant.SettingTypeString, Description: "资源站 API 访问密钥",
				})
			}
		}
	}
	if len(upserts) == 0 {
		return nil, errcode.WithMessage(errcode.ParamError, "无更新内容")
	}
	if err := s.repo.UpsertMany(ctx, upserts); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	s.InvalidateCache(ctx)
	return s.Get(ctx)
}

func (s *settingsService) InvalidateCache(ctx context.Context) {
	_ = s.cache.Delete(ctx, cacheKeySettingsAll)
	_ = s.cache.Delete(ctx, cacheKeySettingsPublic)
}

func (s *settingsService) loadMap(ctx context.Context) (map[string]model.SystemSettings, error) {
	if v, err := s.cache.Get(ctx, cacheKeySettingsAll); err == nil {
		if m, ok := v.(map[string]model.SystemSettings); ok && m != nil {
			return m, nil
		}
	}
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	m := make(map[string]model.SystemSettings, len(items))
	for _, it := range items {
		m[it.SettingKey] = it
	}
	_ = s.cache.Set(ctx, cacheKeySettingsAll, m, settingsCacheTTL)
	return m, nil
}

func mapToAdminSettings(m map[string]model.SystemSettings) *admindto.SettingsResponse {
	key := strVal(m, constant.SettingResourceAPIKey)
	masked := ""
	if key != "" {
		masked = "******"
	}
	return &admindto.SettingsResponse{
		Site: admindto.SiteSettings{
			Name:        defaultStr(strVal(m, constant.SettingSiteName), "Orange TV"),
			Logo:        strVal(m, constant.SettingSiteLogo),
			Copyright:   strVal(m, constant.SettingSiteCopyright),
			ICP:         strVal(m, constant.SettingSiteICP),
			SEOKeywords: strVal(m, constant.SettingSiteSEOKeywords),
			Description: strVal(m, constant.SettingSiteDescription),
		},
		API: admindto.APISettings{
			SiteMode:                defaultStr(strVal(m, constant.SettingSiteMode), constant.SiteModeVideoSite),
			APIOutputFormat:         normalizeAPIOutputFormat(strVal(m, constant.SettingAPIOutputFormat)),
			EnableThirdPartyCollect: boolVal(m, constant.SettingEnableThirdPartyCollect, true),
			ResourceAPIKeySet:       key != "",
			ResourceAPIKeyMasked:    masked,
		},
	}
}

// normalizeAPIOutputFormat returns default or apple_cms only.
func normalizeAPIOutputFormat(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == constant.APIOutputAppleCMS {
		return constant.APIOutputAppleCMS
	}
	return constant.APIOutputDefault
}

func strVal(m map[string]model.SystemSettings, key string) string {
	it, ok := m[key]
	if !ok || it.SettingValue == nil {
		return ""
	}
	return *it.SettingValue
}

func boolVal(m map[string]model.SystemSettings, key string, def bool) bool {
	v := strings.TrimSpace(strVal(m, key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		if n, err := strconv.Atoi(v); err == nil {
			return n != 0
		}
		return def
	}
}

func defaultStr(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

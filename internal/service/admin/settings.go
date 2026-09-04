package admin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/validator"
	"go.uber.org/zap"
)

// SettingsService manages system settings.
type SettingsService interface {
	Get(ctx context.Context, group string) (any, error)
	Update(ctx context.Context, group string, data json.RawMessage) (any, error)
}

type settingsService struct {
	shared service.SettingsService
	log    *zap.Logger
}

// NewSettingsService creates a SettingsService.
func NewSettingsService(shared service.SettingsService, log *zap.Logger) SettingsService {
	if log == nil {
		log = zap.NewNop()
	}
	return &settingsService{shared: shared, log: log}
}

func (s *settingsService) Get(ctx context.Context, group string) (any, error) {
	m, err := s.shared.LoadMapByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	return s.mapToResponse(group, m), nil
}

func (s *settingsService) Update(ctx context.Context, group string, data json.RawMessage) (any, error) {
	if len(data) == 0 {
		return nil, errcode.WithMessage(errcode.ParamError, "无更新内容")
	}

	upserts, err := s.parseUpdateData(group, data)
	if err != nil {
		return nil, err
	}
	if len(upserts) == 0 {
		return nil, errcode.WithMessage(errcode.ParamError, "无更新内容")
	}

	if upsertErr := s.shared.UpsertMany(ctx, group, upserts); upsertErr != nil {
		return nil, upsertErr
	}

	m, err := s.shared.LoadMapByGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	return s.mapToResponse(group, m), nil
}

func (s *settingsService) parseUpdateData(group string, data json.RawMessage) ([]repository.SettingUpsert, error) {
	switch group {
	case constant.SettingGroupSite:
		var req admindto.UpdateSiteSettings
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
		}
		if err := validator.Validate(&req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		return s.buildSiteUpserts(&req), nil
	case constant.SettingGroupAPI:
		var req admindto.UpdateAPISettings
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
		}
		if err := validator.Validate(&req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		return s.buildAPIUpserts(&req), nil
	case constant.SettingGroupFeature:
		var req admindto.UpdateFeatureSettings
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
		}
		if err := validator.Validate(&req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		return s.buildFeatureUpserts(&req), nil
	case constant.SettingGroupSEO:
		var req admindto.UpdateSEOSettings
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
		}
		if err := validator.Validate(&req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		return s.buildSEOUpserts(&req)
	default:
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置分组")
	}
}

func (s *settingsService) buildSiteUpserts(site *admindto.UpdateSiteSettings) []repository.SettingUpsert {
	var upserts []repository.SettingUpsert
	if site.Name != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteName, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.Name),
			SettingType: constant.SettingTypeString, Description: "站点名称",
		})
	}
	if site.Logo != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteLogo, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.Logo),
			SettingType: constant.SettingTypeString, Description: "站点 Logo URL",
		})
	}
	if site.Copyright != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteCopyright, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.Copyright),
			SettingType: constant.SettingTypeString, Description: "站点版权信息",
		})
	}
	if site.ICP != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteICP, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.ICP),
			SettingType: constant.SettingTypeString, Description: "备案号",
		})
	}
	if site.SEOKeywords != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteSEOKeywords, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.SEOKeywords),
			SettingType: constant.SettingTypeString, Description: "SEO 关键词",
		})
	}
	if site.Description != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteDescription, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.Description),
			SettingType: constant.SettingTypeString, Description: "站点描述",
		})
	}
	if site.AnalyticsCode != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSiteAnalyticsCode, Group: constant.SettingGroupSite, Value: strings.TrimSpace(*site.AnalyticsCode),
			SettingType: constant.SettingTypeString, Description: "站点统计代码",
		})
	}
	return upserts
}

func (s *settingsService) buildAPIUpserts(api *admindto.UpdateAPISettings) []repository.SettingUpsert {
	var upserts []repository.SettingUpsert
	if api.EnableThirdPartyCollect != nil {
		v := "0"
		if *api.EnableThirdPartyCollect {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingEnableThirdPartyCollect, Group: constant.SettingGroupAPI, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "是否允许第三方采集",
		})
	}
	return upserts
}

func (s *settingsService) mapToResponse(group string, m map[string]model.SystemSettings) any {
	switch group {
	case constant.SettingGroupSite, constant.SettingGroupFeature:
		resp, err := s.shared.MapGroupToResponse(group, m)
		if err != nil {
			s.log.Error("settings: map group to response failed", zap.String("group", group), zap.Error(err))
			return nil
		}
		return resp
	case constant.SettingGroupAPI:
		return mapToAPISettings(m)
	case constant.SettingGroupSEO:
		return service.MapToSEOSettings(m)
	default:
		return nil
	}
}

func mapToAPISettings(m map[string]model.SystemSettings) admindto.APISettings {
	return admindto.APISettings{
		EnableThirdPartyCollect: service.BoolVal(m, constant.SettingEnableThirdPartyCollect, true),
	}
}

func (s *settingsService) buildFeatureUpserts(f *admindto.UpdateFeatureSettings) []repository.SettingUpsert {
	var upserts []repository.SettingUpsert
	if f.LiveTVEnabled != nil {
		v := "0"
		if *f.LiveTVEnabled {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureLiveTVEnabled, Group: constant.SettingGroupFeature, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "电视直播开关",
		})
	}
	if f.CommentEnabled != nil {
		v := "0"
		if *f.CommentEnabled {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureCommentEnabled, Group: constant.SettingGroupFeature, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "视频评论开关",
		})
	}
	if f.CommentReview != nil {
		v := "0"
		if *f.CommentReview {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureCommentReview, Group: constant.SettingGroupFeature, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "评论是否需要审核",
		})
	}
	if f.RatingEnabled != nil {
		v := "0"
		if *f.RatingEnabled {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureRatingEnabled, Group: constant.SettingGroupFeature, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "视频评分开关",
		})
	}
	return upserts
}

func (s *settingsService) buildSEOUpserts(seo *admindto.UpdateSEOSettings) ([]repository.SettingUpsert, error) {
	var upserts []repository.SettingUpsert
	if seo.PublicBaseURL != nil {
		normalized, err := service.NormalizePublicBaseURL(*seo.PublicBaseURL)
		if err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOPublicBaseURL, Group: constant.SettingGroupSEO, Value: normalized,
			SettingType: constant.SettingTypeString, Description: "公开站点根地址（无尾斜杠）",
		})
	}
	if seo.DefaultOGImage != nil {
		normalized, err := service.NormalizeOptionalHTTPURL(*seo.DefaultOGImage)
		if err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEODefaultOGImage, Group: constant.SettingGroupSEO, Value: normalized,
			SettingType: constant.SettingTypeString, Description: "默认 Open Graph 图片地址",
		})
	}
	if seo.SitemapEnabled != nil {
		v := "0"
		if *seo.SitemapEnabled {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOSitemapEnabled, Group: constant.SettingGroupSEO, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "是否输出 sitemap",
		})
	}
	if seo.LLMsEnabled != nil {
		v := "0"
		if *seo.LLMsEnabled {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOLLMsEnabled, Group: constant.SettingGroupSEO, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "是否输出 llms.txt",
		})
	}
	if seo.LLMsIntro != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOLLMsIntro, Group: constant.SettingGroupSEO, Value: strings.TrimSpace(*seo.LLMsIntro),
			SettingType: constant.SettingTypeString, Description: "llms.txt 站点简介",
		})
	}
	if seo.AllowAISearch != nil {
		v := "0"
		if *seo.AllowAISearch {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOAllowAISearch, Group: constant.SettingGroupSEO, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "是否允许 AI 检索类爬虫",
		})
	}
	if seo.AllowAITraining != nil {
		v := "0"
		if *seo.AllowAITraining {
			v = "1"
		}
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOAllowAITraining, Group: constant.SettingGroupSEO, Value: v,
			SettingType: constant.SettingTypeBoolean, Description: "是否允许 AI 训练类爬虫",
		})
	}
	if seo.GoogleSiteVerification != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOGoogleSiteVerification, Group: constant.SettingGroupSEO, Value: strings.TrimSpace(*seo.GoogleSiteVerification),
			SettingType: constant.SettingTypeString, Description: "Google 站点验证码",
		})
	}
	if seo.BaiduSiteVerification != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOBaiduSiteVerification, Group: constant.SettingGroupSEO, Value: strings.TrimSpace(*seo.BaiduSiteVerification),
			SettingType: constant.SettingTypeString, Description: "百度站点验证码",
		})
	}
	if seo.BingSiteVerification != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingSEOBingSiteVerification, Group: constant.SettingGroupSEO, Value: strings.TrimSpace(*seo.BingSiteVerification),
			SettingType: constant.SettingTypeString, Description: "Bing 站点验证码",
		})
	}
	return upserts, nil
}

package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	dto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"github.com/ilaziness/orange-tv/internal/validator"
	"github.com/ilaziness/orange-tv/pkg/objectstorage"
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

	upserts, err := s.parseUpdateData(ctx, group, data)
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

func (s *settingsService) parseUpdateData(ctx context.Context, group string, data json.RawMessage) ([]repository.SettingUpsert, error) {
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
		return s.parseFeatureUpdate(ctx, data)
	case constant.SettingGroupSEO:
		var req admindto.UpdateSEOSettings
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
		}
		if err := validator.Validate(&req); err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, err.Error())
		}
		return s.buildSEOUpserts(&req)
	case constant.SettingGroupStorage:
		return s.parseStorageUpdate(ctx, data)
	default:
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置分组")
	}
}

func (s *settingsService) parseFeatureUpdate(ctx context.Context, data json.RawMessage) ([]repository.SettingUpsert, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
	}

	var req admindto.UpdateFeatureSettings
	decode := func(field string) (*dto.PlatformFlags, error) {
		v, ok := raw[field]
		if !ok {
			return nil, nil
		}
		flags, err := service.DecodePlatformFlags(v)
		if err != nil {
			return nil, errcode.WithMessage(errcode.ParamError, field+": "+err.Error())
		}
		return &flags, nil
	}

	var err error
	if req.LiveTVEnabled, err = decode("livetv_enabled"); err != nil {
		return nil, err
	}
	if req.CommentEnabled, err = decode("comment_enabled"); err != nil {
		return nil, err
	}
	if req.CommentReview, err = decode("comment_review"); err != nil {
		return nil, err
	}
	if req.RatingEnabled, err = decode("rating_enabled"); err != nil {
		return nil, err
	}

	if req.LiveTVEnabled == nil && req.CommentEnabled == nil && req.CommentReview == nil && req.RatingEnabled == nil {
		return nil, nil
	}

	// Keep comment_review consistent with comment_enabled per platform.
	if req.CommentEnabled != nil || req.CommentReview != nil {
		commentEnabled := req.CommentEnabled
		commentReview := req.CommentReview
		if commentEnabled == nil || commentReview == nil {
			current, loadErr := s.shared.LoadMapByGroup(ctx, constant.SettingGroupFeature)
			if loadErr != nil {
				return nil, loadErr
			}
			if commentEnabled == nil {
				flags := service.ParsePlatformFlags(service.StrVal(current, constant.SettingFeatureCommentEnabled), true)
				commentEnabled = &flags
			}
			if commentReview == nil {
				flags := service.ParsePlatformFlags(service.StrVal(current, constant.SettingFeatureCommentReview), true)
				commentReview = &flags
			}
		}
		normalized := service.AndPlatformFlags(*commentReview, *commentEnabled)
		req.CommentReview = &normalized
		if req.CommentEnabled != nil {
			req.CommentEnabled = commentEnabled
		}
	}

	return s.buildFeatureUpserts(&req), nil
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
	case constant.SettingGroupSite:
		resp, err := s.shared.MapGroupToResponse(group, m)
		if err != nil {
			s.log.Error("settings: map group to response failed", zap.String("group", group), zap.Error(err))
			return nil
		}
		return resp
	case constant.SettingGroupFeature:
		return service.MapToFeatureMatrix(m)
	case constant.SettingGroupAPI:
		return mapToAPISettings(m)
	case constant.SettingGroupSEO:
		return service.MapToSEOSettings(m)
	case constant.SettingGroupStorage:
		return service.MapToStorageSettings(m)
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
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureLiveTVEnabled, Group: constant.SettingGroupFeature,
			Value:       service.MarshalPlatformFlags(*f.LiveTVEnabled),
			SettingType: constant.SettingTypeJSON, Description: "电视直播开关（按端）",
		})
	}
	if f.CommentEnabled != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureCommentEnabled, Group: constant.SettingGroupFeature,
			Value:       service.MarshalPlatformFlags(*f.CommentEnabled),
			SettingType: constant.SettingTypeJSON, Description: "视频评论开关（按端）",
		})
	}
	if f.CommentReview != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureCommentReview, Group: constant.SettingGroupFeature,
			Value:       service.MarshalPlatformFlags(*f.CommentReview),
			SettingType: constant.SettingTypeJSON, Description: "评论是否需要审核（按端）",
		})
	}
	if f.RatingEnabled != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingFeatureRatingEnabled, Group: constant.SettingGroupFeature,
			Value:       service.MarshalPlatformFlags(*f.RatingEnabled),
			SettingType: constant.SettingTypeJSON, Description: "视频评分开关（按端）",
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

func (s *settingsService) parseStorageUpdate(ctx context.Context, data json.RawMessage) ([]repository.SettingUpsert, error) {
	var req admindto.UpdateStorageSettings
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
	}
	if err := validator.Validate(&req); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, err.Error())
	}
	if req.Provider == nil && req.Aliyun == nil && req.Tencent == nil && req.Qiniu == nil {
		return nil, nil
	}

	current, err := s.shared.LoadMapByGroup(ctx, constant.SettingGroupStorage)
	if err != nil {
		return nil, err
	}

	aliyun := service.ParseStorageProviderRaw(service.StrVal(current, constant.SettingStorageAliyun))
	tencent := service.ParseStorageProviderRaw(service.StrVal(current, constant.SettingStorageTencent))
	qiniu := service.ParseStorageProviderRaw(service.StrVal(current, constant.SettingStorageQiniu))
	provider := service.StrVal(current, constant.SettingStorageProvider)
	if provider == "" {
		provider = constant.StorageProviderNone
	}

	merge := func(dst *service.StorageProviderRaw, src *admindto.UpdateStorageProviderConfig) error {
		if src == nil {
			return nil
		}
		if src.Bucket != nil {
			dst.Bucket = strings.TrimSpace(*src.Bucket)
		}
		if src.Region != nil {
			dst.Region = strings.TrimSpace(*src.Region)
		}
		if src.Endpoint != nil {
			dst.Endpoint = strings.TrimSpace(*src.Endpoint)
		}
		if src.AccessKey != nil && strings.TrimSpace(*src.AccessKey) != "" {
			dst.AccessKey = strings.TrimSpace(*src.AccessKey)
		}
		if src.SecretKey != nil && strings.TrimSpace(*src.SecretKey) != "" {
			dst.SecretKey = strings.TrimSpace(*src.SecretKey)
		}
		if src.CDNDomain != nil {
			cdn, normErr := objectstorage.NormalizeCDNDomain(*src.CDNDomain)
			if normErr != nil {
				return fmt.Errorf("加速域名格式无效")
			}
			dst.CDNDomain = cdn
		}
		return nil
	}

	if err := merge(&aliyun, req.Aliyun); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "阿里云: "+err.Error())
	}
	if err := merge(&tencent, req.Tencent); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "腾讯云: "+err.Error())
	}
	if err := merge(&qiniu, req.Qiniu); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "七牛云: "+err.Error())
	}

	if req.Provider != nil {
		p, normErr := service.NormalizeStorageProviderName(*req.Provider)
		if normErr != nil {
			return nil, errcode.WithMessage(errcode.ParamError, normErr.Error())
		}
		provider = p
	}

	// Validate each vendor that has any fields; enabled vendor must be complete.
	if err := service.ValidateStorageProviderConfig(constant.StorageProviderAliyun, aliyun, provider == constant.StorageProviderAliyun); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "阿里云: "+err.Error())
	}
	if err := service.ValidateStorageProviderConfig(constant.StorageProviderTencent, tencent, provider == constant.StorageProviderTencent); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "腾讯云: "+err.Error())
	}
	if err := service.ValidateStorageProviderConfig(constant.StorageProviderQiniu, qiniu, provider == constant.StorageProviderQiniu); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "七牛云: "+err.Error())
	}

	var upserts []repository.SettingUpsert
	if req.Provider != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingStorageProvider, Group: constant.SettingGroupStorage, Value: provider,
			SettingType: constant.SettingTypeString, Description: "当前启用的云存储厂商",
		})
	}
	if req.Aliyun != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingStorageAliyun, Group: constant.SettingGroupStorage,
			Value:       service.MarshalStorageProviderRaw(aliyun),
			SettingType: constant.SettingTypeJSON, Description: "阿里云 OSS 配置 JSON",
		})
	}
	if req.Tencent != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingStorageTencent, Group: constant.SettingGroupStorage,
			Value:       service.MarshalStorageProviderRaw(tencent),
			SettingType: constant.SettingTypeJSON, Description: "腾讯云 COS 配置 JSON",
		})
	}
	if req.Qiniu != nil {
		upserts = append(upserts, repository.SettingUpsert{
			Key: constant.SettingStorageQiniu, Group: constant.SettingGroupStorage,
			Value:       service.MarshalStorageProviderRaw(qiniu),
			SettingType: constant.SettingTypeJSON, Description: "七牛云 Kodo 配置 JSON",
		})
	}
	return upserts, nil
}

package admin

import (
	"context"
	"encoding/json"
	"strconv"
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

func (s *settingsService) liveSettingsView(ctx context.Context, m map[string]model.SystemSettings) (any, error) {
	seo, err := s.shared.LoadMapByGroup(ctx, constant.SettingGroupSEO)
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(strings.TrimSpace(service.StrVal(seo, constant.SettingSEOPublicBaseURL)), "/")
	return service.MapToLiveSettings(m, base), nil
}

func (s *settingsService) parseLiveUpdate(ctx context.Context, data json.RawMessage) ([]repository.SettingUpsert, error) {
	var req admindto.UpdateLiveSettings
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的设置数据")
	}
	if err := validator.Validate(&req); err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, err.Error())
	}
	if liveUpdateEmpty(&req) {
		return nil, nil
	}
	n, err := s.live.CountLivingRooms(ctx)
	if err != nil {
		s.log.Error("settings: count living rooms failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if n > 0 {
		return nil, errcode.WithMessage(errcode.LiveRoomLiving, "有房间正在直播，请先停播后再修改推流配置")
	}

	currentMap, err := s.shared.LoadMapByGroup(ctx, constant.SettingGroupLive)
	if err != nil {
		return nil, err
	}
	cfg := mergeLiveSettings(service.ParseLiveStreamConfig(currentMap), &req)
	if err := cfg.ValidateForSave(); err != nil {
		return nil, err
	}
	return liveUpserts(cfg)
}

func liveUpdateEmpty(req *admindto.UpdateLiveSettings) bool {
	return req.Provider == nil && req.AuthEnabled == nil && req.PushTTLSeconds == nil && req.PlayTTLSeconds == nil &&
		req.SignSecret == nil && req.HookSecret == nil && req.SRSAPIBase == nil &&
		req.PushServer == nil && req.PlayHLSTemplate == nil && req.PlayFLVTemplate == nil &&
		req.Aliyun == nil && req.Qiniu == nil
}

func mergeLiveSettings(cfg service.LiveStreamConfig, req *admindto.UpdateLiveSettings) service.LiveStreamConfig {
	if req.Provider != nil {
		cfg.Provider = strings.TrimSpace(*req.Provider)
	}
	if req.AuthEnabled != nil {
		cfg.AuthEnabled = *req.AuthEnabled
	}
	if req.PushTTLSeconds != nil {
		cfg.PushTTLSeconds = *req.PushTTLSeconds
	}
	if req.PlayTTLSeconds != nil {
		cfg.PlayTTLSeconds = *req.PlayTTLSeconds
	}
	keepSecret(&cfg.SignSecret, req.SignSecret)
	keepSecret(&cfg.HookSecret, req.HookSecret)
	if req.SRSAPIBase != nil {
		cfg.SRSAPIBase = strings.TrimRight(strings.TrimSpace(*req.SRSAPIBase), "/")
	}
	if req.PushServer != nil {
		cfg.PushServer = strings.TrimSpace(*req.PushServer)
	}
	if req.PlayHLSTemplate != nil {
		cfg.PlayHLSTemplate = strings.TrimSpace(*req.PlayHLSTemplate)
	}
	if req.PlayFLVTemplate != nil {
		cfg.PlayFLVTemplate = strings.TrimSpace(*req.PlayFLVTemplate)
	}
	if req.Aliyun != nil {
		src := req.Aliyun
		keepString(&cfg.Aliyun.PushDomain, src.PushDomain)
		keepString(&cfg.Aliyun.PlayDomain, src.PlayDomain)
		keepString(&cfg.Aliyun.AppName, src.AppName)
		keepString(&cfg.Aliyun.Region, src.Region)
		keepSecret(&cfg.Aliyun.PushAuthKey, src.PushAuthKey)
		keepSecret(&cfg.Aliyun.PlayAuthKey, src.PlayAuthKey)
		keepSecret(&cfg.Aliyun.AccessKeyID, src.AccessKeyID)
		keepSecret(&cfg.Aliyun.AccessKeySecret, src.AccessKeySecret)
	}
	if cfg.Aliyun.Region == "" {
		cfg.Aliyun.Region = constant.LiveAliyunRegionDefault
	}
	if req.Qiniu != nil {
		src := req.Qiniu
		keepString(&cfg.Qiniu.Hub, src.Hub)
		keepString(&cfg.Qiniu.PushDomain, src.PushDomain)
		keepString(&cfg.Qiniu.PlayDomain, src.PlayDomain)
		keepSecret(&cfg.Qiniu.AccessKey, src.AccessKey)
		keepSecret(&cfg.Qiniu.SecretKey, src.SecretKey)
	}
	if cfg.Provider == "" {
		cfg.Provider = constant.LiveProviderSelf
	}
	return cfg
}

func liveUpserts(cfg service.LiveStreamConfig) ([]repository.SettingUpsert, error) {
	aliyun, err := service.MarshalLiveAliyun(cfg.Aliyun)
	if err != nil {
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	qiniu, err := service.MarshalLiveQiniu(cfg.Qiniu)
	if err != nil {
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	auth := "0"
	if cfg.AuthEnabled {
		auth = "1"
	}
	rows := []repository.SettingUpsert{
		{Key: constant.SettingLiveProvider, Group: constant.SettingGroupLive, Value: cfg.ProviderName(), SettingType: constant.SettingTypeString, Description: "直播厂商 self/aliyun/qiniu"},
		{Key: constant.SettingLiveAuthEnabled, Group: constant.SettingGroupLive, Value: auth, SettingType: constant.SettingTypeBoolean, Description: "自建推流播放鉴权开关"},
		{Key: constant.SettingLivePushTTLSeconds, Group: constant.SettingGroupLive, Value: strconv.Itoa(cfg.PushTTLSeconds), SettingType: constant.SettingTypeNumber, Description: "推流签名有效秒数"},
		{Key: constant.SettingLivePlayTTLSeconds, Group: constant.SettingGroupLive, Value: strconv.Itoa(cfg.PlayTTLSeconds), SettingType: constant.SettingTypeNumber, Description: "播放签名有效秒数"},
		{Key: constant.SettingLiveSignSecret, Group: constant.SettingGroupLive, Value: cfg.SignSecret, SettingType: constant.SettingTypeString, Description: "自建 HMAC 签名密钥"},
		{Key: constant.SettingLiveHookSecret, Group: constant.SettingGroupLive, Value: cfg.HookSecret, SettingType: constant.SettingTypeString, Description: "SRS 回调密钥"},
		{Key: constant.SettingLiveSRSAPIBase, Group: constant.SettingGroupLive, Value: cfg.SRSAPIBase, SettingType: constant.SettingTypeString, Description: "SRS HTTP API 根地址"},
		{Key: constant.SettingLivePushServer, Group: constant.SettingGroupLive, Value: cfg.PushServer, SettingType: constant.SettingTypeString, Description: "自建 OBS 推流服务器地址"},
		{Key: constant.SettingLivePlayHLSTemplate, Group: constant.SettingGroupLive, Value: cfg.PlayHLSTemplate, SettingType: constant.SettingTypeString, Description: "自建 HLS 播放地址模板（含 {stream_key}）"},
		{Key: constant.SettingLivePlayFLVTemplate, Group: constant.SettingGroupLive, Value: cfg.PlayFLVTemplate, SettingType: constant.SettingTypeString, Description: "自建 FLV 播放地址模板（含 {stream_key}，可空）"},
		{Key: constant.SettingLiveAliyun, Group: constant.SettingGroupLive, Value: aliyun, SettingType: constant.SettingTypeJSON, Description: "阿里云直播配置 JSON"},
		{Key: constant.SettingLiveQiniu, Group: constant.SettingGroupLive, Value: qiniu, SettingType: constant.SettingTypeJSON, Description: "七牛直播配置 JSON"},
	}
	return rows, nil
}

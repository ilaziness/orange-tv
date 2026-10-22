package service

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
)

const (
	liveSRSHookPath  = "/api/open/v1/live/hooks/srs"
	livePlayAuthPath = "/api/open/v1/live/auth/play"
)

type liveAliyunRaw struct {
	PushDomain      string `json:"push_domain"`
	PlayDomain      string `json:"play_domain"`
	AppName         string `json:"app_name"`
	Region          string `json:"region"`
	PushAuthKey     string `json:"push_auth_key"`
	PlayAuthKey     string `json:"play_auth_key"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

type liveQiniuRaw struct {
	Hub        string `json:"hub"`
	PushDomain string `json:"push_domain"`
	PlayDomain string `json:"play_domain"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
}

func parseLiveAliyun(raw string) LiveAliyunConfig {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return LiveAliyunConfig{Region: constant.LiveAliyunRegionDefault}
	}
	var v liveAliyunRaw
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return LiveAliyunConfig{Region: constant.LiveAliyunRegionDefault}
	}
	region := strings.TrimSpace(v.Region)
	if region == "" {
		region = constant.LiveAliyunRegionDefault
	}
	return LiveAliyunConfig{
		PushDomain:      strings.TrimSpace(v.PushDomain),
		PlayDomain:      strings.TrimSpace(v.PlayDomain),
		AppName:         strings.TrimSpace(v.AppName),
		Region:          region,
		PushAuthKey:     strings.TrimSpace(v.PushAuthKey),
		PlayAuthKey:     strings.TrimSpace(v.PlayAuthKey),
		AccessKeyID:     strings.TrimSpace(v.AccessKeyID),
		AccessKeySecret: strings.TrimSpace(v.AccessKeySecret),
	}
}

func parseLiveQiniu(raw string) LiveQiniuConfig {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return LiveQiniuConfig{}
	}
	var v liveQiniuRaw
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return LiveQiniuConfig{}
	}
	return LiveQiniuConfig{
		Hub:        strings.TrimSpace(v.Hub),
		PushDomain: strings.TrimSpace(v.PushDomain),
		PlayDomain: strings.TrimSpace(v.PlayDomain),
		AccessKey:  strings.TrimSpace(v.AccessKey),
		SecretKey:  strings.TrimSpace(v.SecretKey),
	}
}

// MarshalLiveAliyun encodes Aliyun credentials for system_settings.
func MarshalLiveAliyun(cfg LiveAliyunConfig) (string, error) {
	b, err := json.Marshal(liveAliyunRaw(cfg))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalLiveQiniu encodes Qiniu credentials for system_settings.
func MarshalLiveQiniu(cfg LiveQiniuConfig) (string, error) {
	// #nosec G117 -- access_key and secret_key are stored in system_settings on purpose
	b, err := json.Marshal(liveQiniuRaw(cfg))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func maskLiveSettings(cfg LiveStreamConfig, publicBase string) dto.LiveSettings {
	publicBase = strings.TrimRight(strings.TrimSpace(publicBase), "/")
	return dto.LiveSettings{
		Provider:             cfg.ProviderName(),
		AuthEnabled:          cfg.AuthEnabled,
		PushTTLSeconds:       cfg.PushTTLSeconds,
		PlayTTLSeconds:       cfg.PlayTTLSeconds,
		SignSecretConfigured: cfg.SignSecret != "",
		SRSAPIBase:           cfg.SRSAPIBase,
		SRSHookURL:           liveHookURL(publicBase, liveSRSHookPath, cfg.HookSecret),
		PlayAuthURL:          liveHookURL(publicBase, livePlayAuthPath, cfg.HookSecret),
		PushServer:           cfg.PushServer,
		PlayHLSTemplate:      cfg.PlayHLSTemplate,
		PlayFLVTemplate:      cfg.PlayFLVTemplate,
		Aliyun: dto.LiveAliyunSettings{
			PushDomain:                cfg.Aliyun.PushDomain,
			PlayDomain:                cfg.Aliyun.PlayDomain,
			AppName:                   cfg.Aliyun.AppName,
			Region:                    cfg.Aliyun.Region,
			PushAuthKeyConfigured:     cfg.Aliyun.PushAuthKey != "",
			PlayAuthKeyConfigured:     cfg.Aliyun.PlayAuthKey != "",
			AccessKeyIDConfigured:     cfg.Aliyun.AccessKeyID != "",
			AccessKeySecretConfigured: cfg.Aliyun.AccessKeySecret != "",
		},
		Qiniu: dto.LiveQiniuSettings{
			Hub:              cfg.Qiniu.Hub,
			PushDomain:       cfg.Qiniu.PushDomain,
			PlayDomain:       cfg.Qiniu.PlayDomain,
			AccessConfigured: cfg.Qiniu.AccessKey != "",
			SecretConfigured: cfg.Qiniu.SecretKey != "",
		},
	}
}

func liveHookURL(publicBase, path, hookSecret string) string {
	if strings.TrimSpace(hookSecret) == "" {
		return ""
	}
	q := url.Values{}
	q.Set("hook_secret", hookSecret)
	rel := path + "?" + q.Encode()
	if publicBase == "" {
		return rel
	}
	return publicBase + rel
}

// ValidateForSave checks a settings write. Self with auth off may be incomplete.
// Aliyun and Qiniu must be fully configured. Self auth on must include secrets,
// an SRS API base, a push app path, and one play template.
func (c LiveStreamConfig) ValidateForSave() error {
	if err := c.validateTTL(); err != nil {
		return err
	}
	if c.SignSecret != "" && c.HookSecret != "" && c.SignSecret == c.HookSecret {
		return errcode.WithMessage(errcode.ParamError, "签名密钥与回调密钥不能相同")
	}
	if err := validateLivePlayTemplate(c.PlayHLSTemplate); err != nil {
		return err
	}
	if err := validateLivePlayTemplate(c.PlayFLVTemplate); err != nil {
		return err
	}
	if c.PushServer != "" {
		if _, err := parsePushServer(c.PushServer); err != nil {
			return errcode.WithMessage(errcode.ParamError, "推流服务器须为合法 URL")
		}
	}
	if c.SRSAPIBase != "" {
		if err := validateHTTPBase(c.SRSAPIBase); err != nil {
			return errcode.WithMessage(errcode.ParamError, "SRS API 地址须为 http 或 https")
		}
	}
	switch c.ProviderName() {
	case constant.LiveProviderSelf:
		if !c.AuthEnabled {
			return nil
		}
		return c.validateSelf()
	case constant.LiveProviderAliyun:
		return c.validateAliyun()
	case constant.LiveProviderQiniu:
		return c.validateQiniu()
	default:
		return errcode.WithMessage(errcode.SettingInvalid, "直播厂商无效")
	}
}

// Validate checks the active provider can issue addresses.
// Self with auth disabled only needs the existing push server and one play template.
func (c LiveStreamConfig) Validate() error {
	switch c.ProviderName() {
	case constant.LiveProviderSelf:
		return c.validateSelf()
	case constant.LiveProviderAliyun:
		return c.validateAliyun()
	case constant.LiveProviderQiniu:
		return c.validateQiniu()
	default:
		return errcode.WithMessage(errcode.SettingInvalid, "直播厂商无效")
	}
}

func (c LiveStreamConfig) pushTTL() int {
	if c.PushTTLSeconds <= 0 {
		return constant.LivePushTTLDefault
	}
	return c.PushTTLSeconds
}

func (c LiveStreamConfig) playTTL() int {
	if c.PlayTTLSeconds <= 0 {
		return constant.LivePlayTTLDefault
	}
	return c.PlayTTLSeconds
}

func (c LiveStreamConfig) validateTTL() error {
	if c.pushTTL() < constant.LiveTTLMinSeconds || c.pushTTL() > constant.LiveTTLMaxSeconds {
		return errcode.WithMessage(errcode.ParamError, "推流签名有效期须在 600 到 86400 秒之间")
	}
	if c.playTTL() < constant.LiveTTLMinSeconds || c.playTTL() > constant.LiveTTLMaxSeconds {
		return errcode.WithMessage(errcode.ParamError, "播放签名有效期须在 600 到 86400 秒之间")
	}
	return nil
}

func (c LiveStreamConfig) validateSelf() error {
	if err := c.validateTTL(); err != nil {
		return err
	}
	if c.SignSecret != "" && c.HookSecret != "" && c.SignSecret == c.HookSecret {
		return errcode.WithMessage(errcode.ParamError, "签名密钥与回调密钥不能相同")
	}
	if err := validateLivePlayTemplate(c.PlayHLSTemplate); err != nil {
		return err
	}
	if err := validateLivePlayTemplate(c.PlayFLVTemplate); err != nil {
		return err
	}
	if c.PushServer != "" {
		if _, err := parsePushServer(c.PushServer); err != nil {
			return errcode.WithMessage(errcode.ParamError, "推流服务器须为合法 URL")
		}
	}
	if !c.AuthEnabled {
		if c.PushServer == "" || (c.PlayHLSTemplate == "" && c.PlayFLVTemplate == "") {
			return errcode.LiveConfigIncomplete
		}
		return nil
	}
	if c.SignSecret == "" || c.HookSecret == "" {
		return errcode.WithMessage(errcode.LiveConfigIncomplete, "打开自建鉴权需要签名密钥和回调密钥")
	}
	if c.SRSAPIBase == "" {
		return errcode.WithMessage(errcode.LiveConfigIncomplete, "打开自建鉴权需要填写 SRS API 地址")
	}
	if err := validateHTTPBase(c.SRSAPIBase); err != nil {
		return errcode.WithMessage(errcode.ParamError, "SRS API 地址须为 http 或 https")
	}
	app, err := pushApp(c.PushServer)
	if err != nil || app == "" {
		return errcode.WithMessage(errcode.ParamError, "打开自建鉴权时推流服务器路径须为应用名，例如 rtmp://主机/live")
	}
	if c.PlayHLSTemplate == "" && c.PlayFLVTemplate == "" {
		return errcode.LiveConfigIncomplete
	}
	return nil
}

func (c LiveStreamConfig) validateAliyun() error {
	if err := c.validateTTL(); err != nil {
		return err
	}
	a := c.Aliyun
	if err := requireDomain(a.PushDomain, "阿里云推流域名"); err != nil {
		return err
	}
	if err := requireDomain(a.PlayDomain, "阿里云播放域名"); err != nil {
		return err
	}
	if !validNameToken(a.AppName) {
		return errcode.WithMessage(errcode.ParamError, "阿里云 AppName 无效")
	}
	region := a.Region
	if region == "" {
		region = constant.LiveAliyunRegionDefault
	}
	if !validRegion(region) {
		return errcode.WithMessage(errcode.ParamError, "阿里云地域无效")
	}
	if a.PushAuthKey == "" || a.PlayAuthKey == "" || a.AccessKeyID == "" || a.AccessKeySecret == "" {
		return errcode.WithMessage(errcode.LiveConfigIncomplete, "阿里云域名和密钥必须配齐")
	}
	return nil
}

func (c LiveStreamConfig) validateQiniu() error {
	if err := c.validateTTL(); err != nil {
		return err
	}
	q := c.Qiniu
	if !validNameToken(q.Hub) {
		return errcode.WithMessage(errcode.ParamError, "七牛 Hub 无效")
	}
	if err := requireDomain(q.PushDomain, "七牛推流域名"); err != nil {
		return err
	}
	if err := requireDomain(q.PlayDomain, "七牛播放域名"); err != nil {
		return err
	}
	if q.AccessKey == "" || q.SecretKey == "" {
		return errcode.WithMessage(errcode.LiveConfigIncomplete, "七牛域名和密钥必须配齐")
	}
	return nil
}

func validateLivePlayTemplate(template string) error {
	if template == "" {
		return nil
	}
	if !strings.Contains(template, "{stream_key}") {
		return errcode.WithMessage(errcode.ParamError, "播放模板须包含 {stream_key} 占位符")
	}
	lower := strings.ToLower(template)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return errcode.WithMessage(errcode.ParamError, "播放模板须为 http 或 https 地址")
	}
	return nil
}

func validateHTTPBase(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errcode.ParamError
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errcode.ParamError
	}
	return nil
}

func parsePushServer(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, errcode.ParamError
	}
	switch strings.ToLower(u.Scheme) {
	case "rtmp", "http", "https":
		return u, nil
	default:
		return nil, errcode.ParamError
	}
}

func pushApp(pushServer string) (string, error) {
	u, err := parsePushServer(pushServer)
	if err != nil {
		return "", err
	}
	return strings.Trim(u.Path, "/"), nil
}

func requireDomain(raw, label string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(raw, "://") || strings.Contains(raw, "/") || strings.ContainsAny(raw, " ?#") {
		return errcode.WithMessage(errcode.ParamError, label+"不能包含协议或路径")
	}
	if strings.Contains(raw, "..") {
		return errcode.WithMessage(errcode.ParamError, label+"无效")
	}
	return nil
}

func validNameToken(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func validRegion(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

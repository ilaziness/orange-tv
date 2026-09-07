package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/pkg/objectstorage"
	"go.uber.org/zap"
)

// StorageProviderRaw is the JSON shape stored in system_settings for one vendor.
type StorageProviderRaw struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	CDNDomain string `json:"cdn_domain"`
}

// ParseStorageProviderRaw unmarshals provider JSON; empty/invalid becomes zero value.
func ParseStorageProviderRaw(raw string) StorageProviderRaw {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "null" {
		return StorageProviderRaw{}
	}
	var out StorageProviderRaw
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return StorageProviderRaw{}
	}
	return out
}

// MarshalStorageProviderRaw encodes provider config for storage.
func MarshalStorageProviderRaw(cfg StorageProviderRaw) string {
	b, err := json.Marshal(cfg)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// MaskStorageProvider converts raw config to API response with credentials masked.
// AccessKey/SecretKey are never returned; use access_configured / secret_configured.
func MaskStorageProvider(raw StorageProviderRaw) dto.StorageProviderConfig {
	cdn := raw.CDNDomain
	if normalized, err := objectstorage.NormalizeCDNDomain(raw.CDNDomain); err == nil {
		cdn = normalized
	}
	return dto.StorageProviderConfig{
		Bucket:           raw.Bucket,
		Region:           raw.Region,
		Endpoint:         raw.Endpoint,
		AccessKey:        "",
		SecretKey:        "",
		AccessConfigured: strings.TrimSpace(raw.AccessKey) != "",
		SecretConfigured: strings.TrimSpace(raw.SecretKey) != "",
		CDNDomain:        cdn,
	}
}

// MapToStorageSettings maps settings rows to admin StorageSettings (secrets masked).
func MapToStorageSettings(m map[string]model.SystemSettings) dto.StorageSettings {
	provider := strings.TrimSpace(StrVal(m, constant.SettingStorageProvider))
	if provider == "" {
		provider = constant.StorageProviderNone
	}
	return dto.StorageSettings{
		Provider: provider,
		Aliyun:   MaskStorageProvider(ParseStorageProviderRaw(StrVal(m, constant.SettingStorageAliyun))),
		Tencent:  MaskStorageProvider(ParseStorageProviderRaw(StrVal(m, constant.SettingStorageTencent))),
		Qiniu:    MaskStorageProvider(ParseStorageProviderRaw(StrVal(m, constant.SettingStorageQiniu))),
	}
}

// NormalizeStorageProviderName validates provider enum.
func NormalizeStorageProviderName(raw string) (string, error) {
	p := strings.ToLower(strings.TrimSpace(raw))
	switch p {
	case "", constant.StorageProviderNone:
		return constant.StorageProviderNone, nil
	case constant.StorageProviderAliyun, constant.StorageProviderTencent, constant.StorageProviderQiniu:
		return p, nil
	default:
		return "", fmt.Errorf("无效的云存储厂商")
	}
}

// ValidateStorageProviderConfig ensures required fields when saving/enabling a vendor.
// provider is the vendor key (aliyun/tencent/qiniu) for region rules; empty skips vendor-specific checks.
// When requireComplete is true (enabling that vendor), access/secret/cdn/bucket are required.
func ValidateStorageProviderConfig(provider string, cfg StorageProviderRaw, requireComplete bool) error {
	cdn, err := objectstorage.NormalizeCDNDomain(cfg.CDNDomain)
	if err != nil {
		return fmt.Errorf("加速域名格式无效：%w", err)
	}
	cfg.CDNDomain = cdn

	hasAny := strings.TrimSpace(cfg.Bucket) != "" ||
		strings.TrimSpace(cfg.AccessKey) != "" ||
		strings.TrimSpace(cfg.SecretKey) != "" ||
		strings.TrimSpace(cfg.CDNDomain) != "" ||
		strings.TrimSpace(cfg.Region) != "" ||
		strings.TrimSpace(cfg.Endpoint) != ""

	if !hasAny && !requireComplete {
		return nil
	}

	if strings.TrimSpace(cfg.CDNDomain) == "" {
		return fmt.Errorf("加速域名不能为空")
	}
	if requireComplete || hasAny {
		if strings.TrimSpace(cfg.Bucket) == "" {
			return fmt.Errorf("Bucket 不能为空")
		}
		if strings.TrimSpace(cfg.AccessKey) == "" {
			return fmt.Errorf("AccessKey 不能为空")
		}
		if strings.TrimSpace(cfg.SecretKey) == "" {
			return fmt.Errorf("SecretKey 不能为空")
		}
		switch provider {
		case constant.StorageProviderAliyun:
			if strings.TrimSpace(cfg.Region) == "" && strings.TrimSpace(cfg.Endpoint) == "" {
				return fmt.Errorf("Region 或 Endpoint 至少填写一项")
			}
		case constant.StorageProviderTencent:
			if strings.TrimSpace(cfg.Region) == "" {
				return fmt.Errorf("Region 不能为空")
			}
		}
	}
	return nil
}

// StorageResolver builds the currently enabled object storage driver from settings.
type StorageResolver interface {
	// Resolve returns the active Storage or StorageNotEnabled / StorageConfigInvalid.
	Resolve(ctx context.Context) (objectstorage.Storage, error)
	// ResolveProvider builds a driver for a specific vendor using that vendor's saved config
	// (does not require it to be the currently enabled provider).
	ResolveProvider(ctx context.Context, provider string) (objectstorage.Storage, error)
	// Provider returns the configured provider name (none/aliyun/...).
	Provider(ctx context.Context) (string, error)
}

type storageResolver struct {
	settings SettingsService
	log      *zap.Logger
}

// NewStorageResolver creates a StorageResolver.
func NewStorageResolver(settings SettingsService, log *zap.Logger) StorageResolver {
	if log == nil {
		log = zap.NewNop()
	}
	return &storageResolver{settings: settings, log: log}
}

func (r *storageResolver) loadMap(ctx context.Context) (map[string]model.SystemSettings, error) {
	return r.settings.LoadMapByGroup(ctx, constant.SettingGroupStorage)
}

func (r *storageResolver) Provider(ctx context.Context) (string, error) {
	m, err := r.loadMap(ctx)
	if err != nil {
		return "", err
	}
	p, err := NormalizeStorageProviderName(StrVal(m, constant.SettingStorageProvider))
	if err != nil {
		return "", errcode.WithMessage(errcode.StorageConfigInvalid, err.Error())
	}
	return p, nil
}

func (r *storageResolver) Resolve(ctx context.Context) (objectstorage.Storage, error) {
	m, err := r.loadMap(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := NormalizeStorageProviderName(StrVal(m, constant.SettingStorageProvider))
	if err != nil {
		return nil, errcode.WithMessage(errcode.StorageConfigInvalid, err.Error())
	}
	if provider == constant.StorageProviderNone {
		return nil, errcode.StorageNotEnabled
	}
	return r.buildFromMap(m, provider)
}

func (r *storageResolver) ResolveProvider(ctx context.Context, provider string) (objectstorage.Storage, error) {
	p, err := NormalizeStorageProviderName(provider)
	if err != nil || p == constant.StorageProviderNone {
		return nil, errcode.WithMessage(errcode.StorageConfigInvalid, "无效的云存储厂商")
	}
	m, err := r.loadMap(ctx)
	if err != nil {
		return nil, err
	}
	return r.buildFromMap(m, p)
}

func (r *storageResolver) buildFromMap(m map[string]model.SystemSettings, provider string) (objectstorage.Storage, error) {
	var raw StorageProviderRaw
	switch provider {
	case constant.StorageProviderAliyun:
		raw = ParseStorageProviderRaw(StrVal(m, constant.SettingStorageAliyun))
	case constant.StorageProviderTencent:
		raw = ParseStorageProviderRaw(StrVal(m, constant.SettingStorageTencent))
	case constant.StorageProviderQiniu:
		raw = ParseStorageProviderRaw(StrVal(m, constant.SettingStorageQiniu))
	default:
		return nil, errcode.WithMessage(errcode.StorageConfigInvalid, "无效的云存储厂商")
	}

	if err := ValidateStorageProviderConfig(provider, raw, true); err != nil {
		return nil, errcode.WithMessage(errcode.StorageConfigInvalid, err.Error())
	}
	cfg := objectstorage.Config{
		Bucket:    strings.TrimSpace(raw.Bucket),
		Region:    strings.TrimSpace(raw.Region),
		Endpoint:  strings.TrimSpace(raw.Endpoint),
		AccessKey: strings.TrimSpace(raw.AccessKey),
		SecretKey: strings.TrimSpace(raw.SecretKey),
		CDNDomain: strings.TrimSpace(raw.CDNDomain),
	}
	store, err := objectstorage.New(provider, cfg)
	if err != nil {
		r.log.Error("storage: create driver failed", zap.String("provider", provider), zap.Error(err))
		return nil, errcode.WithMessage(errcode.StorageConfigInvalid, err.Error())
	}
	return store, nil
}

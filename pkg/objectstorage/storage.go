// Package objectstorage provides cloud object storage drivers (Aliyun OSS,
// Tencent COS, Qiniu Kodo). It is intentionally free of application business
// logic and may be reused across projects.
package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Provider names.
const (
	ProviderAliyun  = "aliyun"
	ProviderTencent = "tencent"
	ProviderQiniu   = "qiniu"
)

// Storage is the cloud object storage abstraction.
// PublicURL must always return the CDN (acceleration) URL, never the origin endpoint.
type Storage interface {
	Name() string
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

// Config holds common credentials and CDN settings shared by all providers.
type Config struct {
	// Bucket is the storage bucket / space name.
	Bucket string
	// Region is the provider region (e.g. oss-cn-hangzhou, ap-guangzhou).
	Region string
	// Endpoint is an optional custom API endpoint (Aliyun/Qiniu may use it).
	Endpoint string
	// AccessKey is the access key / secret id.
	AccessKey string
	// SecretKey is the secret key / secret key.
	SecretKey string
	// CDNDomain is the required HTTPS acceleration domain without trailing slash
	// (e.g. https://cdn.example.com). PublicURL always uses this domain.
	CDNDomain string
}

// Validate checks required fields for constructing a driver.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Bucket) == "" {
		return fmt.Errorf("objectstorage: bucket is required")
	}
	if strings.TrimSpace(c.AccessKey) == "" {
		return fmt.Errorf("objectstorage: access key is required")
	}
	if strings.TrimSpace(c.SecretKey) == "" {
		return fmt.Errorf("objectstorage: secret key is required")
	}
	cdn, err := NormalizeCDNDomain(c.CDNDomain)
	if err != nil {
		return err
	}
	if cdn == "" {
		return fmt.Errorf("objectstorage: cdn domain is required")
	}
	return nil
}

// NormalizeCDNDomain validates and normalizes an HTTPS CDN domain (no trailing slash).
// Empty input returns ("", nil) so callers can distinguish "missing" vs "invalid".
func NormalizeCDNDomain(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	u, err := url.Parse(v)
	if err != nil {
		return "", fmt.Errorf("objectstorage: invalid cdn domain")
	}
	if u.Scheme != "https" {
		return "", fmt.Errorf("objectstorage: cdn domain must use https")
	}
	if u.Host == "" {
		return "", fmt.Errorf("objectstorage: cdn domain missing host")
	}
	if u.User != nil {
		return "", fmt.Errorf("objectstorage: cdn domain must not contain credentials")
	}
	u.Fragment = ""
	u.RawQuery = ""
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "/" {
		u.Path = ""
	}
	if u.Path != "" {
		return "", fmt.Errorf("objectstorage: cdn domain must not contain a path")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

// JoinPublicURL joins a normalized CDN domain with an object key.
func JoinPublicURL(cdnDomain, key string) string {
	cdn := strings.TrimRight(strings.TrimSpace(cdnDomain), "/")
	k := strings.TrimLeft(strings.TrimSpace(key), "/")
	if cdn == "" {
		return ""
	}
	if k == "" {
		return cdn
	}
	return cdn + "/" + k
}

// New creates a Storage for the given provider name.
func New(provider string, cfg Config) (Storage, error) {
	cdn, err := NormalizeCDNDomain(cfg.CDNDomain)
	if err != nil {
		return nil, err
	}
	cfg.CDNDomain = cdn
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderAliyun:
		return NewAliyun(cfg)
	case ProviderTencent:
		return NewTencent(cfg)
	case ProviderQiniu:
		return NewQiniu(cfg)
	default:
		return nil, fmt.Errorf("objectstorage: unsupported provider %q", provider)
	}
}

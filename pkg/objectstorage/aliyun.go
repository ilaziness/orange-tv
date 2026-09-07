package objectstorage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type aliyunStorage struct {
	client    *oss.Client
	bucket    *oss.Bucket
	cdnDomain string
}

// NewAliyun creates an Aliyun OSS storage driver.
// Region should be like "oss-cn-hangzhou". Endpoint may override the default
// "{region}.aliyuncs.com" when provided.
func NewAliyun(cfg Config) (Storage, error) {
	cdn, err := NormalizeCDNDomain(cfg.CDNDomain)
	if err != nil {
		return nil, err
	}
	cfg.CDNDomain = cdn
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		region := strings.TrimSpace(cfg.Region)
		if region == "" {
			return nil, fmt.Errorf("objectstorage: aliyun region or endpoint is required")
		}
		if !strings.HasPrefix(region, "oss-") {
			region = "oss-" + region
		}
		endpoint = region + ".aliyuncs.com"
	}
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := oss.New(endpoint, cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("objectstorage: create aliyun client: %w", err)
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("objectstorage: open aliyun bucket: %w", err)
	}
	return &aliyunStorage{client: client, bucket: bucket, cdnDomain: cfg.CDNDomain}, nil
}

func (s *aliyunStorage) Name() string { return ProviderAliyun }

func (s *aliyunStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	opts := []oss.Option{}
	if contentType != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	if size >= 0 {
		opts = append(opts, oss.ContentLength(size))
	}
	if err := s.bucket.PutObject(key, r, opts...); err != nil {
		return fmt.Errorf("objectstorage: aliyun put: %w", err)
	}
	return nil
}

func (s *aliyunStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.bucket.DeleteObject(key); err != nil {
		return fmt.Errorf("objectstorage: aliyun delete: %w", err)
	}
	return nil
}

func (s *aliyunStorage) PublicURL(key string) string {
	return JoinPublicURL(s.cdnDomain, key)
}

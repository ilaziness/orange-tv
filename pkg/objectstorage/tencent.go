package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type tencentStorage struct {
	client    *cos.Client
	cdnDomain string
}

// NewTencent creates a Tencent COS storage driver.
// Region should be like "ap-guangzhou". Bucket is the bucket name (without APPID suffix
// is fine if already included as "name-appid").
func NewTencent(cfg Config) (Storage, error) {
	cdn, err := NormalizeCDNDomain(cfg.CDNDomain)
	if err != nil {
		return nil, err
	}
	cfg.CDNDomain = cdn
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		return nil, fmt.Errorf("objectstorage: tencent region is required")
	}
	bucketURL := strings.TrimSpace(cfg.Endpoint)
	if bucketURL == "" {
		bucketURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, region)
	}
	if !strings.HasPrefix(bucketURL, "http://") && !strings.HasPrefix(bucketURL, "https://") {
		bucketURL = "https://" + bucketURL
	}
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("objectstorage: invalid tencent endpoint: %w", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		},
	})
	return &tencentStorage{client: client, cdnDomain: cfg.CDNDomain}, nil
}

func (s *tencentStorage) Name() string { return ProviderTencent }

func (s *tencentStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{},
	}
	if contentType != "" {
		opt.ContentType = contentType
	}
	if size >= 0 {
		opt.ContentLength = size
	}
	_, err := s.client.Object.Put(ctx, key, r, opt)
	if err != nil {
		return fmt.Errorf("objectstorage: tencent put: %w", err)
	}
	return nil
}

func (s *tencentStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("objectstorage: tencent delete: %w", err)
	}
	return nil
}

func (s *tencentStorage) PublicURL(key string) string {
	return JoinPublicURL(s.cdnDomain, key)
}

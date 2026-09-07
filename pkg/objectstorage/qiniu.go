package objectstorage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type qiniuStorage struct {
	mac       *qbox.Mac
	cfg       storage.Config
	bucket    string
	cdnDomain string
}

// NewQiniu creates a Qiniu Kodo storage driver.
// Region maps to Qiniu zones (z0/z1/z2/na0/as0 or cn-east-1 style). Endpoint is unused.
func NewQiniu(cfg Config) (Storage, error) {
	cdn, err := NormalizeCDNDomain(cfg.CDNDomain)
	if err != nil {
		return nil, err
	}
	cfg.CDNDomain = cdn
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	mac := qbox.NewMac(cfg.AccessKey, cfg.SecretKey)
	scfg := storage.Config{
		UseHTTPS: true,
	}
	region := strings.TrimSpace(cfg.Region)
	if region != "" {
		scfg.Region = qiniuRegion(region)
	}
	return &qiniuStorage{
		mac:       mac,
		cfg:       scfg,
		bucket:    cfg.Bucket,
		cdnDomain: cfg.CDNDomain,
	}, nil
}

func qiniuRegion(region string) *storage.Region {
	switch strings.ToLower(region) {
	case "z0", "cn-east-1", "huadong":
		return &storage.ZoneHuadong
	case "z1", "cn-north-1", "huabei":
		return &storage.ZoneHuabei
	case "z2", "cn-south-1", "huanan":
		return &storage.ZoneHuanan
	case "na0", "us-north-1", "beimei":
		return &storage.ZoneBeimei
	case "as0", "ap-southeast-1", "xinjiapo":
		return &storage.ZoneXinjiapo
	default:
		// Let SDK auto-detect when unknown.
		return nil
	}
}

func (s *qiniuStorage) Name() string { return ProviderQiniu }

func (s *qiniuStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	putPolicy := storage.PutPolicy{Scope: s.bucket}
	upToken := putPolicy.UploadToken(s.mac)
	formUploader := storage.NewFormUploader(&s.cfg)
	ret := storage.PutRet{}
	extra := &storage.PutExtra{}
	if contentType != "" {
		extra.MimeType = contentType
	}
	if size < 0 {
		size = -1
	}
	if err := formUploader.Put(ctx, &ret, upToken, key, r, size, extra); err != nil {
		return fmt.Errorf("objectstorage: qiniu put: %w", err)
	}
	return nil
}

func (s *qiniuStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	manager := storage.NewBucketManager(s.mac, &s.cfg)
	if err := manager.Delete(s.bucket, key); err != nil {
		return fmt.Errorf("objectstorage: qiniu delete: %w", err)
	}
	return nil
}

func (s *qiniuStorage) PublicURL(key string) string {
	return JoinPublicURL(s.cdnDomain, key)
}

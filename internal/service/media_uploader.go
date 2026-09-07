package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
	_ "golang.org/x/image/webp"
)

const (
	// MaxImageUploadBytes is the max allowed image size (5 MiB).
	MaxImageUploadBytes = 5 << 20
)

// UploadImageInput is the input for MediaUploader.UploadImage.
type UploadImageInput struct {
	OwnerKind    string
	OwnerID      uint32
	OriginalName string
	Size         int64
	Reader       io.Reader
}

// MediaUploader uploads images to object storage and inserts media_assets.
type MediaUploader interface {
	UploadImage(ctx context.Context, in UploadImageInput) (*model.MediaAssets, error)
}

type mediaUploader struct {
	resolver StorageResolver
	repo     repository.MediaRepository
	log      *zap.Logger
}

// NewMediaUploader creates a MediaUploader.
func NewMediaUploader(resolver StorageResolver, repo repository.MediaRepository, log *zap.Logger) MediaUploader {
	if log == nil {
		log = zap.NewNop()
	}
	return &mediaUploader{resolver: resolver, repo: repo, log: log}
}

func (u *mediaUploader) UploadImage(ctx context.Context, in UploadImageInput) (*model.MediaAssets, error) {
	if in.OwnerKind != constant.MediaOwnerAdmin && in.OwnerKind != constant.MediaOwnerUser {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的上传方")
	}
	if in.OwnerID == 0 {
		return nil, errcode.WithMessage(errcode.ParamError, "无效的上传方 ID")
	}
	// multipart FileHeader.Size 在部分客户端可能为 0，仅作上限预检，最终以实际读取为准。
	if in.Size > MaxImageUploadBytes {
		return nil, errcode.MediaTooLarge
	}

	store, err := u.resolver.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	limited := io.LimitReader(in.Reader, MaxImageUploadBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, errcode.Wrap(errcode.ParamError, err)
	}
	if len(data) == 0 {
		return nil, errcode.WithMessage(errcode.ParamError, "文件不能为空")
	}
	if int64(len(data)) > MaxImageUploadBytes {
		return nil, errcode.MediaTooLarge
	}

	mime, ext, err := detectImage(data)
	if err != nil {
		return nil, err
	}

	var width, height uint32
	if cfg, _, decErr := image.DecodeConfig(bytes.NewReader(data)); decErr == nil {
		if cfg.Width > 0 {
			width = uint32(cfg.Width)
		}
		if cfg.Height > 0 {
			height = uint32(cfg.Height)
		}
	}

	now := time.Now()
	key := fmt.Sprintf("%s/%s/%04d/%02d/%s%s",
		in.OwnerKind, constant.MediaTypeImage, now.Year(), int(now.Month()),
		strings.ReplaceAll(uuid.NewString(), "-", ""), ext)

	if putErr := store.Put(ctx, key, bytes.NewReader(data), int64(len(data)), mime); putErr != nil {
		u.log.Error("media: put object failed", zap.String("key", key), zap.Error(putErr))
		return nil, errcode.Wrap(errcode.StorageUploadFailed, putErr)
	}

	url := store.PublicURL(key)
	asset := &model.MediaAssets{
		MediaType:    constant.MediaTypeImage,
		OwnerKind:    in.OwnerKind,
		OwnerID:      in.OwnerID,
		Provider:     store.Name(),
		ObjectKey:    key,
		URL:          url,
		Mime:         mime,
		Size:         uint64(len(data)),
		OriginalName: truncateName(in.OriginalName, 255),
		Width:        width,
		Height:       height,
	}
	if createErr := u.repo.Create(ctx, asset); createErr != nil {
		u.log.Error("media: insert asset failed", zap.String("key", key), zap.Error(createErr))
		_ = store.Delete(ctx, key)
		return nil, errcode.Wrap(errcode.DatabaseError, createErr)
	}
	return asset, nil
}

func detectImage(data []byte) (mime, ext string, err error) {
	// 仅依据内容嗅探 / 魔数判定类型，不信任客户端 Content-Type 或扩展名。
	sniffed := strings.ToLower(http.DetectContentType(data))
	if i := strings.Index(sniffed, ";"); i >= 0 {
		sniffed = strings.TrimSpace(sniffed[:i])
	}

	switch sniffed {
	case "image/jpeg":
		return "image/jpeg", ".jpg", nil
	case "image/png":
		return "image/png", ".png", nil
	case "image/gif":
		return "image/gif", ".gif", nil
	case "image/webp":
		return "image/webp", ".webp", nil
	}

	// DetectContentType 对 webp 常返回 application/octet-stream，回退到魔数。
	switch {
	case hasJPEGMagic(data):
		return "image/jpeg", ".jpg", nil
	case hasPNGMagic(data):
		return "image/png", ".png", nil
	case hasGIFMagic(data):
		return "image/gif", ".gif", nil
	case hasWebPMagic(data):
		return "image/webp", ".webp", nil
	}
	return "", "", errcode.MediaInvalidType
}

func hasJPEGMagic(b []byte) bool {
	return len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff
}

func hasPNGMagic(b []byte) bool {
	return len(b) >= 8 && b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4e && b[3] == 0x47
}

func hasGIFMagic(b []byte) bool {
	return len(b) >= 6 && (string(b[:6]) == "GIF87a" || string(b[:6]) == "GIF89a")
}

func hasWebPMagic(b []byte) bool {
	return len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP"
}

func truncateName(name string, max int) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "." || name == "/" || name == "\\" {
		return ""
	}
	runes := []rune(name)
	if len(runes) <= max {
		return name
	}
	return string(runes[:max])
}

// MediaAssetToDTO maps a model row to the shared DTO.
func MediaAssetToDTO(m *model.MediaAssets) dto.MediaAsset {
	if m == nil {
		return dto.MediaAsset{}
	}
	return dto.MediaAsset{
		ID:           m.ID,
		MediaType:    m.MediaType,
		OwnerKind:    m.OwnerKind,
		OwnerID:      m.OwnerID,
		Provider:     m.Provider,
		URL:          m.URL,
		Mime:         m.Mime,
		Size:         m.Size,
		OriginalName: m.OriginalName,
		Width:        m.Width,
		Height:       m.Height,
		CreatedAt:    m.CreatedAt.Format(time.RFC3339),
	}
}

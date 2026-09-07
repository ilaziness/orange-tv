package admin

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	dto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	"go.uber.org/zap"
)

// MediaService manages admin media library operations.
type MediaService interface {
	Upload(ctx context.Context, adminID uint32, fh *multipart.FileHeader) (*dto.MediaAsset, error)
	List(ctx context.Context, mediaType string, offset, limit int) ([]dto.MediaAsset, int, error)
	Delete(ctx context.Context, id uint64) error
	Ping(ctx context.Context) (*admindto.StoragePingResponse, error)
}

type mediaService struct {
	uploader service.MediaUploader
	resolver service.StorageResolver
	repo     repository.MediaRepository
	log      *zap.Logger
}

// NewMediaService creates an admin MediaService.
func NewMediaService(
	uploader service.MediaUploader,
	resolver service.StorageResolver,
	repo repository.MediaRepository,
	log *zap.Logger,
) MediaService {
	if log == nil {
		log = zap.NewNop()
	}
	return &mediaService{uploader: uploader, resolver: resolver, repo: repo, log: log}
}

func (s *mediaService) Upload(ctx context.Context, adminID uint32, fh *multipart.FileHeader) (*dto.MediaAsset, error) {
	if fh == nil {
		return nil, errcode.WithMessage(errcode.ParamError, "请选择文件")
	}
	if fh.Size > service.MaxImageUploadBytes {
		return nil, errcode.MediaTooLarge
	}
	f, err := fh.Open()
	if err != nil {
		return nil, errcode.WithMessage(errcode.ParamError, "无法读取上传文件")
	}
	defer f.Close()

	asset, err := s.uploader.UploadImage(ctx, service.UploadImageInput{
		OwnerKind:    constant.MediaOwnerAdmin,
		OwnerID:      adminID,
		OriginalName: fh.Filename,
		Size:         fh.Size,
		Reader:       f,
	})
	if err != nil {
		return nil, err
	}
	out := service.MediaAssetToDTO(asset)
	return &out, nil
}

func (s *mediaService) List(ctx context.Context, mediaType string, offset, limit int) ([]dto.MediaAsset, int, error) {
	items, total, err := s.repo.List(ctx, mediaType, offset, limit)
	if err != nil {
		s.log.Error("admin media: list failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	list := make([]dto.MediaAsset, 0, len(items))
	for i := range items {
		list = append(list, service.MediaAssetToDTO(&items[i]))
	}
	return list, total, nil
}

func (s *mediaService) Delete(ctx context.Context, id uint64) error {
	asset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("admin media: get for delete failed", zap.Uint64("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if asset == nil {
		return errcode.MediaNotFound
	}
	// 按资源原厂商配置删除云对象（即使当前启用的是另一家）。
	store, resolveErr := s.resolver.ResolveProvider(ctx, asset.Provider)
	if resolveErr != nil {
		s.log.Warn("admin media: resolve provider for delete failed, deleting DB row only",
			zap.Uint64("id", id), zap.String("provider", asset.Provider), zap.Error(resolveErr))
	} else if delErr := store.Delete(ctx, asset.ObjectKey); delErr != nil {
		s.log.Error("admin media: delete object failed", zap.Uint64("id", id), zap.String("key", asset.ObjectKey), zap.Error(delErr))
		return errcode.Wrap(errcode.MediaDeleteFailed, delErr)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("admin media: delete row failed", zap.Uint64("id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *mediaService) Ping(ctx context.Context) (*admindto.StoragePingResponse, error) {
	store, err := s.resolver.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("admin/image/ping/%d.txt", time.Now().UnixNano())
	payload := []byte("orange-tv-storage-ping")
	if putErr := store.Put(ctx, key, bytes.NewReader(payload), int64(len(payload)), "text/plain"); putErr != nil {
		s.log.Error("admin media: storage ping put failed", zap.Error(putErr))
		return nil, errcode.Wrap(errcode.StorageUploadFailed, putErr)
	}
	url := store.PublicURL(key)
	if delErr := store.Delete(ctx, key); delErr != nil {
		s.log.Warn("admin media: storage ping delete failed", zap.String("key", key), zap.Error(delErr))
	}
	return &admindto.StoragePingResponse{
		Provider: store.Name(),
		URL:      url,
		OK:       true,
	}, nil
}

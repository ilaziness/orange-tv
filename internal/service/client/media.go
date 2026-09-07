package client

import (
	"context"
	"mime/multipart"

	"github.com/ilaziness/orange-tv/internal/constant"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/service"
)

// MediaService handles client user media uploads.
type MediaService interface {
	Upload(ctx context.Context, userID uint32, fh *multipart.FileHeader) (*clientdto.UploadMediaResponse, error)
}

type mediaService struct {
	uploader service.MediaUploader
}

// NewMediaService creates a client MediaService.
func NewMediaService(uploader service.MediaUploader) MediaService {
	return &mediaService{uploader: uploader}
}

func (s *mediaService) Upload(ctx context.Context, userID uint32, fh *multipart.FileHeader) (*clientdto.UploadMediaResponse, error) {
	if userID == 0 {
		return nil, errcode.AuthFailed
	}
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
		OwnerKind:    constant.MediaOwnerUser,
		OwnerID:      userID,
		OriginalName: fh.Filename,
		Size:         fh.Size,
		Reader:       f,
	})
	if err != nil {
		return nil, err
	}
	return &clientdto.UploadMediaResponse{
		ID:        asset.ID,
		URL:       asset.URL,
		MediaType: asset.MediaType,
		Mime:      asset.Mime,
		Size:      asset.Size,
	}, nil
}

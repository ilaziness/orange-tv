package client

import (
	"mime/multipart"

	dto "github.com/ilaziness/orange-tv/internal/dto"
)

// MediaAsset is an alias for client convenience.
type MediaAsset = dto.MediaAsset

// UploadMediaForm is multipart upload request.
type UploadMediaForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UploadMediaResponse is the client upload result (subset).
type UploadMediaResponse struct {
	ID        uint64 `json:"id"`
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	Mime      string `json:"mime"`
	Size      uint64 `json:"size"`
}

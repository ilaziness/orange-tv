package admin

import (
	"mime/multipart"

	dto "github.com/ilaziness/orange-tv/internal/dto"
)

// StorageSettings is an alias for admin convenience.
type StorageSettings = dto.StorageSettings

// StorageProviderConfig is an alias for admin convenience.
type StorageProviderConfig = dto.StorageProviderConfig

// MediaAsset is an alias for admin convenience.
type MediaAsset = dto.MediaAsset

// UpdateStorageProviderConfig updates one vendor config (all optional fields).
// Empty secret_key means keep existing secret.
type UpdateStorageProviderConfig struct {
	Bucket    *string `json:"bucket" binding:"omitempty,max=128"`
	Region    *string `json:"region" binding:"omitempty,max=64"`
	Endpoint  *string `json:"endpoint" binding:"omitempty,max=255"`
	AccessKey *string `json:"access_key" binding:"omitempty,max=128"`
	SecretKey *string `json:"secret_key" binding:"omitempty,max=255"`
	CDNDomain *string `json:"cdn_domain" binding:"omitempty,max=256"`
}

// UpdateStorageSettings updates the storage group (all optional).
type UpdateStorageSettings struct {
	// 当前启用厂商
	Provider *string                      `json:"provider" binding:"omitempty,oneof=none aliyun tencent qiniu"`
	Aliyun   *UpdateStorageProviderConfig `json:"aliyun"`
	Tencent  *UpdateStorageProviderConfig `json:"tencent"`
	Qiniu    *UpdateStorageProviderConfig `json:"qiniu"`
}

// ListMediaQuery lists media assets.
type ListMediaQuery struct {
	dto.PaginationRequest
	// 媒体类型筛选（当前仅 image）
	MediaType string `form:"media_type" binding:"omitempty,oneof=image"`
}

// MediaIDURI binds /media/:id.
type MediaIDURI struct {
	ID uint64 `uri:"id" binding:"required,min=1"`
}

// UploadMediaForm is multipart upload request.
type UploadMediaForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// StoragePingResponse is the result of storage connectivity probe.
type StoragePingResponse struct {
	// 当前厂商
	Provider string `json:"provider"`
	// 探针对象加速域名地址（已删除）
	URL string `json:"url"`
	// 是否成功
	OK bool `json:"ok"`
}

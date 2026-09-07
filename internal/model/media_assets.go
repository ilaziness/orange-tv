package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// MediaAssets represents the media_assets table.
type MediaAssets struct {
	bun.BaseModel `bun:"table:media_assets,alias:ma"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 媒体类型：image（预留 video 等）
	MediaType string `bun:"media_type,notnull" json:"media_type"`
	// 上传方：admin / user
	OwnerKind string `bun:"owner_kind,notnull" json:"owner_kind"`
	// 管理员或用户 ID
	OwnerID uint32 `bun:"owner_id,notnull" json:"owner_id"`
	// 云厂商：aliyun/tencent/qiniu
	Provider string `bun:"provider,notnull" json:"provider"`
	// 对象存储 key
	ObjectKey string `bun:"object_key,notnull" json:"object_key"`
	// 加速域名公网地址
	URL string `bun:"url,notnull" json:"url"`
	// MIME 类型
	Mime string `bun:"mime,notnull" json:"mime"`
	// 文件大小（字节）
	Size uint64 `bun:"size,notnull" json:"size"`
	// 原始文件名
	OriginalName string `bun:"original_name,notnull" json:"original_name"`
	// 图片宽度（可选）
	Width uint32 `bun:"width,notnull" json:"width"`
	// 图片高度（可选）
	Height    uint32    `bun:"height,notnull" json:"height"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
}

var _ bun.BeforeAppendModelHook = (*MediaAssets)(nil)

func (m *MediaAssets) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	now := time.Now()
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = now
	}
	return nil
}

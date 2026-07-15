package model

import (
	"time"

	"github.com/uptrace/bun"
)

// LiveChannels represents the live_channels table.
type LiveChannels struct {
	bun.BaseModel `bun:"table:live_channels,alias:l"`

	ID int64 `bun:"id" json:"id"`
	// 频道名称
	Name string `bun:"name" json:"name"`
	// 频道分类
	Category string `bun:"category" json:"category"`
	// 直播流地址
	StreamURL string `bun:"stream_url" json:"stream_url"`
	// 频道Logo
	Logo string `bun:"logo" json:"logo"`
	// 频道描述
	Description *string `bun:"description" json:"description"`
	// 排序
	SortOrder int32 `bun:"sort_order" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    int8       `bun:"status" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

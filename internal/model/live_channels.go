package model

import (
	"time"

	"github.com/uptrace/bun"
)

// LiveChannels represents the live_channels table.
type LiveChannels struct {
	bun.BaseModel `bun:"table:live_channels,alias:lc"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 频道名称
	Name string `bun:"name,notnull" json:"name"`
	// 频道分类
	Category string `bun:"category,notnull" json:"category"`
	// 直播流地址
	StreamURL string `bun:"stream_url,notnull" json:"stream_url"`
	// 频道Logo
	Logo string `bun:"logo,notnull" json:"logo"`
	// 频道描述
	Description *string `bun:"description" json:"description"`
	// 排序
	SortOrder uint32 `bun:"sort_order,notnull" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    uint8      `bun:"status,notnull" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

// PlaySources represents the play_sources table.
type PlaySources struct {
	bun.BaseModel `bun:"table:play_sources,alias:ps"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 源名称（如"播放源1"、"采集源A"）
	Name string `bun:"name,notnull" json:"name"`
	// 排序
	SortOrder uint32 `bun:"sort_order,notnull" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    uint8      `bun:"status,notnull" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt         *time.Time         `bun:"deleted_at" json:"deleted_at"`
	CollectSources    []*CollectSources  `bun:"rel:has-many,join:id=play_source_id" json:"-"`
	PlayEpisodes      []*PlayEpisodes    `bun:"rel:has-many,join:id=source_id" json:"-"`
	UserPlayHistories []*UserPlayHistory `bun:"rel:has-many,join:id=play_source_id" json:"-"`
}

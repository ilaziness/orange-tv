package model

import (
	"time"

	"github.com/uptrace/bun"
)

// PlaySources represents the play_sources table.
type PlaySources struct {
	bun.BaseModel `bun:"table:play_sources,alias:p"`

	ID int64 `bun:"id" json:"id"`
	// 源名称（如"播放源1"、"采集源A"）
	Name string `bun:"name" json:"name"`
	// 排序
	SortOrder int32 `bun:"sort_order" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    int8       `bun:"status" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

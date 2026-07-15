package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Categories represents the categories table.
type Categories struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID int64 `bun:"id" json:"id"`
	// 分类名称
	Name string `bun:"name" json:"name"`
	// 父分类ID
	ParentID int64 `bun:"parent_id" json:"parent_id"`
	// 排序
	SortOrder int32 `bun:"sort_order" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    int8       `bun:"status" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

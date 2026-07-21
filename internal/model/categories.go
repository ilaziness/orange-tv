package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Categories represents the categories table.
type Categories struct {
	bun.BaseModel `bun:"table:categories,alias:ca"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 分类名称
	Name string `bun:"name,notnull" json:"name"`
	// 父分类ID
	// Relation: parent_id -> Categories(ID)
	ParentID uint64 `bun:"parent_id,notnull" json:"parent_id"`
	// 排序
	SortOrder uint32 `bun:"sort_order,notnull" json:"sort_order"`
	// 状态：1启用 0禁用
	Status uint8 `bun:"status,notnull" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
	Parent *Categories `bun:"rel:belongs-to,join:parent_id=id" json:"-"`
	SubCategories []*Categories `bun:"rel:has-many,join:id=parent_id" json:"-"`
	CollectSourceCategories []*CollectSourceCategories `bun:"rel:has-many,join:id=category_id" json:"-"`
	Videos []*Videos `bun:"rel:has-many,join:id=category_id" json:"-"`
}

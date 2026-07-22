package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CollectSourceCategories represents the collect_source_categories table.
type CollectSourceCategories struct {
	bun.BaseModel `bun:"table:collect_source_categories,alias:csc"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 采集源ID
	// Relation: source_id -> CollectSources(ID)
	SourceID uint64 `bun:"source_id,notnull,unique:uk_source_external" json:"source_id"`
	// 外部分类名称（采集源返回的分类）
	ExternalCategory string `bun:"external_category,notnull,unique:uk_source_external" json:"external_category"`
	// 系统内分类ID
	// Relation: category_id -> Categories(ID)
	CategoryID    uint64          `bun:"category_id,notnull" json:"category_id"`
	CreatedAt     *time.Time      `bun:"created_at" json:"created_at"`
	Category      *Categories     `bun:"rel:belongs-to,join:category_id=id" json:"-"`
	CollectSource *CollectSources `bun:"rel:belongs-to,join:source_id=id" json:"-"`
}

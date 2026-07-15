package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CollectSourceCategories represents the collect_source_categories table.
type CollectSourceCategories struct {
	bun.BaseModel `bun:"table:collect_source_categories,alias:c"`

	ID int64 `bun:"id" json:"id"`
	// 采集源ID
	SourceID int64 `bun:"source_id" json:"source_id"`
	// 外部分类名称（采集源返回的分类）
	ExternalCategory string `bun:"external_category" json:"external_category"`
	// 系统内分类ID
	CategoryID int64      `bun:"category_id" json:"category_id"`
	CreatedAt  *time.Time `bun:"created_at" json:"created_at"`
}

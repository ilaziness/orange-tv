package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// LiveCategories represents the live_categories table.
type LiveCategories struct {
	bun.BaseModel `bun:"table:live_categories,alias:lc"`

	ID uint32 `bun:"id,pk,autoincrement" json:"id"`
	// 分类名称
	Name string `bun:"name,notnull" json:"name"`
	// 排序，越小越前
	Sort uint32 `bun:"sort,notnull" json:"sort"`
	// 状态：1启用 0禁用
	Status    uint8     `bun:"status,notnull" json:"status"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

var _ bun.BeforeAppendModelHook = (*LiveCategories)(nil)

func (m *LiveCategories) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	now := time.Now()
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = now
		m.UpdatedAt = now
	case *bun.UpdateQuery:
		m.UpdatedAt = now
	}
	return nil
}

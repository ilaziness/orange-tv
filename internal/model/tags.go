package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Tags represents the tags table.
type Tags struct {
	bun.BaseModel `bun:"table:tags,alias:ta"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 标签名称
	Name string `bun:"name,notnull,unique" json:"name"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
	VideoTags []*VideoTags `bun:"rel:has-many,join:id=tag_id" json:"-"`
}

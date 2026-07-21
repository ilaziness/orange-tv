package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Directors represents the directors table.
type Directors struct {
	bun.BaseModel `bun:"table:directors,alias:di"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 导演名称
	Name string `bun:"name,notnull,unique" json:"name"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
	VideoDirectors []*VideoDirectors `bun:"rel:has-many,join:id=director_id" json:"-"`
}

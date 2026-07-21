package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Actors represents the actors table.
type Actors struct {
	bun.BaseModel `bun:"table:actors,alias:ac"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 演员名称
	Name      string     `bun:"name,notnull,unique" json:"name"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

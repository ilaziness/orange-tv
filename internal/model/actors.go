package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Actors represents the actors table.
type Actors struct {
	bun.BaseModel `bun:"table:actors,alias:a"`

	ID int64 `bun:"id" json:"id"`
	// 演员名称
	Name      string     `bun:"name" json:"name"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

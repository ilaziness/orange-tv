package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Directors represents the directors table.
type Directors struct {
	bun.BaseModel `bun:"table:directors,alias:d"`

	ID int64 `bun:"id" json:"id"`
	// 导演名称
	Name      string     `bun:"name" json:"name"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Banners represents the banners table.
type Banners struct {
	bun.BaseModel `bun:"table:banners,alias:b"`

	ID        int64     `bun:"id" json:"id"`
	Title     string    `bun:"title" json:"title"`
	Cover     string    `bun:"cover" json:"cover"`
	Link      string    `bun:"link" json:"link"`
	VideoID   int64     `bun:"video_id" json:"video_id"`
	Sort      int32     `bun:"sort" json:"sort"`
	Status    int8      `bun:"status" json:"status"`
	CreatedAt time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at" json:"updated_at"`
}

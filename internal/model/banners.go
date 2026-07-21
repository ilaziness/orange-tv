package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Banners represents the banners table.
type Banners struct {
	bun.BaseModel `bun:"table:banners,alias:ba"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	Title string `bun:"title,notnull" json:"title"`
	Cover string `bun:"cover,notnull" json:"cover"`
	Link string `bun:"link,notnull" json:"link"`
	// Relation: video_id -> Videos(ID)
	VideoID uint64 `bun:"video_id,notnull" json:"video_id"`
	Sort uint32 `bun:"sort,notnull" json:"sort"`
	// 1启用 0禁用
	Status uint8 `bun:"status,notnull" json:"status"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull" json:"updated_at"`
	Video *Videos `bun:"rel:belongs-to,join:video_id=id" json:"-"`
}

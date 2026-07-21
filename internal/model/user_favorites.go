package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserFavorites represents the user_favorites table.
type UserFavorites struct {
	bun.BaseModel `bun:"table:user_favorites,alias:uf"`

	ID        uint64    `bun:"id,pk,autoincrement" json:"id"`
	UserID    uint64    `bun:"user_id,notnull,unique:uk_user_video" json:"user_id"`
	VideoID   uint64    `bun:"video_id,notnull,unique:uk_user_video" json:"video_id"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
}

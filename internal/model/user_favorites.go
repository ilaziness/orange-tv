package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserFavorites represents the user_favorites table.
type UserFavorites struct {
	bun.BaseModel `bun:"table:user_favorites,alias:uf"`

	ID        int64     `bun:"id" json:"id"`
	UserID    int64     `bun:"user_id" json:"user_id"`
	VideoID   int64     `bun:"video_id" json:"video_id"`
	CreatedAt time.Time `bun:"created_at" json:"created_at"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserPlayHistory represents the user_play_history table.
type UserPlayHistory struct {
	bun.BaseModel `bun:"table:user_play_history,alias:uph"`

	ID           int64     `bun:"id" json:"id"`
	UserID       int64     `bun:"user_id" json:"user_id"`
	VideoID      int64     `bun:"video_id" json:"video_id"`
	PlaySourceID int64     `bun:"play_source_id" json:"play_source_id"`
	EpisodeID    int64     `bun:"episode_id" json:"episode_id"`
	Progress     uint32    `bun:"progress" json:"progress"`
	Duration     uint32    `bun:"duration" json:"duration"`
	LastPlayedAt time.Time `bun:"last_played_at" json:"last_played_at"`
	CreatedAt    time.Time `bun:"created_at" json:"created_at"`
}

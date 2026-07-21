package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserPlayHistory represents the user_play_history table.
type UserPlayHistory struct {
	bun.BaseModel `bun:"table:user_play_history,alias:uph"`

	ID           uint64 `bun:"id,pk,autoincrement" json:"id"`
	UserID       uint64 `bun:"user_id,notnull,unique:uk_user_video" json:"user_id"`
	VideoID      uint64 `bun:"video_id,notnull,unique:uk_user_video" json:"video_id"`
	PlaySourceID uint64 `bun:"play_source_id,notnull" json:"play_source_id"`
	EpisodeID    uint64 `bun:"episode_id,notnull" json:"episode_id"`
	// 播放进度（秒）
	Progress uint32 `bun:"progress,notnull" json:"progress"`
	// 总时长（秒）
	Duration     uint32    `bun:"duration,notnull" json:"duration"`
	LastPlayedAt time.Time `bun:"last_played_at,notnull" json:"last_played_at"`
	CreatedAt    time.Time `bun:"created_at,notnull" json:"created_at"`
}

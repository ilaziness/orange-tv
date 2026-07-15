package model

import (
	"github.com/uptrace/bun"
)

// VideoActors represents the video_actors table.
type VideoActors struct {
	bun.BaseModel `bun:"table:video_actors,alias:v"`

	ID int64 `bun:"id" json:"id"`
	// 影视ID
	VideoID int64 `bun:"video_id" json:"video_id"`
	// 演员ID
	ActorID int64 `bun:"actor_id" json:"actor_id"`
	// 角色名
	Role string `bun:"role" json:"role"`
}

package model

import (
	"github.com/uptrace/bun"
)

// VideoActors represents the video_actors table.
type VideoActors struct {
	bun.BaseModel `bun:"table:video_actors,alias:va"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 影视ID
	// Relation: video_id -> Videos(ID)
	VideoID uint64 `bun:"video_id,notnull" json:"video_id"`
	// 演员ID
	// Relation: actor_id -> Actors(ID)
	ActorID uint64 `bun:"actor_id,notnull" json:"actor_id"`
	// 角色名
	Role string `bun:"role,notnull" json:"role"`
	Actor *Actors `bun:"rel:belongs-to,join:actor_id=id" json:"-"`
	Video *Videos `bun:"rel:belongs-to,join:video_id=id" json:"-"`
}

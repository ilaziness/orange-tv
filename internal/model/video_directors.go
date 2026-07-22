package model

import (
	"github.com/uptrace/bun"
)

// VideoDirectors represents the video_directors table.
type VideoDirectors struct {
	bun.BaseModel `bun:"table:video_directors,alias:vd"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 影视ID
	// Relation: video_id -> Videos(ID)
	VideoID uint64 `bun:"video_id,notnull" json:"video_id"`
	// 导演ID
	// Relation: director_id -> Directors(ID)
	DirectorID uint64     `bun:"director_id,notnull" json:"director_id"`
	Director   *Directors `bun:"rel:belongs-to,join:director_id=id" json:"-"`
	Video      *Videos    `bun:"rel:belongs-to,join:video_id=id" json:"-"`
}

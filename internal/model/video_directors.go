package model

import (
	"github.com/uptrace/bun"
)

// VideoDirectors represents the video_directors table.
type VideoDirectors struct {
	bun.BaseModel `bun:"table:video_directors,alias:v"`

	ID int64 `bun:"id" json:"id"`
	// 影视ID
	VideoID int64 `bun:"video_id" json:"video_id"`
	// 导演ID
	DirectorID int64 `bun:"director_id" json:"director_id"`
}

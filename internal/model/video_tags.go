package model

import (
	"github.com/uptrace/bun"
)

// VideoTags represents the video_tags table.
type VideoTags struct {
	bun.BaseModel `bun:"table:video_tags,alias:v"`

	ID int64 `bun:"id" json:"id"`
	// 影视ID
	VideoID int64 `bun:"video_id" json:"video_id"`
	// 标签ID
	TagID int64 `bun:"tag_id" json:"tag_id"`
}

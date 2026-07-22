package model

import (
	"github.com/uptrace/bun"
)

// VideoTags represents the video_tags table.
type VideoTags struct {
	bun.BaseModel `bun:"table:video_tags,alias:vt"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 影视ID
	// Relation: video_id -> Videos(ID)
	VideoID uint64 `bun:"video_id,notnull" json:"video_id"`
	// 标签ID
	// Relation: tag_id -> Tags(ID)
	TagID uint64  `bun:"tag_id,notnull" json:"tag_id"`
	Tag   *Tags   `bun:"rel:belongs-to,join:tag_id=id" json:"-"`
	Video *Videos `bun:"rel:belongs-to,join:video_id=id" json:"-"`
}

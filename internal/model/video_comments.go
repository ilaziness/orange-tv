package model

import (
	"time"

	"github.com/uptrace/bun"
)

// VideoComments represents the video_comments table.
type VideoComments struct {
	bun.BaseModel `bun:"table:video_comments,alias:vc"`

	ID        int64     `bun:"id" json:"id"`
	VideoID   int64     `bun:"video_id" json:"video_id"`
	UserID    int64     `bun:"user_id" json:"user_id"`
	ParentID  int64     `bun:"parent_id" json:"parent_id"`
	Content   string    `bun:"content" json:"content"`
	Status    int8      `bun:"status" json:"status"`
	LikeCount uint32    `bun:"like_count" json:"like_count"`
	CreatedAt time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at" json:"updated_at"`
}

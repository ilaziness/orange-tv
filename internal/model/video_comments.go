package model

import (
	"time"

	"github.com/uptrace/bun"
)

// VideoComments represents the video_comments table.
type VideoComments struct {
	bun.BaseModel `bun:"table:video_comments,alias:vc"`

	ID      uint64 `bun:"id,pk,autoincrement" json:"id"`
	VideoID uint64 `bun:"video_id,notnull" json:"video_id"`
	UserID  uint64 `bun:"user_id,notnull" json:"user_id"`
	// 父评论ID，0为顶级
	ParentID uint64 `bun:"parent_id,notnull" json:"parent_id"`
	Content  string `bun:"content,notnull" json:"content"`
	// 1正常 0隐藏
	Status    uint8     `bun:"status,notnull" json:"status"`
	LikeCount uint32    `bun:"like_count,notnull" json:"like_count"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull" json:"updated_at"`
}

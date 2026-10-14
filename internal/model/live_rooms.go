package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// LiveRooms represents the live_rooms table.
type LiveRooms struct {
	bun.BaseModel `bun:"table:live_rooms,alias:lr"`

	ID uint32 `bun:"id,pk,autoincrement" json:"id"`
	// 主播用户ID
	UserID uint32 `bun:"user_id,notnull,unique" json:"user_id"`
	// 直播分类ID
	CategoryID uint32 `bun:"category_id,notnull" json:"category_id"`
	// 直播间标题
	Title string `bun:"title,notnull" json:"title"`
	// 封面URL
	Cover string `bun:"cover,notnull" json:"cover"`
	// 公告
	Notice string `bun:"notice,notnull" json:"notice"`
	// 推流密钥（明文）
	StreamKey string `bun:"stream_key,notnull,unique" json:"stream_key"`
	// 状态：0空闲 1直播中 2禁播
	Status uint8 `bun:"status,notnull" json:"status"`
	// 当前场次ID
	CurrentSessionID *uint32 `bun:"current_session_id" json:"current_session_id"`
	// 本次开播时间
	StartedAt *time.Time `bun:"started_at" json:"started_at"`
	// 最近停播时间
	EndedAt   *time.Time `bun:"ended_at" json:"ended_at"`
	CreatedAt time.Time  `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,notnull" json:"updated_at"`
}

var _ bun.BeforeAppendModelHook = (*LiveRooms)(nil)

func (m *LiveRooms) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	now := time.Now()
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = now
		m.UpdatedAt = now
	case *bun.UpdateQuery:
		m.UpdatedAt = now
	}
	return nil
}

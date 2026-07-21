package model

import (
	"time"

	"github.com/uptrace/bun"
)

// PlayEpisodes represents the play_episodes table.
type PlayEpisodes struct {
	bun.BaseModel `bun:"table:play_episodes,alias:pe"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 播放源ID
	SourceID uint64 `bun:"source_id,notnull,unique:uk_source_video_episode" json:"source_id"`
	// 影视ID
	VideoID uint64 `bun:"video_id,notnull,unique:uk_source_video_episode" json:"video_id"`
	// 集数编号
	EpisodeNumber uint32 `bun:"episode_number,notnull,unique:uk_source_video_episode" json:"episode_number"`
	// 集标题（如"第1集"）
	Title string `bun:"title,notnull" json:"title"`
	// 播放地址
	PlayURL string `bun:"play_url,notnull" json:"play_url"`
	// 清晰度
	Quality string `bun:"quality,notnull" json:"quality"`
	// 格式（hls/mp4/dash/flv）
	Format string `bun:"format,notnull" json:"format"`
	// 排序
	SortOrder uint32 `bun:"sort_order,notnull" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    uint8      `bun:"status,notnull" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

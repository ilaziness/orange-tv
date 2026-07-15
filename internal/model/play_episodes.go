package model

import (
	"time"

	"github.com/uptrace/bun"
)

// PlayEpisodes represents the play_episodes table.
type PlayEpisodes struct {
	bun.BaseModel `bun:"table:play_episodes,alias:p"`

	ID int64 `bun:"id" json:"id"`
	// 播放源ID
	SourceID int64 `bun:"source_id" json:"source_id"`
	// 影视ID
	VideoID int64 `bun:"video_id" json:"video_id"`
	// 集数编号
	EpisodeNumber int32 `bun:"episode_number" json:"episode_number"`
	// 集标题（如"第1集"）
	Title string `bun:"title" json:"title"`
	// 播放地址
	PlayURL string `bun:"play_url" json:"play_url"`
	// 清晰度
	Quality string `bun:"quality" json:"quality"`
	// 格式（hls/mp4/dash/flv）
	Format string `bun:"format" json:"format"`
	// 排序
	SortOrder int32 `bun:"sort_order" json:"sort_order"`
	// 状态：1启用 0禁用
	Status    int8       `bun:"status" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

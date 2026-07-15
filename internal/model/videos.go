package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Videos represents the videos table.
type Videos struct {
	bun.BaseModel `bun:"table:videos,alias:v"`

	ID int64 `bun:"id" json:"id"`
	// 标题
	Title string `bun:"title" json:"title"`
	// 副标题
	Subtitle string `bun:"subtitle" json:"subtitle"`
	// 描述
	Description *string `bun:"description" json:"description"`
	// 分类ID
	CategoryID int64 `bun:"category_id" json:"category_id"`
	// 上下架状态：1上架 0下架
	PublishStatus int8 `bun:"publish_status" json:"publish_status"`
	// 连载状态：1连载中 2已完结 3即将上线
	SerialStatus int8 `bun:"serial_status" json:"serial_status"`
	// 封面图
	CoverImage string `bun:"cover_image" json:"cover_image"`
	// 海报图
	PosterImage string `bun:"poster_image" json:"poster_image"`
	// 年份
	Year int32 `bun:"year" json:"year"`
	// 地区
	Region string `bun:"region" json:"region"`
	// 评分
	Rating float64 `bun:"rating" json:"rating"`
	// 播放次数
	ViewCount int32 `bun:"view_count" json:"view_count"`
	// 时长（分钟）
	Duration int32 `bun:"duration" json:"duration"`
	// 语言
	Language string `bun:"language" json:"language"`
	// 上映日期
	ReleaseDate *time.Time `bun:"release_date" json:"release_date"`
	CreatedAt   *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

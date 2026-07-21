package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Videos represents the videos table.
type Videos struct {
	bun.BaseModel `bun:"table:videos,alias:vi"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 标题
	Title string `bun:"title,notnull" json:"title"`
	// 副标题
	Subtitle string `bun:"subtitle,notnull" json:"subtitle"`
	// 描述
	Description *string `bun:"description" json:"description"`
	// 分类ID
	CategoryID uint64 `bun:"category_id,notnull" json:"category_id"`
	// 上下架状态：1上架 0下架
	PublishStatus uint8 `bun:"publish_status,notnull" json:"publish_status"`
	// 连载状态：1连载中 2已完结 3即将上线
	SerialStatus uint8 `bun:"serial_status,notnull" json:"serial_status"`
	// 封面图
	CoverImage string `bun:"cover_image,notnull" json:"cover_image"`
	// 海报图
	PosterImage string `bun:"poster_image,notnull" json:"poster_image"`
	// 年份
	Year uint32 `bun:"year,notnull" json:"year"`
	// 地区
	Region string `bun:"region,notnull" json:"region"`
	// 评分
	Rating float64 `bun:"rating,notnull" json:"rating"`
	// 播放次数
	ViewCount uint32 `bun:"view_count,notnull" json:"view_count"`
	// 时长（分钟）
	Duration uint32 `bun:"duration,notnull" json:"duration"`
	// 语言
	Language string `bun:"language,notnull" json:"language"`
	// 上映日期
	ReleaseDate *time.Time `bun:"release_date" json:"release_date"`
	CreatedAt   *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

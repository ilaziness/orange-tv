package model

import (
	"time"

	"github.com/uptrace/bun"
)

// Themes represents the themes table.
type Themes struct {
	bun.BaseModel `bun:"table:themes,alias:t"`

	ID int64 `bun:"id" json:"id"`
	// 主题名称
	Name string `bun:"name" json:"name"`
	// 主题标识
	Identifier string `bun:"identifier" json:"identifier"`
	// 版本
	Version string `bun:"version" json:"version"`
	// 作者
	Author string `bun:"author" json:"author"`
	// 描述
	Description *string `bun:"description" json:"description"`
	// 预览图
	PreviewImage string `bun:"preview_image" json:"preview_image"`
	// 主题配置（管理员覆盖后的最终配置，合并自theme.json默认值）
	Config *string `bun:"config" json:"config"`
	// 自定义CSS
	CustomCss *string `bun:"custom_css" json:"custom_css"`
	// 自定义JS
	CustomJs *string `bun:"custom_js" json:"custom_js"`
	// 是否默认
	IsDefault int8 `bun:"is_default" json:"is_default"`
	// 是否当前启用
	IsActive  int8       `bun:"is_active" json:"is_active"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

// SystemSettings represents the system_settings table.
type SystemSettings struct {
	bun.BaseModel `bun:"table:system_settings,alias:s"`

	ID int64 `bun:"id" json:"id"`
	// 设置键
	SettingKey string `bun:"setting_key" json:"setting_key"`
	// 设置值
	SettingValue *string `bun:"setting_value" json:"setting_value"`
	// 设置类型：1string 2number 3boolean 4json
	SettingType int8 `bun:"setting_type" json:"setting_type"`
	// 描述
	Description string     `bun:"description" json:"description"`
	CreatedAt   *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updated_at" json:"updated_at"`
}

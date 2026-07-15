package model

import (
	"time"

	"github.com/uptrace/bun"
)

// SystemLogs represents the system_logs table.
type SystemLogs struct {
	bun.BaseModel `bun:"table:system_logs,alias:s"`

	ID int64 `bun:"id" json:"id"`
	// 日志级别：1info 2warning 3error 4critical
	Level int8 `bun:"level" json:"level"`
	// 模块
	Module string `bun:"module" json:"module"`
	// 操作
	Action string `bun:"action" json:"action"`
	// 操作管理员ID
	AdminID int64 `bun:"admin_id" json:"admin_id"`
	// 日志内容
	Content *string `bun:"content" json:"content"`
	// IP地址
	IPAddress string     `bun:"ip_address" json:"ip_address"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
}

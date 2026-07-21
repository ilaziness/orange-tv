package model

import (
	"time"

	"github.com/uptrace/bun"
)

// SystemLogs represents the system_logs table.
type SystemLogs struct {
	bun.BaseModel `bun:"table:system_logs,alias:sl"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 日志级别：1info 2warning 3error 4critical
	Level uint8 `bun:"level,notnull" json:"level"`
	// 模块
	Module string `bun:"module,notnull" json:"module"`
	// 操作
	Action string `bun:"action,notnull" json:"action"`
	// 操作管理员ID
	AdminID uint64 `bun:"admin_id,notnull" json:"admin_id"`
	// 日志内容
	Content *string `bun:"content" json:"content"`
	// IP地址
	IPAddress string     `bun:"ip_address,notnull" json:"ip_address"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
}

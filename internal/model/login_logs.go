package model

import (
	"time"

	"github.com/uptrace/bun"
)

// LoginLogs represents the login_logs table.
type LoginLogs struct {
	bun.BaseModel `bun:"table:login_logs,alias:ll"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 用户类型：1管理员 2普通用户
	UserType uint8 `bun:"user_type,notnull" json:"user_type"`
	// 用户ID
	UserID uint64 `bun:"user_id,notnull" json:"user_id"`
	// 用户名
	Username string `bun:"username,notnull" json:"username"`
	// IP地址
	IPAddress string `bun:"ip_address,notnull" json:"ip_address"`
	// User-Agent
	UserAgent string `bun:"user_agent,notnull" json:"user_agent"`
	// 登录状态：1成功 2失败
	Status uint8 `bun:"status,notnull" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
}

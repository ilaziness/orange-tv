package model

import (
	"time"

	"github.com/uptrace/bun"
)

// LoginLogs represents the login_logs table.
type LoginLogs struct {
	bun.BaseModel `bun:"table:login_logs,alias:l"`

	ID int64 `bun:"id" json:"id"`
	// 用户类型：1管理员 2普通用户
	UserType int8 `bun:"user_type" json:"user_type"`
	// 用户ID
	UserID int64 `bun:"user_id" json:"user_id"`
	// 用户名
	Username string `bun:"username" json:"username"`
	// IP地址
	IPAddress string `bun:"ip_address" json:"ip_address"`
	// User-Agent
	UserAgent string `bun:"user_agent" json:"user_agent"`
	// 登录状态：1成功 2失败
	Status    int8       `bun:"status" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
}

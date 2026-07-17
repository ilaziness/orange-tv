package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserLoginLogs represents the user_login_logs table.
type UserLoginLogs struct {
	bun.BaseModel `bun:"table:user_login_logs,alias:ull"`

	ID        int64     `bun:"id" json:"id"`
	UserID    int64     `bun:"user_id" json:"user_id"`
	Username  string    `bun:"username" json:"username"`
	IP        string    `bun:"ip" json:"ip"`
	UserAgent string    `bun:"user_agent" json:"user_agent"`
	Status    int8      `bun:"status" json:"status"`
	Message   string    `bun:"message" json:"message"`
	CreatedAt time.Time `bun:"created_at" json:"created_at"`
}

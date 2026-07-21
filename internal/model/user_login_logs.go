package model

import (
	"time"

	"github.com/uptrace/bun"
)

// UserLoginLogs represents the user_login_logs table.
type UserLoginLogs struct {
	bun.BaseModel `bun:"table:user_login_logs,alias:ull"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// Relation: user_id -> Users(ID)
	UserID uint64 `bun:"user_id,notnull" json:"user_id"`
	Username string `bun:"username,notnull" json:"username"`
	IP string `bun:"ip,notnull" json:"ip"`
	UserAgent string `bun:"user_agent,notnull" json:"user_agent"`
	// 1成功 0失败
	Status uint8 `bun:"status,notnull" json:"status"`
	Message string `bun:"message,notnull" json:"message"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
	User *Users `bun:"rel:belongs-to,join:user_id=id" json:"-"`
}

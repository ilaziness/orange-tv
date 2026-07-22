package model

import (
	"time"

	"github.com/uptrace/bun"
)

// OnlineSessions represents the online_sessions table.
type OnlineSessions struct {
	bun.BaseModel `bun:"table:online_sessions,alias:os"`

	ID           uint64    `bun:"id,pk,autoincrement" json:"id"`
	SessionKey   string    `bun:"session_key,notnull,unique" json:"session_key"`
	IP           string    `bun:"ip,notnull" json:"ip"`
	LastActiveAt time.Time `bun:"last_active_at,notnull" json:"last_active_at"`
}

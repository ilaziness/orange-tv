package model

import (
	"time"

	"github.com/uptrace/bun"
)

// SiteStatsDaily represents the site_stats_daily table.
type SiteStatsDaily struct {
	bun.BaseModel `bun:"table:site_stats_daily,alias:ssd"`

	ID         int64     `bun:"id" json:"id"`
	StatDate   time.Time `bun:"stat_date" json:"stat_date"`
	PV         int64     `bun:"pv" json:"pv"`
	UV         int64     `bun:"uv" json:"uv"`
	OnlinePeak int32     `bun:"online_peak" json:"online_peak"`
}

// OnlineSessions represents the online_sessions table.
type OnlineSessions struct {
	bun.BaseModel `bun:"table:online_sessions,alias:os"`

	ID           int64     `bun:"id" json:"id"`
	SessionKey   string    `bun:"session_key" json:"session_key"`
	IP           string    `bun:"ip" json:"ip"`
	LastActiveAt time.Time `bun:"last_active_at" json:"last_active_at"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

// SiteStatsDaily represents the site_stats_daily table.
type SiteStatsDaily struct {
	bun.BaseModel `bun:"table:site_stats_daily,alias:ssd"`

	ID       uint64    `bun:"id,pk,autoincrement" json:"id"`
	StatDate time.Time `bun:"stat_date,notnull,unique" json:"stat_date"`
	// 页面浏览量
	PV uint64 `bun:"pv,notnull" json:"pv"`
	// 独立访客（按IP近似）
	UV uint64 `bun:"uv,notnull" json:"uv"`
	// 当日在线峰值（近似）
	OnlinePeak uint32 `bun:"online_peak,notnull" json:"online_peak"`
}

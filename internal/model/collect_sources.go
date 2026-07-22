package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CollectSources represents the collect_sources table.
type CollectSources struct {
	bun.BaseModel `bun:"table:collect_sources,alias:cs"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 采集源名称
	Name string `bun:"name,notnull" json:"name"`
	// 采集源格式：1默认(系统格式) 2苹果CMS格式
	Type uint8 `bun:"type,notnull" json:"type"`
	// 采集地址
	CollectURL string `bun:"collect_url,notnull" json:"collect_url"`
	// API密钥
	APIKey string `bun:"api_key,notnull" json:"-"`
	// 采集配置
	Config *string `bun:"config" json:"config"`
	// 定时采集cron表达式，空表示未开启定时采集
	CronExpr string `bun:"cron_expr,notnull" json:"cron_expr"`
	// 绑定播放源ID，采集到的播放链接存入该播放源
	// Relation: play_source_id -> PlaySources(ID)
	PlaySourceID uint64 `bun:"play_source_id,notnull" json:"play_source_id"`
	// 最后采集时间
	LastCollectAt *time.Time `bun:"last_collect_at" json:"last_collect_at"`
	// 状态：1启用 0禁用
	Status    uint8      `bun:"status,notnull" json:"status"`
	CreatedAt *time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at"`
	// 软删除时间
	DeletedAt               *time.Time                 `bun:"deleted_at" json:"deleted_at"`
	CollectLogs             []*CollectLogs             `bun:"rel:has-many,join:id=source_id" json:"-"`
	CollectSourceCategories []*CollectSourceCategories `bun:"rel:has-many,join:id=source_id" json:"-"`
	PlaySource              *PlaySources               `bun:"rel:belongs-to,join:play_source_id=id" json:"-"`
}

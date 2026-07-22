package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CollectLogs represents the collect_logs table.
type CollectLogs struct {
	bun.BaseModel `bun:"table:collect_logs,alias:cl"`

	ID uint64 `bun:"id,pk,autoincrement" json:"id"`
	// 采集源ID
	// Relation: source_id -> CollectSources(ID)
	SourceID uint64 `bun:"source_id,notnull" json:"source_id"`
	// 采集状态：1成功 2失败 3部分成功
	Status uint8 `bun:"status,notnull" json:"status"`
	// 采集总数
	TotalCount uint32 `bun:"total_count,notnull" json:"total_count"`
	// 成功数
	SuccessCount uint32 `bun:"success_count,notnull" json:"success_count"`
	// 失败数
	FailedCount uint32 `bun:"failed_count,notnull" json:"failed_count"`
	// 错误信息
	ErrorMessage *string `bun:"error_message" json:"error_message"`
	// 耗时(毫秒)
	DurationMs    uint32          `bun:"duration_ms,notnull" json:"duration_ms"`
	CreatedAt     *time.Time      `bun:"created_at" json:"created_at"`
	CollectSource *CollectSources `bun:"rel:belongs-to,join:source_id=id" json:"-"`
}

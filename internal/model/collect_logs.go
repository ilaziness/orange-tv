package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CollectLogs represents the collect_logs table.
type CollectLogs struct {
	bun.BaseModel `bun:"table:collect_logs,alias:c"`

	ID int64 `bun:"id" json:"id"`
	// 采集源ID
	SourceID int64 `bun:"source_id" json:"source_id"`
	// 采集状态：1成功 2失败 3部分成功
	Status int8 `bun:"status" json:"status"`
	// 采集总数
	TotalCount int32 `bun:"total_count" json:"total_count"`
	// 成功数
	SuccessCount int32 `bun:"success_count" json:"success_count"`
	// 失败数
	FailedCount int32 `bun:"failed_count" json:"failed_count"`
	// 错误信息
	ErrorMessage *string `bun:"error_message" json:"error_message"`
	// 耗时(毫秒)
	DurationMs int32      `bun:"duration_ms" json:"duration_ms"`
	CreatedAt  *time.Time `bun:"created_at" json:"created_at"`
}

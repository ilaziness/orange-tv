package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CollectSourceListRequest filters collect sources.
type CollectSourceListRequest struct {
	dto.PaginationRequest
	// 状态筛选（0=禁用，1=启用）
	Status *uint8 `form:"status"`
}

// CreateCollectSourceRequest creates a collect source.
type CreateCollectSourceRequest struct {
	// 采集源名称（必填）
	Name string `json:"name" binding:"required,min=1,max=100"`
	// 采集源类型（1=API，2=其他）
	Type uint8 `json:"type" binding:"required,oneof=1 2"`
	// 采集接口地址（必填）
	CollectURL string `json:"collect_url" binding:"required,min=1,max=500"`
	// API 密钥
	APIKey string `json:"api_key" binding:"omitempty,max=255"`
	// 定时采集 cron 表达式
	CronExpr string `json:"cron_expr" binding:"omitempty,max=100"`
	// 关联播放源ID（必填）
	PlaySourceID uint32 `json:"play_source_id" binding:"required,gt=0"`
	// 采集数据范围（如 today、last1d 等）
	DataRange string `json:"data_range" binding:"omitempty,max=20"`
}

// UpdateCollectSourceRequest updates a collect source.
type UpdateCollectSourceRequest struct {
	// 采集源名称
	Name *string `json:"name" binding:"omitempty,min=1,max=100"`
	// 采集源类型（1=API，2=其他）
	Type *uint8 `json:"type" binding:"omitempty,oneof=1 2"`
	// 采集接口地址
	CollectURL *string `json:"collect_url" binding:"omitempty,min=1,max=500"`
	// API 密钥
	APIKey *string `json:"api_key" binding:"omitempty,max=255"`
	// 定时采集 cron 表达式
	CronExpr *string `json:"cron_expr" binding:"omitempty,max=100"`
	// 关联播放源ID
	PlaySourceID *uint32 `json:"play_source_id" binding:"omitempty,gt=0"`
	// 采集数据范围（如 today、last1d 等）
	DataRange *string `json:"data_range" binding:"omitempty,max=20"`
}

// CollectSourceItem is an admin collect source payload.
// API key is never returned in list/detail responses.
type CollectSourceItem struct {
	// 采集源ID
	ID uint32 `json:"id"`
	// 采集源名称
	Name string `json:"name"`
	// 采集源类型（1=API，2=其他）
	Type uint8 `json:"type"`
	// 采集接口地址
	CollectURL string `json:"collect_url"`
	// 定时采集 cron 表达式
	CronExpr string `json:"cron_expr"`
	// 关联播放源ID
	PlaySourceID uint32 `json:"play_source_id"`
	// 关联播放源名称
	PlaySourceName string `json:"play_source_name,omitempty"`
	// 最近采集时间
	LastCollectAt string `json:"last_collect_at,omitempty"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
	// 是否启用定时采集（0=否，1=是）
	ScheduleEnabled uint8 `json:"schedule_enabled"`
	// 采集数据范围
	DataRange string `json:"data_range,omitempty"`
}

// SetCollectCategoriesRequest replaces category mappings for a source.
// Empty items clears all mappings.
type SetCollectCategoriesRequest struct {
	// 分类映射列表（空数组则清空全部映射）
	Items []CollectCategoryInput `json:"items" binding:"omitempty,dive"`
}

// CollectCategoryInput is one external→internal mapping by integer IDs.
type CollectCategoryInput struct {
	// 外部分类ID（必填）
	ExternalCategoryID uint32 `json:"external_category_id" binding:"required,gt=0"`
	// 系统分类ID（必填）
	CategoryID uint32 `json:"category_id" binding:"required,gt=0"`
}

// CollectCategoryMapItem is an external→internal category mapping.
type CollectCategoryMapItem struct {
	// 映射ID
	ID uint32 `json:"id"`
	// 采集源ID
	SourceID uint32 `json:"source_id"`
	// 外部分类ID
	ExternalCategoryID uint32 `json:"external_category_id"`
	// 系统分类ID
	CategoryID uint32 `json:"category_id"`
}

// CollectLogListRequest filters collect logs.
type CollectLogListRequest struct {
	dto.PaginationRequest
	// 采集源ID筛选
	SourceID uint32 `form:"source_id"`
}

// CollectLogItem is one collect run log entry.
type CollectLogItem struct {
	// 日志ID
	ID uint32 `json:"id"`
	// 采集源ID
	SourceID uint32 `json:"source_id"`
	// 采集源名称
	SourceName string `json:"source_name,omitempty"`
	// 执行状态（0=失败，1=成功）
	Status uint8 `json:"status"`
	// 采集数量
	CollectCount uint32 `json:"collect_count"`
	// 执行耗时（秒）
	DurationSec uint32 `json:"duration_sec"`
	// 执行时间
	CreatedAt string `json:"created_at,omitempty"`
}

// CollectNowRequest triggers an immediate collection.
type CollectNowRequest struct {
	// 采集数据范围（必填：today=今日，last1d/last3d/last1w/last1m=近N天，all=全部）
	DataRange string `json:"data_range" binding:"required,oneof=today last1d last3d last1w last1m all"`
}

// RemoteCategoryItem is one external category from a collect source.
type RemoteCategoryItem struct {
	// 外部类型ID
	TypeID uint32 `json:"type_id"`
	// 外部类型名称
	TypeName string `json:"type_name"`
	// 外部父类型ID
	TypePID uint32 `json:"type_pid"`
}

// RemoteCategoryResponse is the response for fetching remote categories.
type RemoteCategoryResponse struct {
	// 外部分类列表
	List []RemoteCategoryItem `json:"list"`
}

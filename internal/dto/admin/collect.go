package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CollectSourceListRequest filters collect sources.
type CollectSourceListRequest struct {
	dto.PaginationRequest
	Status *uint8 `form:"status"`
}

// CreateCollectSourceRequest creates a collect source.
type CreateCollectSourceRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	Type         uint8  `json:"type" binding:"required,oneof=1 2"`
	CollectURL   string `json:"collect_url" binding:"required,min=1,max=500"`
	APIKey       string `json:"api_key" binding:"omitempty,max=255"`
	CronExpr     string `json:"cron_expr" binding:"omitempty,max=100"`
	PlaySourceID uint32 `json:"play_source_id" binding:"required,gt=0"`
	DataRange    string `json:"data_range" binding:"omitempty,max=20"`
}

// UpdateCollectSourceRequest updates a collect source.
type UpdateCollectSourceRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=100"`
	Type         *uint8  `json:"type" binding:"omitempty,oneof=1 2"`
	CollectURL   *string `json:"collect_url" binding:"omitempty,min=1,max=500"`
	APIKey       *string `json:"api_key" binding:"omitempty,max=255"`
	CronExpr     *string `json:"cron_expr" binding:"omitempty,max=100"`
	PlaySourceID *uint32 `json:"play_source_id" binding:"omitempty,gt=0"`
	DataRange    *string `json:"data_range" binding:"omitempty,max=20"`
}

// CollectSourceItem is an admin collect source payload.
// API key is never returned in list/detail responses.
type CollectSourceItem struct {
	ID              uint32 `json:"id"`
	Name            string `json:"name"`
	Type            uint8  `json:"type"`
	CollectURL      string `json:"collect_url"`
	CronExpr        string `json:"cron_expr"`
	PlaySourceID    uint32 `json:"play_source_id"`
	PlaySourceName  string `json:"play_source_name,omitempty"`
	LastCollectAt   string `json:"last_collect_at,omitempty"`
	Status          uint8  `json:"status"`
	ScheduleEnabled uint8  `json:"schedule_enabled"`
	DataRange       string `json:"data_range,omitempty"`
}

// SetCollectCategoriesRequest replaces category mappings for a source.
// Empty items clears all mappings.
type SetCollectCategoriesRequest struct {
	Items []CollectCategoryInput `json:"items" binding:"omitempty,dive"`
}

// CollectCategoryInput is one external→internal mapping by integer IDs.
type CollectCategoryInput struct {
	ExternalCategoryID uint32 `json:"external_category_id" binding:"required,gt=0"`
	CategoryID         uint32 `json:"category_id" binding:"required,gt=0"`
}

// CollectCategoryMapItem is an external→internal category mapping.
type CollectCategoryMapItem struct {
	ID                 uint32 `json:"id"`
	SourceID           uint32 `json:"source_id"`
	ExternalCategoryID uint32 `json:"external_category_id"`
	CategoryID         uint32 `json:"category_id"`
}

// CollectLogListRequest filters collect logs.
type CollectLogListRequest struct {
	dto.PaginationRequest
	SourceID uint32 `form:"source_id"`
}

// CollectLogItem is one collect run log entry.
type CollectLogItem struct {
	ID           uint32 `json:"id"`
	SourceID     uint32 `json:"source_id"`
	SourceName   string `json:"source_name,omitempty"`
	Status       uint8  `json:"status"`
	CollectCount uint32 `json:"collect_count"`
	DurationSec  uint32 `json:"duration_sec"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// CollectNowRequest triggers an immediate collection.
type CollectNowRequest struct {
	DataRange string `json:"data_range" binding:"required,oneof=today last1d last3d last1w last1m all"`
}

// RemoteCategoryItem is one external category from a collect source.
type RemoteCategoryItem struct {
	TypeID   uint32 `json:"type_id"`
	TypeName string `json:"type_name"`
	TypePID  uint32 `json:"type_pid"`
}

// RemoteCategoryResponse is the response for fetching remote categories.
type RemoteCategoryResponse struct {
	List []RemoteCategoryItem `json:"list"`
}

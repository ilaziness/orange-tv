package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// CollectSourceListRequest filters collect sources.
type CollectSourceListRequest struct {
	dto.PaginationRequest
	Status *uint8 `form:"status"`
}

// CreateCollectSourceRequest creates a collect source.
type CreateCollectSourceRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	Type         uint8  `json:"type" validate:"required,oneof=1 2"`
	CollectURL   string `json:"collect_url" validate:"required,min=1,max=500"`
	APIKey       string `json:"api_key" validate:"omitempty,max=255"`
	CronExpr     string `json:"cron_expr" validate:"omitempty,max=100"`
	PlaySourceID uint64 `json:"play_source_id" validate:"required,gt=0"`
	DataRange    string `json:"data_range" validate:"omitempty,max=20"`
}

// UpdateCollectSourceRequest updates a collect source.
type UpdateCollectSourceRequest struct {
	Name         *string `json:"name" validate:"omitempty,min=1,max=100"`
	Type         *uint8  `json:"type" validate:"omitempty,oneof=1 2"`
	CollectURL   *string `json:"collect_url" validate:"omitempty,min=1,max=500"`
	APIKey       *string `json:"api_key" validate:"omitempty,max=255"`
	CronExpr     *string `json:"cron_expr" validate:"omitempty,max=100"`
	PlaySourceID *uint64 `json:"play_source_id" validate:"omitempty,gt=0"`
	DataRange    *string `json:"data_range" validate:"omitempty,max=20"`
}

// SetCollectCategoriesRequest replaces category mappings for a source.
// Empty items clears all mappings.
type SetCollectCategoriesRequest struct {
	Items []CollectCategoryInput `json:"items" validate:"omitempty,dive"`
}

// CollectCategoryInput is one external→internal mapping.
type CollectCategoryInput struct {
	ExternalCategory string `json:"external_category" validate:"required,min=1,max=100"`
	CategoryID       uint64 `json:"category_id" validate:"required,gt=0"`
}

// CollectLogListRequest filters collect logs.
type CollectLogListRequest struct {
	dto.PaginationRequest
	SourceID uint64 `form:"source_id"`
}

// CollectNowRequest triggers an immediate collection.
type CollectNowRequest struct {
	DataRange string `json:"data_range" validate:"required,oneof=today last1d last3d last1w last1m all"`
}

// RemoteCategoryItem is one external category from a collect source.
type RemoteCategoryItem struct {
	TypeID   string `json:"type_id"`
	TypeName string `json:"type_name"`
	TypePID  string `json:"type_pid"`
}

// RemoteCategoryResponse is the response for fetching remote categories.
type RemoteCategoryResponse struct {
	List []RemoteCategoryItem `json:"list"`
}

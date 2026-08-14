package admin

import "github.com/ilaziness/orange-tv/internal/dto"

// NameSearchRequest is shared by directors/actors/tags list endpoints.
type NameSearchRequest struct {
	dto.PaginationRequest
	// 关键词搜索（名称）
	Keyword string `form:"keyword"`
}

// CreateNamedRequest creates a named resource.
type CreateNamedRequest struct {
	// 名称（必填，1-100 字）
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// UpdateNamedRequest updates a named resource.
type UpdateNamedRequest struct {
	// 名称（必填，1-100 字）
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// NamedResponse is a generic named entity response.
type NamedResponse struct {
	// 条目ID
	ID uint32 `json:"id"`
	// 条目名称
	Name string `json:"name"`
}

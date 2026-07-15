// Package dto provides data transfer objects.
package dto

// PaginationRequest 分页请求参数（兼容 page_size / limit）。
type PaginationRequest struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1,max=1000000"`
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
	Limit    int `form:"limit" json:"limit" binding:"omitempty,min=1,max=100"`
}

// GetPage 获取页码，如果未设置则返回默认值1
func (p *PaginationRequest) GetPage() int {
	if p.Page < 1 {
		return 1
	}
	return p.Page
}

// GetLimit 获取每页数量，优先 page_size，其次 limit，默认 20。
func (p *PaginationRequest) GetLimit() int {
	size := p.PageSize
	if size < 1 {
		size = p.Limit
	}
	if size < 1 || size > 100 {
		return 20
	}
	return size
}

// GetPageSize is an alias for GetLimit used by handlers expecting PRD field names.
func (p *PaginationRequest) GetPageSize() int {
	return p.GetLimit()
}

// GetOffset 获取偏移量
func (p *PaginationRequest) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

// GetTotalPages 计算总页数
func (p *PaginationRequest) GetTotalPages(total int) int {
	if total == 0 || p.GetLimit() == 0 {
		return 0
	}
	pages := total / p.GetLimit()
	if total%p.GetLimit() > 0 {
		pages++
	}
	return pages
}

// IDURI is a common URI path parameter.
type IDURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

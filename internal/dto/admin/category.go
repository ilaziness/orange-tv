package admin

// CategoryResponse is a category tree node for admin (includes status & sort_order).
type CategoryResponse struct {
	// 分类ID
	ID uint32 `json:"id"`
	// 分类名称
	Name string `json:"name"`
	// 父分类ID，0 表示顶级分类
	ParentID uint32 `json:"parent_id"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order"`
	// 状态（0=禁用，1=启用）
	Status uint8 `json:"status"`
	// 子分类列表
	Children []CategoryResponse `json:"children"`
}

// CreateCategoryRequest creates a category.
type CreateCategoryRequest struct {
	// 分类名称（必填，1-100 字）
	Name string `json:"name" binding:"required,min=1,max=100"`
	// 父分类ID，0 表示顶级分类
	ParentID uint32 `json:"parent_id" binding:"omitempty,min=0"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateCategoryRequest updates a category.
type UpdateCategoryRequest struct {
	// 分类名称
	Name *string `json:"name" binding:"omitempty,min=1,max=100"`
	// 父分类ID，0 表示顶级分类
	ParentID *uint32 `json:"parent_id" binding:"omitempty,min=0"`
	// 排序权重，值越小越靠前
	SortOrder *uint32 `json:"sort_order" binding:"omitempty,min=0"`
	// 状态（0=禁用，1=启用）
	Status *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

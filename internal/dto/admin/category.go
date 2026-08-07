package admin

// CategoryResponse is a category tree node for admin (includes status & sort_order).
type CategoryResponse struct {
	ID        uint32             `json:"id"`
	Name      string             `json:"name"`
	ParentID  uint32             `json:"parent_id"`
	SortOrder uint32             `json:"sort_order"`
	Status    uint8              `json:"status"`
	Children  []CategoryResponse `json:"children"`
}

// CreateCategoryRequest creates a category.
type CreateCategoryRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	ParentID  uint32 `json:"parent_id" binding:"omitempty,min=0"`
	SortOrder uint32 `json:"sort_order" binding:"omitempty,min=0"`
	Status    *uint8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateCategoryRequest updates a category.
type UpdateCategoryRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=100"`
	ParentID  *uint32 `json:"parent_id" binding:"omitempty,min=0"`
	SortOrder *uint32 `json:"sort_order" binding:"omitempty,min=0"`
	Status    *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

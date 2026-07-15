package admin

// CreateCategoryRequest creates a category.
type CreateCategoryRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	ParentID  int64  `json:"parent_id" validate:"omitempty,min=0"`
	SortOrder int32  `json:"sort_order" validate:"omitempty,min=0"`
	Status    *int8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateCategoryRequest updates a category.
type UpdateCategoryRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	ParentID  *int64  `json:"parent_id" validate:"omitempty,min=0"`
	SortOrder *int32  `json:"sort_order" validate:"omitempty,min=0"`
	Status    *int8   `json:"status" validate:"omitempty,oneof=0 1"`
}

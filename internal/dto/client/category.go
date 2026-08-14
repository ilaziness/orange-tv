package client

// CategoryResponse is a category tree node for client (no status, only display fields).
type CategoryResponse struct {
	// 分类ID
	ID uint32 `json:"id"`
	// 分类名称
	Name string `json:"name"`
	// 父分类ID，0 表示顶级分类
	ParentID uint32 `json:"parent_id"`
	// 排序权重，值越小越靠前
	SortOrder uint32 `json:"sort_order"`
	// 子分类列表
	Children []CategoryResponse `json:"children"`
}

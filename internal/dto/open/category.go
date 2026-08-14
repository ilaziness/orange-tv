package open

// CategoryItem is a flat category payload for the open API.
type CategoryItem struct {
	// 分类ID
	ID uint32 `json:"id"`
	// 分类名称
	Name string `json:"name"`
	// 父分类ID，0 表示顶级分类
	ParentID uint32 `json:"parent_id"`
}

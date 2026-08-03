package open

// CategoryItem is a flat category payload for the open API.
type CategoryItem struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name"`
	ParentID uint32 `json:"parent_id"`
}

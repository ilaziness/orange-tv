package open

// CategoryItem is a flat category payload for the open API.
type CategoryItem struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	ParentID uint64 `json:"parent_id"`
}

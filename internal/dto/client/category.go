package client

// CategoryResponse is a category tree node for client (no status, only display fields).
type CategoryResponse struct {
	ID        uint64             `json:"id"`
	Name      string             `json:"name"`
	ParentID  uint64             `json:"parent_id"`
	SortOrder uint32             `json:"sort_order"`
	Children  []CategoryResponse `json:"children"`
}

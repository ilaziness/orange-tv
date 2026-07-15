package dto

// NameSearchRequest is shared by directors/actors/tags list endpoints.
type NameSearchRequest struct {
	PaginationRequest
	Keyword string `form:"keyword"`
}

// CreateNamedRequest creates a named resource.
type CreateNamedRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateNamedRequest updates a named resource.
type UpdateNamedRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// NamedResponse is a generic named entity response.
type NamedResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

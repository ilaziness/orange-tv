// Package http provides HTTP request handlers.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/response"
)

// StubHandler provides placeholder endpoints for API surface scaffolding.
// Business logic is implemented in later phases.
type StubHandler struct{}

// NewStubHandler creates a stub handler.
func NewStubHandler() *StubHandler {
	return &StubHandler{}
}

// NotImplemented responds with an empty success payload for scaffolded routes.
func (h *StubHandler) NotImplemented(c *gin.Context) {
	response.Success(c, nil)
}

// EmptyList responds with an empty paginated list for scaffolded list routes.
func (h *StubHandler) EmptyList(c *gin.Context) {
	response.SuccessPage(c, []any{}, 0, 1, 20, 0)
}

// EmptyArray responds with an empty array for scaffolded collection routes.
func (h *StubHandler) EmptyArray(c *gin.Context) {
	c.JSON(http.StatusOK, response.Response{
		Code:    0,
		Message: "success",
		Data:    []any{},
	})
}

package gen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user_name", "UserName"},
		{"users", "Users"},
		{"order_item_id", "OrderItemId"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, ToCamelCase(tt.input))
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"UserName", "user_name"},
		{"OrderItemID", "order_item_id"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, ToSnakeCase(tt.input))
		})
	}
}

func TestMapSQLTypeToGo(t *testing.T) {
	assert.Equal(t, "int64", MapSQLTypeToGo("bigint", false))
	assert.Equal(t, "*string", MapSQLTypeToGo("text", true))
	assert.Equal(t, "string", MapSQLTypeToGo("unknown_type", false))
}

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
		{"order_item_id", "OrderItemID"},
		{"id", "ID"},
		{"play_url", "PlayURL"},
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
	assert.Equal(t, "int64", MapSQLTypeToGo("bigint", false, false))
	assert.Equal(t, "*string", MapSQLTypeToGo("text", true, false))
	assert.Equal(t, "string", MapSQLTypeToGo("unknown_type", false, false))
	assert.Equal(t, "uint32", MapSQLTypeToGo("int", false, true))
	assert.Equal(t, "*uint32", MapSQLTypeToGo("int", true, true))
}

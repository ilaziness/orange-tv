package gen

import (
	"context"
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTables_returnsUserTablesFromSQLite(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE sample_items (id INTEGER PRIMARY KEY, title TEXT NOT NULL)`)
	require.NoError(t, err)

	tables, err := GetTables(ctx, db, constant.DriverSQLite)
	require.NoError(t, err)
	assert.Contains(t, tables, "sample_items")
}

func TestGetTableColumns_readsSQLiteColumns(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE sample_items (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		quantity INTEGER
	)`)
	require.NoError(t, err)

	columns, err := GetTableColumns(ctx, db, constant.DriverSQLite, "sample_items")
	require.NoError(t, err)
	require.Len(t, columns, 3)

	assert.Equal(t, "id", columns[0].Name)
	assert.Equal(t, "INTEGER", columns[0].Type)

	assert.Equal(t, "title", columns[1].Name)
	assert.False(t, columns[1].Nullable)

	assert.Equal(t, "quantity", columns[2].Name)
	assert.True(t, columns[2].Nullable)
}

func TestGenerateModelCode_includesBunAndJSONTags(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName:   "model",
		JSONTags:      true,
		ValidatorTags: false,
	}, "SampleItem", "sample_items", []Column{
		{Name: "id", Type: "INTEGER", Nullable: false},
		{Name: "title", Type: "TEXT", Nullable: false},
	}, nil)

	assert.Contains(t, code, "type SampleItem struct")
	assert.Contains(t, code, `bun:"id"`)
	assert.Contains(t, code, `json:"id"`)
	assert.Contains(t, code, `bun:"table:sample_items,alias:s"`)
}

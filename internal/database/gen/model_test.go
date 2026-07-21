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
	assert.True(t, columns[0].PrimaryKey)
	assert.True(t, columns[0].AutoIncrement)

	assert.Equal(t, "title", columns[1].Name)
	assert.False(t, columns[1].Nullable)

	assert.Equal(t, "quantity", columns[2].Name)
	assert.True(t, columns[2].Nullable)
}

func TestGetTableColumns_readsSQLiteUniqueConstraints(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE sample_items (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		UNIQUE(first_name, last_name)
	)`)
	require.NoError(t, err)

	columns, err := GetTableColumns(ctx, db, constant.DriverSQLite, "sample_items")
	require.NoError(t, err)

	email := findColumn(columns, "email")
	require.NotNil(t, email)
	assert.True(t, email.Unique)
	assert.Empty(t, email.UniqueGroup)

	firstName := findColumn(columns, "first_name")
	lastName := findColumn(columns, "last_name")
	require.NotNil(t, firstName)
	require.NotNil(t, lastName)
	assert.True(t, firstName.Unique)
	assert.True(t, lastName.Unique)
	assert.NotEmpty(t, firstName.UniqueGroup)
	assert.Equal(t, firstName.UniqueGroup, lastName.UniqueGroup)
}

func findColumn(columns []Column, name string) *Column {
	for i := range columns {
		if columns[i].Name == name {
			return &columns[i]
		}
	}
	return nil
}

func TestGenerateModelCode_addsUniqueTag(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
		JSONTags:    true,
	}, "User", "users", []Column{
		{Name: "id", Type: "bigint", Nullable: false, PrimaryKey: true, AutoIncrement: true},
		{Name: "email", Type: "varchar", Nullable: false, Unique: true},
		{Name: "first_name", Type: "varchar", Nullable: true, Unique: true, UniqueGroup: "uk_name"},
	}, nil)

	assert.Contains(t, code, "\tEmail string `bun:\"email,notnull,unique\" json:\"email\"`")
	assert.Contains(t, code, "\tFirstName *string `bun:\"first_name,unique:uk_name\" json:\"first_name\"`")
}

func TestGenerateModelCode_includesBunAndJSONTags(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName:   "model",
		JSONTags:      true,
		ValidatorTags: false,
	}, "SampleItem", "sample_items", []Column{
		{Name: "id", Type: "INTEGER", Nullable: false, PrimaryKey: true, AutoIncrement: true},
		{Name: "title", Type: "TEXT", Nullable: false},
	}, nil)

	assert.Contains(t, code, "type SampleItem struct")
	assert.Contains(t, code, "\tID int32 `bun:\"id,pk,autoincrement\" json:\"id\"`")
	assert.Contains(t, code, "\tTitle string `bun:\"title,notnull\" json:\"title\"`")
	assert.Contains(t, code, `bun:"table:sample_items,alias:si"`)
}

func TestGenerateModelCode_hidesPasswordJSON(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
		JSONTags:    true,
	}, "Users", "users", []Column{
		{Name: "id", Type: "bigint", Nullable: false},
		{Name: "password", Type: "varchar", Nullable: false},
	}, nil)

	assert.Contains(t, code, "\tPassword string `bun:\"password,notnull\" json:\"-\"`")
	assert.NotContains(t, code, `"time"`)
	assert.NotContains(t, code, `json:"password"`)
}

func TestIsSensitiveColumn(t *testing.T) {
	assert.True(t, isSensitiveColumn("password"))
	assert.True(t, isSensitiveColumn("api_key"))
	assert.False(t, isSensitiveColumn("username"))
}

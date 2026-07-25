package gen

import (
	"context"
	"go/format"
	"os"
	"path/filepath"
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

func TestGenerateModels_writesGofmtFormattedFiles(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE sample_items (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL
	)`)
	require.NoError(t, err)

	outputDir := t.TempDir()
	err = GenerateModels(ctx, db, constant.DriverSQLite, []string{"sample_items"}, GenerateOptions{
		OutputDir:   outputDir,
		PackageName: "model",
		JSONTags:    true,
	})
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(outputDir, "sample_items.gen.go"))
	require.NoError(t, err)

	want, err := format.Source(got)
	require.NoError(t, err)
	assert.Equal(t, string(want), string(got), "generated file should already be gofmt-formatted")
}

func TestGenerateModelCode_includesGeneratedComment(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
	}, "SampleItem", "sample_items", []Column{
		{Name: "id", Type: "INTEGER", Nullable: false, PrimaryKey: true, AutoIncrement: true},
		{Name: "title", Type: "TEXT", Nullable: false},
	}, nil)

	assert.Contains(t, code, "// Code generated by orange-tv gen model. DO NOT EDIT.")
}

func TestGenerateModelCode_generatesTimestampHook(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
	}, "Videos", "videos", []Column{
		{Name: "id", Type: "bigint", Nullable: false, PrimaryKey: true, AutoIncrement: true, Unsigned: true},
		{Name: "title", Type: "varchar", Nullable: false},
		{Name: "created_at", Type: "datetime", Nullable: true},
		{Name: "updated_at", Type: "datetime", Nullable: true},
	}, nil)

	assert.Contains(t, code, `"context"`)
	assert.Contains(t, code, "var _ bun.BeforeAppendModelHook = (*Videos)(nil)")
	assert.Contains(t, code, "func (m *Videos) BeforeAppendModel(ctx context.Context, query bun.Query) error {")
	assert.Contains(t, code, "case *bun.InsertQuery:")
	assert.Contains(t, code, "m.CreatedAt = &now")
	assert.Contains(t, code, "case *bun.UpdateQuery:")
	assert.Contains(t, code, "m.UpdatedAt = &now")
}

func TestGenerateModelCode_generatesHookForNonNullableTimestamps(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
	}, "Banners", "banners", []Column{
		{Name: "id", Type: "bigint", Nullable: false, PrimaryKey: true, AutoIncrement: true, Unsigned: true},
		{Name: "created_at", Type: "datetime", Nullable: false},
		{Name: "updated_at", Type: "datetime", Nullable: false},
	}, nil)

	assert.Contains(t, code, "m.CreatedAt = now")
	assert.Contains(t, code, "m.UpdatedAt = now")
	assert.NotContains(t, code, "m.CreatedAt = &now")
	assert.NotContains(t, code, "m.UpdatedAt = &now")
}

func TestGenerateModelCode_generatesHookForCreatedAtOnly(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
	}, "CollectLogs", "collect_logs", []Column{
		{Name: "id", Type: "bigint", Nullable: false, PrimaryKey: true, AutoIncrement: true, Unsigned: true},
		{Name: "created_at", Type: "datetime", Nullable: true},
	}, nil)

	assert.Contains(t, code, "case *bun.InsertQuery:")
	assert.Contains(t, code, "m.CreatedAt = &now")
	assert.NotContains(t, code, "case *bun.UpdateQuery:")
}

func TestGenerateModelCode_noHookWithoutTimestampColumns(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
	}, "VideoActors", "video_actors", []Column{
		{Name: "id", Type: "bigint", Nullable: false, PrimaryKey: true, AutoIncrement: true, Unsigned: true},
		{Name: "video_id", Type: "bigint", Nullable: false, Unsigned: true},
		{Name: "actor_id", Type: "bigint", Nullable: false, Unsigned: true},
	}, nil)

	assert.NotContains(t, code, "BeforeAppendModelHook")
	assert.NotContains(t, code, "BeforeAppendModel")
	assert.NotContains(t, code, `"context"`)
}

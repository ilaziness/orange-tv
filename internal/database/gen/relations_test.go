package gen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type relationTestCategory struct {
	bun.BaseModel `bun:"table:relation_test_categories,alias:rtc"`

	ID     int64                `bun:"id,pk"`
	Videos []*relationTestVideo `bun:"rel:has-many,join:id=category_id"`
}

type relationTestVideo struct {
	bun.BaseModel `bun:"table:relation_test_videos,alias:rtv"`

	ID         int64                 `bun:"id,pk"`
	CategoryID int64                 `bun:"category_id,notnull"`
	Category   *relationTestCategory `bun:"rel:belongs-to,join:category_id=id"`
}

func TestBunRelations_arePreloadable(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE relation_test_categories (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE relation_test_videos (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO relation_test_categories (id) VALUES (1)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO relation_test_videos (id, category_id) VALUES (10, 1)`)
	require.NoError(t, err)

	var video relationTestVideo
	err = db.NewSelect().Model(&video).Relation("Category").Where("rtv.id = ?", 10).Scan(ctx)
	require.NoError(t, err)
	require.NotNil(t, video.Category)
	assert.Equal(t, int64(1), video.Category.ID)

	var category relationTestCategory
	err = db.NewSelect().Model(&category).Relation("Videos").Where("rtc.id = ?", 1).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, category.Videos, 1)
	assert.Equal(t, int64(10), category.Videos[0].ID)
}

func TestLoadRelations_supportsShorthandAndOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model-relations.yaml")
	err := os.WriteFile(path, []byte(`relations:
  - videos.category_id -> categories.id
  - source: videos.created_by
    target: users.id
    field: Creator
    reverse_field: CreatedVideos
`), 0o600)
	require.NoError(t, err)

	relations, err := LoadRelations(path)
	require.NoError(t, err)
	require.Equal(t, []Relation{
		{
			SourceTable: "videos", SourceColumn: "category_id",
			TargetTable: "categories", TargetColumn: "id",
		},
		{
			SourceTable: "videos", SourceColumn: "created_by",
			TargetTable: "users", TargetColumn: "id",
			Field: "Creator", ReverseField: "CreatedVideos",
		},
	}, relations)
}

func TestLoadRelations_rejectsInvalidShorthand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model-relations.yaml")
	err := os.WriteFile(path, []byte("relations:\n  - videos.category_id categories.id\n"), 0o600)
	require.NoError(t, err)

	_, err = LoadRelations(path)
	assert.ErrorContains(t, err, "invalid relation")
}

func TestLoadRelations_rejectsUnknownRootField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model-relations.yaml")
	err := os.WriteFile(path, []byte("relatons: []\n"), 0o600)
	require.NoError(t, err)

	_, err = LoadRelations(path)
	assert.ErrorContains(t, err, "field relatons not found")
}

func TestLoadRelations_rejectsUnknownObjectField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model-relations.yaml")
	err := os.WriteFile(path, []byte(`relations:
  - source: videos.category_id
    target: categories.id
    reverse_feild: Videos
`), 0o600)
	require.NoError(t, err)

	_, err = LoadRelations(path)
	assert.ErrorContains(t, err, "unknown relation field \"reverse_feild\"")
}

func TestMergeRelations_rejectsSelfRelationFieldConflict(t *testing.T) {
	_, err := mergeRelations(nil, []Relation{{
		SourceTable: "categories", SourceColumn: "parent_id",
		TargetTable: "categories", TargetColumn: "id",
		Field: "Parent", ReverseField: "Parent",
	}})

	assert.ErrorContains(t, err, "relation field conflict on categories.Parent")
}

func TestMergeRelations_mergesPhysicalAndLogicalRelation(t *testing.T) {
	physical := []Relation{{
		SourceTable: "videos", SourceColumn: "category_id",
		TargetTable: "categories", TargetColumn: "id",
	}}
	logical := []Relation{{
		SourceTable: "videos", SourceColumn: "category_id",
		TargetTable: "categories", TargetColumn: "id",
		Field: "VideoCategory", ReverseField: "CategoryVideos",
	}}

	relations, err := mergeRelations(physical, logical)
	require.NoError(t, err)
	assert.Equal(t, logical, relations)
}

func TestGenerateModels_rejectsRelationFieldConflictWithBaseModel(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE base_models (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE records (id INTEGER PRIMARY KEY, base_model_id INTEGER NOT NULL)`)
	require.NoError(t, err)

	err = GenerateModels(ctx, db, constant.DriverSQLite, []string{"records"}, GenerateOptions{
		OutputDir:   t.TempDir(),
		PackageName: "model",
		Relations: []Relation{{
			SourceTable: "records", SourceColumn: "base_model_id",
			TargetTable: "base_models", TargetColumn: "id",
		}},
	})

	assert.ErrorContains(t, err, "relation field conflict with column: records.BaseModel")
}

func TestGenerateModels_rejectsRelationFieldConflictWithColumn(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE categories (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE videos (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL, category TEXT NOT NULL)`)
	require.NoError(t, err)

	err = GenerateModels(ctx, db, constant.DriverSQLite, []string{"videos"}, GenerateOptions{
		OutputDir:   t.TempDir(),
		PackageName: "model",
		Relations: []Relation{{
			SourceTable: "videos", SourceColumn: "category_id",
			TargetTable: "categories", TargetColumn: "id",
		}},
	})

	assert.ErrorContains(t, err, "relation field conflict with column: videos.Category")
}

func TestGenerateModels_generatesLogicalRelationsWithoutForeignKeys(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE categories (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE videos (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL)`)
	require.NoError(t, err)

	outputDir := t.TempDir()
	err = GenerateModels(ctx, db, constant.DriverSQLite, []string{"categories", "videos"}, GenerateOptions{
		OutputDir:   outputDir,
		PackageName: "model",
		JSONTags:    true,
		Relations: []Relation{{
			SourceTable: "videos", SourceColumn: "category_id",
			TargetTable: "categories", TargetColumn: "id",
		}},
	})
	require.NoError(t, err)

	videos, err := os.ReadFile(filepath.Join(outputDir, "videos.gen.go"))
	require.NoError(t, err)
	// gofmt may align struct fields; match on relation tag content.
	assert.Contains(t, string(videos), `bun:"rel:belongs-to,join:category_id=id"`)
	assert.Contains(t, string(videos), "Category")
	assert.Contains(t, string(videos), "*Categories")

	categories, err := os.ReadFile(filepath.Join(outputDir, "categories.gen.go"))
	require.NoError(t, err)
	assert.Contains(t, string(categories), `bun:"rel:has-many,join:id=category_id"`)
	assert.Contains(t, string(categories), "Videos")
	assert.Contains(t, string(categories), "[]*Videos")
}

func TestGenerateModelCode_generatesBidirectionalRelations(t *testing.T) {
	code := GenerateModelCode(GenerateOptions{
		PackageName: "model",
		JSONTags:    true,
	}, "Videos", "videos", []Column{
		{Name: "id", Type: "bigint", PrimaryKey: true, AutoIncrement: true, Unsigned: true},
		{Name: "category_id", Type: "bigint", Unsigned: true},
	}, []Relation{{
		SourceTable: "videos", SourceColumn: "category_id",
		TargetTable: "categories", TargetColumn: "id",
	}})

	assert.Contains(t, code, "\tCategory *Categories `bun:\"rel:belongs-to,join:category_id=id\" json:\"-\"`")

	code = GenerateModelCode(GenerateOptions{
		PackageName: "model",
		JSONTags:    true,
	}, "Categories", "categories", []Column{
		{Name: "id", Type: "bigint", PrimaryKey: true, AutoIncrement: true, Unsigned: true},
	}, []Relation{{
		SourceTable: "videos", SourceColumn: "category_id",
		TargetTable: "categories", TargetColumn: "id",
	}})

	assert.Contains(t, code, "\tVideos []*Videos `bun:\"rel:has-many,join:id=category_id\" json:\"-\"`")
}

func TestGenerateModels_expandsTablesToIncludeRelationEndpoints(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE TABLE categories (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE videos (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL)`)
	require.NoError(t, err)

	outputDir := t.TempDir()
	err = GenerateModels(ctx, db, constant.DriverSQLite, []string{"videos"}, GenerateOptions{
		OutputDir:   outputDir,
		PackageName: "model",
		JSONTags:    true,
		Relations: []Relation{{
			SourceTable: "videos", SourceColumn: "category_id",
			TargetTable: "categories", TargetColumn: "id",
		}},
	})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "videos.gen.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(outputDir, "categories.gen.go"))
	require.NoError(t, err)
}

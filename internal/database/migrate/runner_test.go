package migrate

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "github.com/uptrace/bun/driver/sqliteshim"
)

const runnerTestDir = "testdata"

func TestUp_appliesPendingMigrations(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	result, err := Up(ctx, db, runnerTestDir, false)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.AppliedCount)

	var count int
	require.NoError(t, db.NewSelect().Table("test_users").ColumnExpr("COUNT(*)").Scan(ctx, &count))
}

func TestUp_dryRunDoesNotApplyMigrations(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	result, err := Up(ctx, db, runnerTestDir, true)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.PendingCount)
	assert.Equal(t, []string{"20250101120000"}, result.PendingNames)

	var count int
	err = db.NewSelect().Table("test_users").ColumnExpr("COUNT(*)").Scan(ctx, &count)
	require.Error(t, err)
}

func TestDown_rollsBackLastMigrationGroup(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := Up(ctx, db, runnerTestDir, false)
	require.NoError(t, err)

	_, err = Down(ctx, db, runnerTestDir, false)
	require.NoError(t, err)

	var count int
	err = db.NewSelect().Table("test_users").ColumnExpr("COUNT(*)").Scan(ctx, &count)
	require.Error(t, err)
}

func TestStatus_listsPendingAndAppliedMigrations(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	lines, err := Status(ctx, db, runnerTestDir)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "20250101120000", lines[0].Name)
	assert.False(t, lines[0].Applied)

	_, err = Up(ctx, db, runnerTestDir, false)
	require.NoError(t, err)

	lines, err = Status(ctx, db, runnerTestDir)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.True(t, lines[0].Applied)
}

func TestDown_dryRunReportsLastMigrationGroupOnly(t *testing.T) {
	db := testutil.OpenBunDB(t)
	ctx := context.Background()

	_, err := Up(ctx, db, runnerTestDir, false)
	require.NoError(t, err)

	result, err := Down(ctx, db, runnerTestDir, true)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.RollbackCount)
	assert.Equal(t, []string{"20250101120000"}, result.RollbackNames)
}

func writeMigrationPair(t *testing.T, dir, timestamp, name, upSQL, downSQL string) {
	t.Helper()

	up := filepath.Join(dir, timestamp+"_"+name+".up.sql")
	down := filepath.Join(dir, timestamp+"_"+name+".down.sql")
	require.NoError(t, os.WriteFile(up, []byte(upSQL), 0o644))
	require.NoError(t, os.WriteFile(down, []byte(downSQL), 0o644))
}

func openMigrateTestDB(t *testing.T) *database.DB {
	t.Helper()

	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	bunDB := bun.NewDB(sqldb, sqlitedialect.New())
	return &database.DB{DB: bunDB}
}

func TestDown_dryRunRollsBackLatestGroupOnly(t *testing.T) {
	dir, err := os.MkdirTemp("", "migrate-multi-*")
	require.NoError(t, err)

	db := openMigrateTestDB(t)
	ctx := context.Background()

	writeMigrationPair(t, dir, "20250101120000", "first_users",
		`CREATE TABLE IF NOT EXISTS first_users (id INTEGER PRIMARY KEY);`,
		`DROP TABLE IF EXISTS first_users;`,
	)
	_, err = Up(ctx, db, dir, false)
	require.NoError(t, err)

	writeMigrationPair(t, dir, "20250101120001", "second_users",
		`CREATE TABLE IF NOT EXISTS second_users (id INTEGER PRIMARY KEY);`,
		`DROP TABLE IF EXISTS second_users;`,
	)
	_, err = Up(ctx, db, dir, false)
	require.NoError(t, err)

	result, err := Down(ctx, db, dir, true)
	require.NoError(t, err)
	assert.Equal(t, 1, result.RollbackCount)
	assert.Equal(t, []string{"20250101120001"}, result.RollbackNames)

	require.NoError(t, db.Close())
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("temp dir cleanup skipped: %v", err)
	}
}

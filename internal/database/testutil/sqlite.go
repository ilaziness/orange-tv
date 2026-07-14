package testutil

import (
	"database/sql"
	"testing"

	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "github.com/uptrace/bun/driver/sqliteshim"
)

// OpenBunDB opens a shared in-memory SQLite database wrapped as database.DB.
func OpenBunDB(t *testing.T) *database.DB {
	t.Helper()

	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	bunDB := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() {
		require.NoError(t, bunDB.Close())
	})

	return &database.DB{DB: bunDB}
}

package migrate

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_writesMigrationFiles(t *testing.T) {
	dir := t.TempDir()

	upFile, downFile, err := Create(dir, "add_orders_table")
	require.NoError(t, err)

	assert.FileExists(t, upFile)
	assert.FileExists(t, downFile)
}

func TestCreate_rejectsInvalidName(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "dotdot", input: "../escape"},
		{name: "slash", input: "foo/bar"},
		{name: "uppercase", input: "Invalid-Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Create(dir, tt.input)
			require.Error(t, err)
		})
	}
}

func TestCreate_rejectsPathOutsideDirectory(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "..", "outside.sql")

	err := ensurePathWithinDir(dir, outside)
	require.Error(t, err)
}

func TestValidateMigrationName_acceptsValidNames(t *testing.T) {
	require.NoError(t, validateMigrationName("init_users_table"))
}

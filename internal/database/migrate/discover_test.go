package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMigrations_discoversSQLFiles(t *testing.T) {
	m, err := LoadMigrations("testdata")
	require.NoError(t, err)

	sorted := m.Sorted()
	require.Len(t, sorted, 1)
	assert.Equal(t, "20250101120000", sorted[0].Name)
	assert.NotNil(t, sorted[0].Up)
	assert.NotNil(t, sorted[0].Down)
}

func TestLoadMigrations_returnsErrorForMissingDirectory(t *testing.T) {
	_, err := LoadMigrations("testdata/does-not-exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "discover migrations")
}

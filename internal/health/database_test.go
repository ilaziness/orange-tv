package health

import (
	"context"
	"testing"

	"github.com/ilaziness/orange-tv/internal/database/testutil"
	"github.com/stretchr/testify/require"
)

func TestDatabaseChecker_Check_success(t *testing.T) {
	db := testutil.OpenBunDB(t)
	checker := NewDatabaseChecker(db)

	require.NoError(t, checker.Check(context.Background()))
	require.Equal(t, "database", checker.Name())
}

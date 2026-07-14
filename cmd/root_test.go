package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteE_version(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"version"})

	err := ExecuteE()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "orange-tv version")
}

func TestExecuteE_health_requiresHTTPOrURL(t *testing.T) {
	var stderr bytes.Buffer
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"health", "-c", "configs/config.yaml"})

	// Without a running server, probe fails with connection error or non-2xx.
	err := ExecuteE()
	require.Error(t, err)
}

func TestProbePath(t *testing.T) {
	path, err := probePath("readiness")
	require.NoError(t, err)
	assert.Equal(t, "/readiness", path)

	_, err = probePath("invalid")
	require.Error(t, err)
}

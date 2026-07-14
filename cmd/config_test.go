package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteE_configValidate_validConfig(t *testing.T) {
	cfgFile = filepath.Join("..", "internal", "config", "testdata", "config.yaml")

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"config", "validate"})

	err := ExecuteE()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Configuration is valid")
}

func TestExecuteE_configValidate_missingConfig(t *testing.T) {
	cfgFile = filepath.Join("..", "internal", "config", "testdata", "does-not-exist.yaml")

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"config", "validate"})

	err := ExecuteE()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

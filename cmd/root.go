package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	env     string
	version = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "orange-tv",
	Short: "A flexible Go application template framework",
	Long: `A flexible Go application template framework that supports multiple
service types and optional component integration, suitable for quickly
building various backend services.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// ExecuteE runs the root command and returns an error instead of exiting.
func ExecuteE() error {
	return rootCmd.Execute()
}

func Execute() {
	if err := ExecuteE(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file path (string; default: ./configs/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "", "Runtime environment (string: dev, prod, or test); overrides app.env in config")
}

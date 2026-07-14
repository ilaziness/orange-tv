package cmd

import (
	"fmt"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/spf13/cobra"
)

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE:  runConfigValidate,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	RunE:  runConfigShow,
}

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
	}
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithEnv(cfgFile, env)
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Configuration is valid")
	fmt.Fprintf(out, "App Name: %s\n", cfg.App.Name)
	fmt.Fprintf(out, "Environment: %s\n", cfg.App.Env)
	fmt.Fprintf(out, "HTTP: %s:%d (enabled: %v)\n", cfg.HTTP.Host, cfg.HTTP.Port, cfg.HTTP.Enabled)

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithEnv(cfgFile, env)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Current Configuration:")
	fmt.Fprintf(out, "  App Name: %s\n", cfg.App.Name)
	fmt.Fprintf(out, "  Version: %s\n", cfg.App.Version)
	fmt.Fprintf(out, "  Environment: %s\n", cfg.App.Env)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "HTTP Server:")
	fmt.Fprintf(out, "  Enabled: %v\n", cfg.HTTP.Enabled)
	fmt.Fprintf(out, "  Host: %s\n", cfg.HTTP.Host)
	fmt.Fprintf(out, "  Port: %d\n", cfg.HTTP.Port)
	fmt.Fprintf(out, "  Read Timeout: %ds\n", cfg.HTTP.ReadTimeout)
	fmt.Fprintf(out, "  Write Timeout: %ds\n", cfg.HTTP.WriteTimeout)
	fmt.Fprintf(out, "  Shutdown Timeout: %ds\n", cfg.HTTP.ShutdownTimeout)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Logging:")
	fmt.Fprintf(out, "  Level: %s\n", cfg.Log.Level)
	fmt.Fprintf(out, "  Format: %s\n", cfg.Log.Format)
	fmt.Fprintf(out, "  Output: %s\n", cfg.Log.Output)
	if cfg.Log.Output == "file" || cfg.Log.Output == "both" {
		fmt.Fprintf(out, "  Filename: %s\n", cfg.Log.Filename)
		fmt.Fprintf(out, "  Max Size: %dMB\n", cfg.Log.MaxSize)
		fmt.Fprintf(out, "  Max Backups: %d\n", cfg.Log.MaxBackups)
		fmt.Fprintf(out, "  Max Age: %d days\n", cfg.Log.MaxAge)
		fmt.Fprintf(out, "  Compress: %v\n", cfg.Log.Compress)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Database:")
	fmt.Fprintf(out, "  Driver: %s\n", cfg.Database.Driver)
	fmt.Fprintf(out, "  Host: %s\n", cfg.Database.Host)
	fmt.Fprintf(out, "  Port: %d\n", cfg.Database.Port)
	fmt.Fprintf(out, "  User: %s\n", cfg.Database.User)
	fmt.Fprintf(out, "  Password: %s\n", cfg.MaskedDatabasePassword())
	fmt.Fprintf(out, "  Database: %s\n", cfg.Database.Database)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Redis:")
	fmt.Fprintf(out, "  Enabled: %v\n", cfg.Redis.Enabled)
	if cfg.Redis.Enabled {
		fmt.Fprintf(out, "  Host: %s\n", cfg.Redis.Host)
		fmt.Fprintf(out, "  Port: %d\n", cfg.Redis.Port)
		fmt.Fprintf(out, "  Password: %s\n", cfg.MaskedRedisPassword())
		fmt.Fprintf(out, "  DB: %d\n", cfg.Redis.DB)
		fmt.Fprintf(out, "  Pool Size: %d\n", cfg.Redis.PoolSize)
	}

	return nil
}

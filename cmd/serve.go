package cmd

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/app"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the application server",
	Long: `Start the application server with the specified configuration.
Supports HTTP services based on configuration.`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithEnv(cfgFile, env)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if validateErr := cfg.Validate(); validateErr != nil {
		return fmt.Errorf("configuration validation failed: %w", validateErr)
	}

	if cfg.App.Env == constant.EnvProd {
		gin.SetMode(gin.ReleaseMode)
	}

	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	if err := application.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}

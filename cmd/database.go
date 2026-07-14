package cmd

import (
	"fmt"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/logger"
)

func newLoggerFromConfig(cfg *config.Config) (*logger.Logger, error) {
	logInst, err := logger.New(logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return logInst, nil
}

func loadDatabase() (*database.DB, error) {
	cfg, err := config.LoadWithEnv(cfgFile, env)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logInst, err := newLoggerFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	db, err := database.NewDB(cfg, logInst.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

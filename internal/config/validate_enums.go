package config

import (
	"fmt"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
)

func validateOneOf(field, value string, allowed ...string) error {
	if value == "" {
		return nil
	}
	for _, item := range allowed {
		if value == item {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", field, strings.Join(allowed, ", "))
}

func validateOneOfFold(field, value string, allowed ...string) error {
	if value == "" {
		return nil
	}
	lower := strings.ToLower(value)
	for _, item := range allowed {
		if lower == strings.ToLower(item) {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", field, strings.Join(allowed, ", "))
}

func (c *Config) validateEnums() error {
	if err := validateOneOf("app.env", c.App.Env, constant.EnvDev, constant.EnvProd, constant.EnvTest); err != nil {
		return err
	}
	if err := validateOneOfFold("log.level", c.Log.Level, constant.LogLevelDebug, constant.LogLevelInfo, constant.LogLevelWarn, constant.LogLevelError); err != nil {
		return err
	}
	if err := validateOneOfFold("log.output", c.Log.Output, constant.LogOutputStdout, constant.LogOutputFile, constant.LogOutputBoth); err != nil {
		return err
	}
	if c.Log.Output == constant.LogOutputFile || c.Log.Output == constant.LogOutputBoth ||
		strings.EqualFold(c.Log.Output, constant.LogOutputFile) || strings.EqualFold(c.Log.Output, constant.LogOutputBoth) {
		if c.Log.Filename == "" {
			return fmt.Errorf("log.filename is required when log.output is file or both")
		}
	}

	if err := validateOneOf(
		"database.driver",
		c.Database.Driver,
		constant.DriverMySQL,
		constant.DriverPostgreSQL,
		constant.DriverPostgres,
		constant.DriverSQLite,
	); err != nil {
		return err
	}

	if c.Database.Driver == constant.DriverPostgreSQL || c.Database.Driver == constant.DriverPostgres {
		if err := validateOneOf("database.ssl_mode", c.Database.SSLMode, "disable", "require", "verify-ca", "verify-full"); err != nil {
			return err
		}
	}

	if err := validateOneOf("http.tls.min_version", c.HTTP.TLS.MinVersion, "1.2", "1.3"); err != nil {
		return err
	}

	if c.RateLimit.Enabled {
		if err := validateOneOf("rate_limit.store", c.RateLimit.Store, "memory", "redis"); err != nil {
			return err
		}
	}

	if c.Cache.Enabled {
		if err := validateOneOf("cache.driver", c.Cache.Driver, "memory", "redis", "multi"); err != nil {
			return err
		}
	}

	return nil
}

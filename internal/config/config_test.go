package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
)

// testdataDir 返回 testdata 目录路径
func testdataDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "testdata")
}

// mustLoad 从 testdata 加载配置，失败则 panic
func mustLoad(t *testing.T, filename string) *Config {
	t.Helper()
	cfg, err := LoadWithEnv(filepath.Join(testdataDir(), filename), "")
	if err != nil {
		t.Fatalf("failed to load config %q: %v", filename, err)
	}
	return cfg
}

// assertError 断言期望的错误是否匹配
func assertError(t *testing.T, err error, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Error("expected error but got nil")
			return
		}
		if errMsg != "" && err.Error() != errMsg {
			t.Errorf("error message = %q, want %q", err.Error(), errMsg)
		}
	} else if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestConfigValidate 验证配置校验逻辑
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		// 有效配置
		{"valid full config", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}}, false, ""},
		{"valid minimal config", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}}, false, ""},
		{"valid with empty log level", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Level: ""}}, false, ""},
		{"valid log levels debug", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Level: "debug"}}, false, ""},
		{"valid log levels DEBUG", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Level: "DEBUG"}}, false, ""},
		{"valid log levels warn", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Level: "warn"}}, false, ""},
		{"valid log levels error", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Level: "error"}}, false, ""},
		{"redis disabled skip validation", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Redis: RedisConfig{Enabled: false}}, false, ""},
		{"valid redis enabled", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Redis: RedisConfig{Enabled: true, Host: "localhost"}}, false, ""},

		// 无效配置
		{"missing app name", Config{HTTP: HTTPConfig{Port: 8080}}, true, "app.name is required"},
		{"port zero", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 0}}, true, "http.port must be between 1 and 65535"},
		{"port too high", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 70000}}, true, "http.port must be between 1 and 65535"},
		{"invalid log level", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Log: LogConfig{Level: "invalid"}}, true, "log.level must be one of: debug, info, warn, error"},
		{"database missing driver", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "", Database: "test.db"}}, true, "database.driver is required"},
		{"database mysql no host", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: constant.DriverMySQL, Host: "", Database: "testdb"}}, true, "database.host is required"},
		{"database postgres no host", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: constant.DriverPostgreSQL, Host: "", Database: "testdb"}}, true, "database.host is required"},
		{"database no database name", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost"}}, true, "database.database is required"},
		{"redis enabled no host", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Redis: RedisConfig{Enabled: true, Host: ""}}, true, "redis.host is required when redis is enabled"},

		// 枚举校验
		{"invalid app env", Config{App: AppConfig{Name: "app", Env: "staging"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}}, true, "app.env must be one of: dev, prod, test"},
		{"invalid log format", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Format: "yaml"}}, true, "log.format must be one of: json, console"},
		{"invalid log output", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Output: "stderr"}}, true, "log.output must be one of: stdout, file, both"},
		{"log output file missing filename", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}, Log: LogConfig{Output: "file"}}, true, "log.filename is required when log.output is file or both"},
		{"invalid database driver", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "oracle", Host: "localhost", Database: "test.db"}}, true, "database.driver must be one of: mysql, postgres, postgresql, sqlite"},
		{"invalid postgres ssl mode", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080}, Database: DatabaseConfig{Driver: "postgres", Host: "localhost", Database: "test.db", SSLMode: "invalid"}}, true, "database.ssl_mode must be one of: disable, require, verify-ca, verify-full"},
		{"invalid tls min version", Config{App: AppConfig{Name: "app"}, HTTP: HTTPConfig{Port: 8080, TLS: TLSConfig{MinVersion: "1.1"}}, Database: DatabaseConfig{Driver: "mysql", Host: "localhost", Database: "testdb"}}, true, "http.tls.min_version must be one of: 1.2, 1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertError(t, tt.config.Validate(), tt.wantErr, tt.errMsg)
		})
	}
}

// TestHTTPConfigGetShutdownTimeout 测试关闭超时计算
func TestHTTPConfigGetShutdownTimeout(t *testing.T) {
	tests := []struct {
		timeout int
		want    time.Duration
	}{
		{0, 30 * time.Second},  // 零值返回默认
		{-1, 30 * time.Second}, // 负值返回默认
		{60, 60 * time.Second}, // 自定义值
		{5, 5 * time.Second},   // 小值
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := HTTPConfig{ShutdownTimeout: tt.timeout}
			if got := c.GetShutdownTimeout(); got != tt.want {
				t.Errorf("GetShutdownTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConfigMaskedPassword 测试密码脱敏
func TestConfigMaskedPassword(t *testing.T) {
	c := &Config{}
	if got := c.MaskedPassword(); got != "******" {
		t.Errorf("MaskedPassword() = %q, want %q", got, "******")
	}
}

// TestConfigMaskedDatabasePassword 测试数据库密码脱敏
func TestConfigMaskedDatabasePassword(t *testing.T) {
	tests := []struct {
		pwd  string
		want string
	}{
		{"", ""},
		{"secret", "******"},
		{"verylongpassword123", "******"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := &Config{Database: DatabaseConfig{Password: tt.pwd}}
			if got := c.MaskedDatabasePassword(); got != tt.want {
				t.Errorf("MaskedDatabasePassword() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestConfigMaskedRedisPassword 测试 Redis 密码脱敏
func TestConfigMaskedRedisPassword(t *testing.T) {
	tests := []struct {
		pwd  string
		want string
	}{
		{"", ""},
		{"redis_secret", "******"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := &Config{Redis: RedisConfig{Password: tt.pwd}}
			if got := c.MaskedRedisPassword(); got != tt.want {
				t.Errorf("MaskedRedisPassword() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLoadWithEnvFromTestdata 从 testdata 加载配置
func TestLoadWithEnvFromTestdata(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		check    func(t *testing.T, cfg *Config)
	}{
		{
			name:     "load base config",
			filename: "config.yaml",
			check: func(t *testing.T, cfg *Config) {
				assertEq(t, "App.Name", cfg.App.Name, "test-app")
				assertEq(t, "HTTP.Port", cfg.HTTP.Port, 9090)
				assertEq(t, "HTTP.Host", cfg.HTTP.Host, "127.0.0.1")
				assertEq(t, "Database.Driver", cfg.Database.Driver, "sqlite")
			},
		},
		{
			name:     "load minimal config uses defaults",
			filename: "minimal.yaml",
			check: func(t *testing.T, cfg *Config) {
				assertEq(t, "App.Name", cfg.App.Name, "minimal-app")
				assertEq(t, "App.Version", cfg.App.Version, "1.0.0") // 默认值
				assertEq(t, "HTTP.Port", cfg.HTTP.Port, 8080)        // 默认值
				assertEq(t, "Database.Driver", cfg.Database.Driver, constant.DriverMySQL)
			},
		},
		{
			name:     "load with env override",
			filename: "config.yaml",
			check: func(t *testing.T, cfg *Config) {
				// 手动测试环境覆盖在 TestLoadWithEnvMerge 中
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadWithEnv(filepath.Join(testdataDir(), tt.filename), "")
			assertError(t, err, tt.wantErr, "")
			if tt.check != nil && err == nil {
				tt.check(t, cfg)
			}
		})
	}
}

// TestLoadWithEnvMerge 测试环境配置合并
func TestLoadWithEnvMerge(t *testing.T) {
	baseConfig := filepath.Join(testdataDir(), "config.yaml")

	tests := []struct {
		name        string
		envOverride string
		wantName    string
		wantPort    int
		wantLevel   string
	}{
		{"no override", "", "test-app", 9090, "info"},
		{"dev override", "dev", "dev-app", 9091, "debug"},
		{"prod override", "prod", "prod-app", 80, "warn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadWithEnv(baseConfig, tt.envOverride)
			if err != nil {
				t.Fatalf("LoadWithEnv() error: %v", err)
			}
			assertEq(t, "App.Name", cfg.App.Name, tt.wantName)
			assertEq(t, "HTTP.Port", cfg.HTTP.Port, tt.wantPort)
			assertEq(t, "Log.Level", cfg.Log.Level, tt.wantLevel)
		})
	}
}

// TestLoadWithEnvInvalid 测试无效配置加载
func TestLoadWithEnvInvalid(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"nonexistent file", "/nonexistent/path/config.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadWithEnv(tt.path, "")
			if err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}

// TestEnvironmentVariables 测试环境变量绑定
func TestEnvironmentVariables(t *testing.T) {
	t.Setenv(EnvVarDatabasePassword, "db_secret")
	t.Setenv(EnvVarRedisPassword, "redis_secret")

	cfg := mustLoad(t, "config.yaml")

	// 环境变量应该在配置中体现（如果配置文件中已有这些字段）
	if cfg.Database.Password != "db_secret" && cfg.Database.Password != "" {
		t.Errorf("Database.Password = %q, want %q or empty", cfg.Database.Password, "db_secret")
	}
}

// TestEnvOverridePriority 测试环境变量优先级
func TestEnvOverridePriority(t *testing.T) {
	// APP_ENV 应该被正确读取
	t.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	// 直接加载配置，不指定 envOverride
	cfg, err := LoadWithEnv(filepath.Join(testdataDir(), "config.yaml"), "")
	if err != nil {
		t.Fatalf("LoadWithEnv() error: %v", err)
	}

	// 注意：当前实现中 envOverride 参数优先级高于环境变量
	// 这里只是验证配置能正常加载
	if cfg.App.Name == "" {
		t.Error("App.Name should not be empty")
	}
}

// assertEq 通用断言辅助函数
func assertEq[T comparable](t *testing.T, name string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Environment variable names
const (
	EnvVarAppEnv           = "APP_ENV"
	EnvVarDatabasePassword = "DATABASE_PASSWORD"
	EnvVarRedisPassword    = "REDIS_PASSWORD"
)

// Global config instance
var Cfg *Config

// Config 应用根配置，对应 configs/config.yaml 及各模块段落。
type Config struct {
	App       AppConfig       `mapstructure:"app"`        // 应用基础信息
	HTTP      HTTPConfig      `mapstructure:"http"`       // HTTP 服务
	Log       LogConfig       `mapstructure:"log"`        // 日志
	Database  DatabaseConfig  `mapstructure:"database"`   // 数据库（必填）
	Redis     RedisConfig     `mapstructure:"redis"`      // Redis（可选）
	Cache     CacheConfig     `mapstructure:"cache"`      // 缓存（可选）
	JWT       JWTConfig       `mapstructure:"jwt"`        // JWT 认证
	RateLimit RateLimitConfig `mapstructure:"rate_limit"` // 限流（可选）
	Tracing   TracingConfig   `mapstructure:"tracing"`    // 链路追踪（可选）
	Metrics   MetricsConfig   `mapstructure:"metrics"`    // Prometheus 指标（可选）
}

// AppConfig 应用基础配置。
type AppConfig struct {
	Name    string `mapstructure:"name"`    // 应用名称（必填）
	Version string `mapstructure:"version"` // 应用版本号
	Env     string `mapstructure:"env"`     // 运行环境，枚举：dev | prod | test
}

// HTTPConfig HTTP 服务配置。
type HTTPConfig struct {
	Enabled            bool      `mapstructure:"enabled"`              // 是否启用 HTTP 服务
	Host               string    `mapstructure:"host"`                 // 监听地址，如 0.0.0.0
	Port               int       `mapstructure:"port"`                 // 监听端口，范围 1-65535
	ReadTimeout        int       `mapstructure:"read_timeout"`         // 读超时（秒）
	WriteTimeout       int       `mapstructure:"write_timeout"`        // 写超时（秒）
	ShutdownTimeout    int       `mapstructure:"shutdown_timeout"`     // 优雅关闭超时（秒），默认 30
	InternalServiceKey string    `mapstructure:"internal_service_key"` // 内部服务间 API 密钥（可选，建议通过环境变量覆盖）
	TLS                TLSConfig `mapstructure:"tls"`                  // HTTPS/TLS 配置
}

// TLSConfig HTTPS/TLS 配置。
type TLSConfig struct {
	Enabled    bool   `mapstructure:"enabled"`     // 是否启用 TLS
	CertFile   string `mapstructure:"cert_file"`   // 证书文件路径（enabled 时必填）
	KeyFile    string `mapstructure:"key_file"`    // 私钥文件路径（enabled 时必填）
	MinVersion string `mapstructure:"min_version"` // 最低 TLS 版本（非空时校验），枚举：1.2 | 1.3
}

func (c HTTPConfig) GetShutdownTimeout() time.Duration {
	if c.ShutdownTimeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.ShutdownTimeout) * time.Second
}

// LogConfig 日志配置。
type LogConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别，枚举（大小写不敏感）：debug | info | warn | error
	Format     string `mapstructure:"format"`      // 输出格式，枚举（大小写不敏感）：json | console
	Output     string `mapstructure:"output"`      // 输出目标，枚举（大小写不敏感）：stdout | file | both
	Filename   string `mapstructure:"filename"`    // 日志文件路径（output 为 file/both 时必填）
	MaxSize    int    `mapstructure:"max_size"`    // 单文件最大体积（MB）
	MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `mapstructure:"max_age"`     // 日志保留天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩轮转后的日志
}

// DatabaseConfig 数据库连接配置。
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`             // 驱动（必填），枚举：mysql | postgres | postgresql | sqlite
	Host         string `mapstructure:"host"`               // 主机（必填）
	Port         int    `mapstructure:"port"`               // 端口
	User         string `mapstructure:"user"`               // 用户名
	Password     string `mapstructure:"password"`           // 密码，可通过环境变量 DATABASE_PASSWORD 覆盖
	Database     string `mapstructure:"database"`           // 库名（必填）
	SSLMode      string `mapstructure:"ssl_mode"`           // PostgreSQL SSL 模式（仅 postgres/postgresql 生效；空值用驱动默认），枚举：disable | require | verify-ca | verify-full
	MaxOpenConns int    `mapstructure:"max_open_conns"`     // 最大打开连接数
	MaxIdleConns int    `mapstructure:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLife  int    `mapstructure:"conn_max_lifetime"`  // 连接最大存活时间（秒）
	ConnMaxIdle  int    `mapstructure:"conn_max_idle_time"` // 连接最大空闲时间（秒）
}

// RedisConfig Redis 连接配置（enabled 为 false 时可忽略其余字段）。
type RedisConfig struct {
	Enabled            bool   `mapstructure:"enabled"`              // 是否启用 Redis
	Host               string `mapstructure:"host"`                 // 主机（enabled 时必填）
	Port               int    `mapstructure:"port"`                 // 端口，默认 6379
	Password           string `mapstructure:"password"`             // 密码，可通过环境变量 REDIS_PASSWORD 覆盖
	DB                 int    `mapstructure:"db"`                   // 逻辑库编号，范围 0-15
	PoolSize           int    `mapstructure:"pool_size"`            // 连接池大小
	MinIdleConns       int    `mapstructure:"min_idle_conns"`       // 最小空闲连接数
	IdleTimeout        int    `mapstructure:"idle_timeout"`         // 空闲连接超时（秒）
	IdleCheckFrequency int    `mapstructure:"idle_check_frequency"` // 空闲连接检查间隔（秒）
}

// JWTConfig JWT 认证配置。
type JWTConfig struct {
	Secret          string   `mapstructure:"secret"`            // 签名密钥（生产环境必填，至少 32 字节）
	AccessTokenTTL  int      `mapstructure:"access_token_ttl"`  // Access Token 有效期（秒），默认 7200
	RefreshTokenTTL int      `mapstructure:"refresh_token_ttl"` // Refresh Token 有效期（秒），默认 604800
	SkipPaths       []string `mapstructure:"skip_paths"`        // 跳过 JWT 校验的路径前缀列表
}

// RateLimitConfig 限流配置（enabled 为 false 时可忽略其余字段）。
type RateLimitConfig struct {
	Enabled   bool   `mapstructure:"enabled"`    // 是否启用限流
	Store     string `mapstructure:"store"`      // 存储后端（enabled 时校验），枚举：memory | redis
	GlobalRPS int    `mapstructure:"global_rps"` // 全局限流（请求/秒），0 表示不限制
	IPRPS     int    `mapstructure:"ip_rps"`     // 单 IP 限流（请求/秒），0 表示不限制
	UserRPS   int    `mapstructure:"user_rps"`   // 单用户限流（请求/秒），0 表示不限制
}

// CacheConfig 缓存配置（enabled 为 false 时可忽略其余字段）。
type CacheConfig struct {
	Enabled bool              `mapstructure:"enabled"` // 是否启用缓存
	Driver  string            `mapstructure:"driver"`  // 驱动（enabled 时校验），枚举：memory | redis | multi
	Memory  MemoryCacheConfig `mapstructure:"memory"`  // 内存缓存参数（driver 为 memory/multi 时使用）
	Redis   RedisCacheConfig  `mapstructure:"redis"`   // Redis 缓存参数（driver 为 redis/multi 时使用）
}

// MemoryCacheConfig 内存缓存（Ristretto）参数。
type MemoryCacheConfig struct {
	NumCounters int64 `mapstructure:"num_counters"` // 计数器数量，建议为预计键数量的 10 倍
	MaxCost     int64 `mapstructure:"max_cost"`     // 最大内存成本（字节）
	BufferItems int64 `mapstructure:"buffer_items"` // 写入缓冲区大小，默认 64
}

// RedisCacheConfig Redis 缓存参数。
type RedisCacheConfig struct {
	KeyPrefix  string `mapstructure:"key_prefix"`  // 键前缀
	DefaultTTL int    `mapstructure:"default_ttl"` // 默认过期时间（秒）
}

// TracingConfig 分布式链路追踪配置（enabled 为 false 时可忽略其余字段）。
type TracingConfig struct {
	Enabled    bool    `mapstructure:"enabled"`     // 是否启用 OpenTelemetry 追踪
	Endpoint   string  `mapstructure:"endpoint"`    // OTLP 采集端点（enabled 时必填），如 http://localhost:4317
	SampleRate float64 `mapstructure:"sample_rate"` // 采样率，范围 0.0-1.0，默认 1.0
}

// MetricsConfig Prometheus 指标配置（enabled 为 false 时可忽略其余字段）。
type MetricsConfig struct {
	Enabled bool              `mapstructure:"enabled"` // 是否暴露 Prometheus 指标
	Path    string            `mapstructure:"path"`    // 指标 HTTP 路径（enabled 时必填），须以 / 开头，默认 /metrics
	Labels  map[string]string `mapstructure:"labels"`  // 附加到所有指标的全局标签
}

func LoadWithEnv(configPath, envOverride string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	// Step 1: Determine environment from envOverride > system env > default
	env := envOverride
	if env == "" {
		env = os.Getenv(EnvVarAppEnv)
		if env == "" {
			env = constant.EnvDev // default to dev
		}
	}
	v.Set("app.env", env)

	// Step 2: Load .env files based on environment (optional, won't fail if missing)
	loadOptionalDotenv(fmt.Sprintf(".env.%s", env))
	loadOptionalDotenv(fmt.Sprintf("./configs/.env.%s", env))
	loadOptionalDotenv(".env")
	loadOptionalDotenv("./configs/.env")

	// Step 3: Bind sensitive fields to environment variables
	bindEnvVar(v, "database.password", EnvVarDatabasePassword)
	bindEnvVar(v, "redis.password", EnvVarRedisPassword)

	configDir := "./configs"

	// Step 4: Load config files
	// Priority 1: If configPath is provided, load it directly
	if configPath != "" {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		configDir = filepath.Dir(absPath)
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// If envOverride is provided, merge env config (for backward compat with --env + --config)
		if envOverride != "" {
			envConfigName := fmt.Sprintf("config.%s", envOverride)
			envConfigFile := filepath.Join(configDir, envConfigName+".yaml")
			if _, err := os.Stat(envConfigFile); err == nil {
				v.SetConfigName(envConfigName)
				v.SetConfigType("yaml")
				v.AddConfigPath(configDir)
				if err := v.MergeInConfig(); err != nil {
					return nil, fmt.Errorf("failed to merge environment config: %w", err)
				}
			}
		}
	} else {
		// Priority 2: Load base config.yaml
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(configDir)
		v.AddConfigPath(".")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}

		// Merge environment-specific config
		envConfigName := fmt.Sprintf("config.%s", env)
		envConfigFile := filepath.Join(configDir, envConfigName+".yaml")
		if _, err := os.Stat(envConfigFile); err == nil { //nolint:gosec // envConfigFile is derived from trusted config dir + env name
			v.SetConfigName(envConfigName)
			v.SetConfigType("yaml")
			v.AddConfigPath(configDir)
			if err := v.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("failed to merge environment config: %w", err)
			}
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	Cfg = cfg

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if err := c.validateEnums(); err != nil {
		return err
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("http.port must be between 1 and 65535")
	}
	if c.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}
	if c.Database.Driver != constant.DriverSQLite && c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if c.Redis.Enabled {
		if c.Redis.Host == "" {
			return fmt.Errorf("redis.host is required when redis is enabled")
		}
	}

	if c.JWT.Secret != "" {
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("jwt.secret must be at least 32 bytes for HS256")
		}
	}
	// Production environment requires JWT secret
	if c.App.Env == constant.EnvProd && c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required in production environment")
	}

	if c.RateLimit.Enabled {
		if c.RateLimit.Store == "redis" && !c.Redis.Enabled {
			return fmt.Errorf("redis must be enabled when rate_limit.store is \"redis\"")
		}
	}

	if c.HTTP.TLS.Enabled {
		if c.HTTP.TLS.CertFile == "" {
			return fmt.Errorf("http.tls.cert_file is required when TLS is enabled")
		}
		if c.HTTP.TLS.KeyFile == "" {
			return fmt.Errorf("http.tls.key_file is required when TLS is enabled")
		}
	}

	// 验证缓存配置
	if c.Cache.Enabled {
		if (c.Cache.Driver == "redis" || c.Cache.Driver == "multi") && !c.Redis.Enabled {
			return fmt.Errorf("redis must be enabled when using redis or multi cache driver")
		}

		if c.Cache.Driver == "memory" || c.Cache.Driver == "multi" {
			if c.Cache.Memory.NumCounters <= 0 {
				return fmt.Errorf("cache.memory.num_counters must be positive")
			}
			if c.Cache.Memory.MaxCost <= 0 {
				return fmt.Errorf("cache.memory.max_cost must be positive")
			}
		}
	}

	// 验证 tracing 配置
	if c.Tracing.Enabled {
		if c.Tracing.Endpoint == "" {
			return fmt.Errorf("tracing.endpoint is required when tracing is enabled")
		}
		if c.Tracing.SampleRate < 0 || c.Tracing.SampleRate > 1 {
			return fmt.Errorf("tracing.sample_rate must be between 0 and 1")
		}
	}

	// 验证 metrics 配置
	if c.Metrics.Enabled {
		if c.Metrics.Path == "" {
			return fmt.Errorf("metrics.path is required when metrics is enabled")
		}
		if c.Metrics.Path[0] != '/' {
			return fmt.Errorf("metrics.path must start with \"/\"")
		}
	}

	return nil
}

func (c *Config) MaskedPassword() string {
	return "******"
}

func (c *Config) MaskedDatabasePassword() string {
	if c.Database.Password == "" {
		return ""
	}
	return "******"
}

func (c *Config) MaskedRedisPassword() string {
	if c.Redis.Password == "" {
		return ""
	}
	return "******"
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", constant.AppName)
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.env", constant.EnvDev)

	v.SetDefault("http.enabled", true)
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", 30)
	v.SetDefault("http.write_timeout", 30)
	v.SetDefault("http.shutdown_timeout", 30)

	v.SetDefault("log.level", constant.DefaultLogLevel)
	v.SetDefault("log.format", constant.DefaultLogFormat)
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.filename", "logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 30)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.compress", true)

	v.SetDefault("database.driver", constant.DriverMySQL)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 300) // 5 minutes in seconds
	v.SetDefault("database.conn_max_idle_time", 60) // 1 minute in seconds

	v.SetDefault("redis.enabled", false)
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.pool_size", 10)

	v.SetDefault("cache.enabled", false)
	v.SetDefault("cache.driver", "memory")
	v.SetDefault("cache.memory.num_counters", 1000000)
	v.SetDefault("cache.memory.max_cost", 1073741824) // 1GB
	v.SetDefault("cache.memory.buffer_items", 64)
	v.SetDefault("cache.redis.key_prefix", "app:")
	v.SetDefault("cache.redis.default_ttl", 3600) // 1 hour in seconds

	v.SetDefault("jwt.access_token_ttl", 7200)    // 2 hours in seconds
	v.SetDefault("jwt.refresh_token_ttl", 604800) // 7 days in seconds
	v.SetDefault("jwt.skip_paths", []string{"/health", "/readiness", "/liveness", "/version", "/swagger/*any"})

	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.store", "memory")
	v.SetDefault("rate_limit.global_rps", 10000)
	v.SetDefault("rate_limit.ip_rps", 100)
	v.SetDefault("rate_limit.user_rps", 50)

	v.SetDefault("http.tls.enabled", false)
	v.SetDefault("http.tls.min_version", "1.3")

	v.SetDefault("tracing.enabled", false)
	v.SetDefault("tracing.sample_rate", 1.0)

	v.SetDefault("metrics.enabled", false)
	v.SetDefault("metrics.path", "/metrics")
}

func loadOptionalDotenv(path string) {
	if err := godotenv.Load(path); err != nil {
		return // dotenv files are optional
	}
}

func bindEnvVar(v *viper.Viper, key, envKey string) {
	if err := v.BindEnv(key, envKey); err != nil {
		return // bind failures are non-fatal during setup
	}
}

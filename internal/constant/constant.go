// Package constant defines application-wide constants.
package constant

// Application constants
const (
	// AppName is the application name
	AppName = "orange-tv"

	// DefaultHTTPPort is the default HTTP server port
	DefaultHTTPPort = 8080

	// DefaultHTTPHost is the default HTTP server host
	DefaultHTTPHost = "0.0.0.0"

	// DefaultLogLevel is the default log level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default log format
	DefaultLogFormat = "json"
)

// Error code constants
const (
	// SuccessCode represents success
	SuccessCode = 0
)

// Environment constants
const (
	// EnvDev represents development environment
	EnvDev = "dev"

	// EnvProd represents production environment
	EnvProd = "prod"

	// EnvTest represents test environment
	EnvTest = "test"
)

// Time constants
const (
	// DefaultHTTPTimeout is the default HTTP timeout
	DefaultHTTPTimeout = 30 // seconds

	// DefaultShutdownTimeout is the default graceful shutdown timeout
	DefaultShutdownTimeout = 30 // seconds
)

// Header constants
const (
	// HeaderRequestID is the HTTP header for request ID
	HeaderRequestID = "X-Request-ID"

	// HeaderContentType is the HTTP header for content type
	HeaderContentType = "Content-Type"

	// HeaderAuthorization is the HTTP header for authorization
	HeaderAuthorization = "Authorization"
)

// Database driver constants
const (
	// DriverMySQL represents MySQL database driver
	DriverMySQL = "mysql"

	// DriverPostgreSQL represents PostgreSQL database driver
	DriverPostgreSQL = "postgres"

	// DriverPostgres is an alias for PostgreSQL driver
	DriverPostgres = "postgresql"

	// DriverSQLite represents SQLite database driver (testing only)
	DriverSQLite = "sqlite"

	// DriverSQLite3 is an alias for SQLite driver (testing only)
	DriverSQLite3 = "sqlite3"
)

// Log level constants
const (
	// LogLevelDebug represents debug log level
	LogLevelDebug = "debug"

	// LogLevelInfo represents info log level
	LogLevelInfo = "info"

	// LogLevelWarn represents warn log level
	LogLevelWarn = "warn"

	// LogLevelError represents error log level
	LogLevelError = "error"
)

// Log format constants
const (
	// LogFormatJSON represents JSON log format
	LogFormatJSON = "json"

	// LogFormatConsole represents console log format
	LogFormatConsole = "console"
)

// Log output constants
const (
	// LogOutputStdout represents stdout log output
	LogOutputStdout = "stdout"

	// LogOutputFile represents file log output
	LogOutputFile = "file"

	// LogOutputBoth represents both stdout and file log output
	LogOutputBoth = "both"
)

// Status constants for enable/disable fields.
const (
	StatusDisabled int8 = 0
	StatusEnabled  int8 = 1
)

// Publish status for videos.
const (
	PublishStatusOffline int8 = 0
	PublishStatusOnline  int8 = 1
)

// Serial status for videos.
const (
	SerialStatusOngoing  int8 = 1
	SerialStatusFinished int8 = 2
	SerialStatusUpcoming int8 = 3
)

// Auth / RBAC constants for phase 2.
const (
	// RoleSuperAdmin is the only preset role in phase 2.
	RoleSuperAdmin = "super_admin"

	// PermissionAll grants full admin access.
	PermissionAll = "*"
)

// Play formats accepted by play_episodes.format.
const (
	PlayFormatHLS  = "hls"
	PlayFormatMP4  = "mp4"
	PlayFormatDASH = "dash"
	PlayFormatFLV  = "flv"
)

// Collect source formats (collect_sources.type).
const (
	CollectTypeDefault  int8 = 1 // system JSON format
	CollectTypeAppleCMS int8 = 2 // 苹果 CMS
)

// Collect log status (collect_logs.status).
const (
	CollectLogSuccess        int8 = 1
	CollectLogFailed         int8 = 2
	CollectLogPartialSuccess int8 = 3
	CollectLogRunning        int8 = 4
	CollectLogCancelled      int8 = 5
)

// System setting keys.
const (
	SettingSiteMode                = "site_mode"
	SettingAPIOutputFormat         = "api_output_format"
	SettingEnableThirdPartyCollect = "enable_third_party_collect"
	SettingSiteName                = "site_name"
	SettingSiteLogo                = "site_logo"
	SettingSiteCopyright           = "site_copyright"
	SettingSiteICP                 = "site_icp"
	SettingSiteSEOKeywords         = "site_seo_keywords"
	SettingSiteDescription         = "site_description"
	SettingResourceAPIKey          = "resource_api_key"
)

// Site mode values.
const (
	SiteModeVideoSite    = "video_site"
	SiteModeResourceSite = "resource_site"
)

// API output format values for resource open API.
// Default is the system native format; apple_cms is the only optional alternative.
const (
	APIOutputDefault  = "default"   // 系统默认/自有 JSON 格式
	APIOutputAppleCMS = "apple_cms" // 苹果 CMS 兼容
)

// Login log user types / status.
const (
	LoginUserTypeAdmin int8 = 1
	LoginUserTypeUser  int8 = 2

	LoginStatusSuccess int8 = 1
	LoginStatusFailed  int8 = 2
)

// System log levels.
const (
	SystemLogLevelInfo     int8 = 1
	SystemLogLevelWarning  int8 = 2
	SystemLogLevelError    int8 = 3
	SystemLogLevelCritical int8 = 4
)

// Setting type values (system_settings.setting_type).
const (
	SettingTypeString  int8 = 1
	SettingTypeNumber  int8 = 2
	SettingTypeBoolean int8 = 3
	SettingTypeJSON    int8 = 4
)

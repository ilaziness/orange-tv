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

// Log output constants
const (
	// LogOutputStdout represents stdout log output
	LogOutputStdout = "stdout"

	// LogOutputFile represents file log output
	LogOutputFile = "file"

	// LogOutputBoth represents both stdout and file log output
	LogOutputBoth = "both"
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

// System setting keys, group constants, and related enum values
// have been moved to settings.go in this package.

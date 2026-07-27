package utils

import "strings"

var mysqlStringReplacer = strings.NewReplacer(
	"\\", "\\\\",
	"'", "\\'",
	"\x00", "\\0",
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t",
	"\b", "\\b",
)

// EscapeMySQLString escapes a string for use in MySQL string literals.
func EscapeMySQLString(s string) string {
	return mysqlStringReplacer.Replace(s)
}

// QuoteMySQL wraps a MySQL identifier in backticks, escaping embedded backticks.
func QuoteMySQL(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// EscapePGString escapes a string for use in PostgreSQL string literals.
func EscapePGString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// QuotePG wraps a PostgreSQL identifier in double quotes, escaping embedded double quotes.
func QuotePG(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// IsMySQLBinaryType reports whether a MySQL data type is a binary/blob type.
func IsMySQLBinaryType(dataType string) bool {
	switch strings.ToLower(dataType) {
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		return true
	}
	return false
}

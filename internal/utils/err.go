package utils

import "strings"

// IsDuplicateKey reports whether err is a database duplicate/unique constraint error.
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

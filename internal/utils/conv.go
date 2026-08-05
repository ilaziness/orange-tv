package utils

import (
	"strconv"
	"strings"
)

// StringToInt converts a string to int, returning 0 on parse failure or empty input.
// Whitespace is trimmed before parsing. Intended for best-effort fields (years,
// durations, counts) where an invalid value should fall back to 0 rather than error.
func StringToInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

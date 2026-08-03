package service

import (
	"strconv"
	"strings"

	"github.com/ilaziness/orange-tv/internal/model"
)

// StrVal reads a string setting from the map, returning "" if missing.
func StrVal(m map[string]model.SystemSettings, key string) string {
	it, ok := m[key]
	if !ok {
		return ""
	}
	return it.SettingValue
}

// IntVal reads an integer setting from the map, returning def if missing or invalid.
func IntVal(m map[string]model.SystemSettings, key string, def int) int {
	v := strings.TrimSpace(StrVal(m, key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// BoolVal reads a boolean setting from the map, returning def if missing or invalid.
func BoolVal(m map[string]model.SystemSettings, key string, def bool) bool {
	v := strings.TrimSpace(StrVal(m, key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		if n, err := strconv.Atoi(v); err == nil {
			return n != 0
		}
		return def
	}
}

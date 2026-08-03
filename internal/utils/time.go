package utils

import (
	"strings"
	"time"
)

// FormatTimeStr safely formats a *time.Time as DateTime string, returning "" for nil/zero.
func FormatTimeStr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}

// DataRangeCutoff returns the earliest time for the given data range.
// Supported values: today/last1d/last3d/last1w/last1m. Returns nil for "all" or empty (no filtering).
func DataRangeCutoff(dataRange string) *time.Time {
	now := time.Now()
	switch strings.TrimSpace(dataRange) {
	case "today":
		t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return &t
	case "last1d":
		t := now.AddDate(0, 0, -1)
		return &t
	case "last3d":
		t := now.AddDate(0, 0, -3)
		return &t
	case "last1w":
		t := now.AddDate(0, 0, -7)
		return &t
	case "last1m":
		t := now.AddDate(0, -1, 0)
		return &t
	default:
		return nil
	}
}

// ParseFlexibleDate tries to parse v using each layout in order, returning the first success.
func ParseFlexibleDate(v string, layouts []string) (time.Time, error) {
	var last error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, v, time.Local)
		if err == nil {
			return t, nil
		}
		last = err
	}
	return time.Time{}, last
}

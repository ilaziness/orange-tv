package utils

import "time"

// FormatTimeStr safely formats a *time.Time as DateTime string, returning "" for nil/zero.
func FormatTimeStr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
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

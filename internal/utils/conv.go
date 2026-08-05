package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// AnyToInt converts an any value to int, handling common JSON types.
func AnyToInt(v any) int {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return int(t)
	case float32:
		return int(t)
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint:
		return Uint64ToInt(uint64(t))
	case uint32:
		return int(t)
	case uint64:
		return Uint64ToInt(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(t))
		return i
	default:
		s := fmt.Sprint(t)
		i, _ := strconv.Atoi(strings.TrimSpace(s))
		return i
	}
}

// AnyToString converts an any value to string, handling common JSON types.
func AnyToString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

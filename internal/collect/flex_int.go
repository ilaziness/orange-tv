package collect

import (
	"encoding/json"
	"strconv"
	"strings"
)

// flexInt unmarshals from either a JSON number or a numeric string.
// Apple CMS is a fragmented ecosystem where different installations (and even
// different endpoints on the same site) return the same field as int in some
// responses and as string in others (e.g. list: "page":"1", detail: "page":1).
type flexInt int

// UnmarshalJSON implements json.Unmarshaler for flexInt.
// It accepts JSON numbers, numeric strings, and null (treated as 0).
func (f *flexInt) UnmarshalJSON(data []byte) error {
	// null → 0
	if string(data) == "null" {
		*f = 0
		return nil
	}
	// quoted string: strip quotes and parse
	if len(data) > 0 && data[0] == '"' {
		s := strings.TrimSpace(string(data[1 : len(data)-1]))
		if s == "" {
			*f = 0
			return nil
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			// non-numeric string → 0 (best-effort)
			*f = 0
			return nil
		}
		*f = flexInt(i)
		return nil
	}
	// plain JSON number
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		*f = 0
		return nil
	}
	i, err := n.Int64()
	if err != nil {
		*f = 0
		return nil
	}
	*f = flexInt(i)
	return nil
}

// int returns the flexInt as a plain int.
func (f flexInt) int() int {
	return int(f)
}

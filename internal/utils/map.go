package utils

// ExtractString safely extracts a string value from a map[string]any by key.
func ExtractString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

package utils

import (
	"regexp"
	"strings"
)

var htmlTagRegexp = regexp.MustCompile(`<[^>]*>`)

// StripHTMLTags removes HTML tags from s and trims whitespace.
func StripHTMLTags(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = htmlTagRegexp.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// SplitNames splits multiple name strings by common delimiters and returns deduplicated names.
func SplitNames(parts ...string) []string {
	out := make([]string, 0)
	seen := map[string]bool{}
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		s = strings.ReplaceAll(s, "，", ",")
		s = strings.ReplaceAll(s, "、", ",")
		s = strings.ReplaceAll(s, "/", ",")
		s = strings.ReplaceAll(s, "|", ",")
		commaParts := strings.Split(s, ",")
		for _, cp := range commaParts {
			spaceParts := strings.Fields(cp)
			for _, p := range spaceParts {
				p = strings.TrimSpace(p)
				if p == "" || seen[p] {
					continue
				}
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out
}

// FirstNonEmpty returns the first non-empty, non-"0" string from vals, trimmed.
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" && v != "0" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// DefaultStr returns v if it is non-empty (after trimming), otherwise def.
func DefaultStr(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

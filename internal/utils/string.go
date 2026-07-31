package utils

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	htmlTagRegexp = regexp.MustCompile(`<[^>]*>`)
	multiSpaceReg = regexp.MustCompile(` +`)
	// zeroWidthRunes are format/zero-width characters that carry no visible
	// glyph. They are not Unicode whitespace, so they must be removed rather
	// than collapsed into a space (which would introduce spurious spaces).
	zeroWidthRunes = map[rune]bool{
		'\u200B': true, // ZERO WIDTH SPACE
		'\u200C': true, // ZERO WIDTH NON-JOINER
		'\u200D': true, // ZERO WIDTH JOINER
		'\u2060': true, // WORD JOINER
		'\uFEFF': true, // ZERO WIDTH NO-BREAK SPACE (BOM)
	}
)

// StripHTMLTags removes HTML tags and HTML entities (such as &nbsp;) from s,
// collapses consecutive whitespace into a single space, and trims the result.
//
// Tags are stripped before entity decoding so that escaped markup
// (e.g. "&lt;script&gt;") is preserved as literal text rather than being
// removed as a tag. Standard entities are decoded via html.UnescapeString;
// unknown "&xxx;" sequences are kept as-is to avoid destroying legitimate
// text such as "Tom & Jerry;".
func StripHTMLTags(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = htmlTagRegexp.ReplaceAllString(s, "")
	// Decode known HTML entities (&nbsp;, &amp;, &#160;, ...). &nbsp; becomes U+00A0.
	s = html.UnescapeString(s)
	// Normalize every Unicode whitespace rune (incl. NBSP, NEL, line/paragraph
	// separators, ideographic space, ...) to a single ASCII space and drop
	// zero-width / format characters that would otherwise linger.
	s = strings.Map(func(r rune) rune {
		if zeroWidthRunes[r] {
			return -1
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
	// Collapse runs of ASCII spaces produced above into one space.
	s = multiSpaceReg.ReplaceAllString(s, " ")
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

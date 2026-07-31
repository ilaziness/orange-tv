package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStripHTMLTagsUnicodeWhitespace verifies that all Unicode whitespace
// characters (including NEL, line/paragraph separators, NBSP, ideographic
// space) are collapsed to a single ASCII space, and that zero-width
// characters are removed rather than turned into spaces.
func TestStripHTMLTagsUnicodeWhitespace(t *testing.T) {
	// Each Unicode whitespace rune should collapse to a single space.
	require.Equal(t, "a b", StripHTMLTags("a\u00A0b")) // NBSP (&nbsp;)
	require.Equal(t, "a b", StripHTMLTags("a\u1680b")) // OGHAM SPACE MARK
	require.Equal(t, "a b", StripHTMLTags("a\u2000b")) // EN QUAD
	require.Equal(t, "a b", StripHTMLTags("a\u2009b")) // THIN SPACE
	require.Equal(t, "a b", StripHTMLTags("a\u202Fb")) // NARROW NBSP
	require.Equal(t, "a b", StripHTMLTags("a\u205Fb")) // MEDIUM MATHEMATICAL SPACE
	require.Equal(t, "a b", StripHTMLTags("a\u3000b")) // IDEOGRAPHIC SPACE
	require.Equal(t, "a b", StripHTMLTags("a\u0085b")) // NEL (Next Line)
	require.Equal(t, "a b", StripHTMLTags("a\u2028b")) // LINE SEPARATOR
	require.Equal(t, "a b", StripHTMLTags("a\u2029b")) // PARAGRAPH SEPARATOR
	require.Equal(t, "a b", StripHTMLTags("a\v\fb"))   // VERTICAL TAB, FORM FEED

	// Consecutive mixed whitespace collapses to a single space.
	require.Equal(t, "a b", StripHTMLTags("a \u00A0\u3000\nb"))

	// Zero-width characters are removed (not turned into spaces).
	require.Equal(t, "ab", StripHTMLTags("a\u200Bb")) // ZERO WIDTH SPACE
	require.Equal(t, "ab", StripHTMLTags("a\u200Cb")) // ZERO WIDTH NON-JOINER
	require.Equal(t, "ab", StripHTMLTags("a\u200Db")) // ZERO WIDTH JOINER
	require.Equal(t, "ab", StripHTMLTags("a\u2060b")) // WORD JOINER
	require.Equal(t, "ab", StripHTMLTags("a\uFEFFb")) // BOM / ZERO WIDTH NO-BREAK SPACE
}

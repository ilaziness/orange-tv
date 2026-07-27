package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripHTMLTags(t *testing.T) {
	require.Equal(t, "纯文本", StripHTMLTags("<p>纯文本</p>"))
	require.Equal(t, "多行文本", StripHTMLTags("<div>多行<br>文本</div>"))
	require.Equal(t, "", StripHTMLTags(""))
	require.Equal(t, "无标签", StripHTMLTags("无标签"))
}

func TestSplitNamesWithSpaces(t *testing.T) {
	result := SplitNames("演员A 演员B,演员C")
	require.Equal(t, []string{"演员A", "演员B", "演员C"}, result)

	result = SplitNames("演员A　演员B")
	require.Equal(t, []string{"演员A", "演员B"}, result)

	result = SplitNames("演员A,演员B、演员C/演员D")
	require.Equal(t, []string{"演员A", "演员B", "演员C", "演员D"}, result)
}

func TestDefaultStr(t *testing.T) {
	require.Equal(t, "val", DefaultStr("val", "def"))
	require.Equal(t, "def", DefaultStr("", "def"))
	require.Equal(t, "def", DefaultStr("  ", "def"))
}

func TestFirstNonEmpty(t *testing.T) {
	require.Equal(t, "a", FirstNonEmpty("", "a", "b"))
	require.Equal(t, "b", FirstNonEmpty("0", "b"))
	require.Equal(t, "", FirstNonEmpty("", "0"))
}

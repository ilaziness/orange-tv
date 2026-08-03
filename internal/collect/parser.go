package collect

import (
	"strconv"
	"strings"
)

// extractEpisodeNumber extracts a positive episode number from a title like
// "第12集" or "12". Returns 0 if no number is found.
func extractEpisodeNumber(s string) int32 {
	s = strings.TrimSpace(s)
	// 第12集 / 12
	var digits strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else if digits.Len() > 0 {
			break
		}
	}
	if digits.Len() == 0 {
		return 0
	}
	n, err := strconv.Atoi(digits.String())
	if err != nil || n <= 0 {
		return 0
	}
	return int32(n)
}

// guessFormat infers the play format from the URL extension.
func guessFormat(url string) string {
	u := strings.ToLower(url)
	switch {
	case strings.Contains(u, ".m3u8"):
		return "hls"
	case strings.Contains(u, ".mpd"):
		return "dash"
	case strings.Contains(u, ".flv"):
		return "flv"
	case strings.Contains(u, ".mp4"):
		return "mp4"
	default:
		return "hls"
	}
}

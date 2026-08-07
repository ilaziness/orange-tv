package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxtParser_Parse(t *testing.T) {
	raw := `央视频道,#genre#
CCTV1,http://example.com/cctv1
CCTV2,http://example.com/cctv2
CCTV1,http://example.com/cctv1-dup
更新时间,#genre#
2024-01-01,http://example.com/time
推广,#genre#
支持作者,http://example.com/donate
`

	parser := &txtParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 3)
	assert.Equal(t, "央视频道", entries[0].Category)
	assert.Equal(t, "CCTV1", entries[0].Name)
	assert.Equal(t, "http://example.com/cctv1", entries[0].StreamURL)
	assert.Equal(t, "CCTV2", entries[1].Name)
	// Duplicate name in same category gets suffix
	assert.Equal(t, "CCTV1(2)", entries[2].Name)
	// "更新时间" category filtered, "支持作者" name filtered
	for _, e := range entries {
		assert.NotEqual(t, "更新时间", e.Category)
		assert.NotContains(t, e.Name, "支持作者")
	}
}

func TestTxtParser_ParseEmpty(t *testing.T) {
	parser := &txtParser{}
	entries := parser.Parse("")
	assert.Empty(t, entries)
}

func TestM3UParser_Parse_FormatA(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 tvg-name="CCTV1" tvg-id="185" tvg-logo="https://example.com/logo1.png" group-title="央视频道", CCTV1
http://example.com/cctv1
#EXTINF:-1 tvg-name="CCTV2" tvg-id="186" tvg-logo="https://example.com/logo2.png" group-title="央视频道", CCTV2
http://example.com/cctv2
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 2)
	assert.Equal(t, "央视频道", entries[0].Category)
	assert.Equal(t, "CCTV1", entries[0].Name)
	assert.Equal(t, "http://example.com/cctv1", entries[0].StreamURL)
	assert.Equal(t, "https://example.com/logo1.png", entries[0].Logo)
	assert.Equal(t, "CCTV2", entries[1].Name)
	assert.Equal(t, "https://example.com/logo2.png", entries[1].Logo)
}

func TestM3UParser_Parse_FormatB(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="卫视频道", 湖南卫视
http://example.com/hntv
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 1)
	assert.Equal(t, "卫视频道", entries[0].Category)
	assert.Equal(t, "湖南卫视", entries[0].Name)
	assert.Equal(t, "http://example.com/hntv", entries[0].StreamURL)
	assert.Equal(t, "https://example.com/logo.png", entries[0].Logo)
}

func TestM3UParser_Parse_FilterPromotion(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="✈️ TG频道@stymei", TG推广
http://example.com/tg
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="☁️CloudFlare直播", CF测试
http://example.com/cf
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="更新时间", 时间
http://example.com/time
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="央视频道", CCTV1
http://example.com/cctv1
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 1)
	assert.Equal(t, "央视频道", entries[0].Category)
	assert.Equal(t, "CCTV1", entries[0].Name)
}

func TestM3UParser_Parse_RetainNormalContent(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="体育-今天08-07", 赛事回放
http://example.com/sports
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="🔯API随机点播", 随机点播
http://example.com/random
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="✡️周杰伦歌曲点播", 晴天
http://example.com/music
#EXTINF:-1 tvg-logo="https://example.com/logo.png" group-title="🔯歌手合集点播", 歌手合集
http://example.com/singer
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 4)
	assert.Equal(t, "体育-今天08-07", entries[0].Category)
	assert.Equal(t, "🔯API随机点播", entries[1].Category)
	assert.Equal(t, "✡️周杰伦歌曲点播", entries[2].Category)
	assert.Equal(t, "🔯歌手合集点播", entries[3].Category)
}

func TestM3UParser_Parse_DuplicateNamePerCategory(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 group-title="央视频道", CCTV1
http://example.com/cctv1-a
#EXTINF:-1 group-title="央视频道", CCTV1
http://example.com/cctv1-b
#EXTINF:-1 group-title="卫视频道", CCTV1
http://example.com/cctv1-c
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 3)
	assert.Equal(t, "CCTV1", entries[0].Name)
	assert.Equal(t, "CCTV1(2)", entries[1].Name)
	// Same name in different category should NOT get suffix
	assert.Equal(t, "CCTV1", entries[2].Name)
}

func TestM3UParser_Parse_FilterSupportAuthor(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 group-title="央视频道", 支持作者
http://example.com/donate
#EXTINF:-1 group-title="央视频道", CCTV1
http://example.com/cctv1
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 1)
	assert.Equal(t, "CCTV1", entries[0].Name)
}

func TestM3UParser_Parse_EmptyGroupTitle(t *testing.T) {
	raw := `#EXTM3U
#EXTINF:-1 , No Category
http://example.com/nocat
#EXTINF:-1 group-title="央视频道", CCTV1
http://example.com/cctv1
`

	parser := &m3uParser{}
	entries := parser.Parse(raw)

	require.Len(t, entries, 1)
	assert.Equal(t, "CCTV1", entries[0].Name)
}

func TestM3UParser_Parse_InvalidM3U(t *testing.T) {
	parser := &m3uParser{}
	entries := parser.Parse("not a valid m3u")
	// Should not panic, may return nil or empty
	assert.Empty(t, entries)
}

func TestSelectParser(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"m3u extension", "https://example.com/live.m3u", false},
		{"txt extension", "https://example.com/live.txt", false},
		{"m3u with query params", "https://example.com/live.m3u?token=abc", false},
		{"txt with query params", "https://example.com/live.txt?token=abc", false},
		{"uppercase M3U", "https://example.com/live.M3U", false},
		{"uppercase TXT", "https://example.com/live.TXT", false},
		{"unsupported extension", "https://example.com/live.xml", true},
		{"no extension", "https://example.com/live", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := selectParser(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, parser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parser)
			}
		})
	}
}

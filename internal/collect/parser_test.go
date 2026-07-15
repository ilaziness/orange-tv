package collect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAppleCMS(t *testing.T) {
	body := []byte(`{
		"code":1,"page":1,"pagecount":2,"total":2,
		"list":[{
			"vod_id":"123","type_id":"1","vod_name":"测试影片","vod_sub":"副标题",
			"blurb":"简介","pic":"http://x/a.jpg","actor":"演员1,演员2","director":"导演A",
			"vod_year":"2024","vod_area":"美国","vod_lang":"英语","vod_duration":"120",
			"douban_score":"8.5","vod_play_url":"第1集$http://a.m3u8#第2集$http://b.m3u8"
		}]
	}`)
	page, err := ParseAppleCMS(body)
	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 2, page.PageCount)
	require.Len(t, page.Items, 1)
	it := page.Items[0]
	require.Equal(t, "测试影片", it.Title)
	require.Equal(t, int32(2024), it.Year)
	require.Equal(t, "1", it.CategoryKey)
	require.Len(t, it.Episodes, 2)
	require.Equal(t, "http://a.m3u8", it.Episodes[0].URL)
	require.Equal(t, "hls", it.Episodes[0].Format)
	require.Equal(t, []string{"导演A"}, it.Directors)
	require.Equal(t, []string{"演员1", "演员2"}, it.Actors)
}

func TestParseDefaultJSON(t *testing.T) {
	body := []byte(`{
		"code":0,"page":1,"page_count":1,"total":1,
		"list":[{
			"title":"系统格式片","year":2023,"category":"动作",
			"play_urls":[{"episode":1,"title":"全集","url":"http://c.mp4","format":"mp4"}]
		}]
	}`)
	page, err := ParseDefaultJSON(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "系统格式片", page.Items[0].Title)
	require.Equal(t, "动作", page.Items[0].CategoryKey)
	require.Equal(t, "mp4", page.Items[0].Episodes[0].Format)
}

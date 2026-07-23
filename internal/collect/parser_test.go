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
		}],
		"class":[
			{"type_id":1,"type_pid":0,"type_name":"电影片"},
			{"type_id":"6","type_pid":"1","type_name":"动作片"}
		]
	}`)
	page, err := ParseAppleCMS(body)
	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 2, page.PageCount)
	require.Len(t, page.Items, 1)
	it := page.Items[0]
	require.Equal(t, "测试影片", it.Title)
	require.Equal(t, int32(2024), it.Year)
	require.Equal(t, int64(1), it.ExternalCategoryID)
	require.Len(t, it.Episodes, 2)
	require.Equal(t, "http://a.m3u8", it.Episodes[0].URL)
	require.Equal(t, "hls", it.Episodes[0].Format)
	require.Equal(t, []string{"导演A"}, it.Directors)
	require.Equal(t, []string{"演员1", "演员2"}, it.Actors)

	require.Len(t, page.Classes, 2)
	require.Equal(t, int64(1), page.Classes[0].TypeID)
	require.Equal(t, "电影片", page.Classes[0].TypeName)
	require.Equal(t, int64(0), page.Classes[0].TypePID)
	require.Equal(t, int64(6), page.Classes[1].TypeID)
	require.Equal(t, int64(1), page.Classes[1].TypePID)
}

func TestParseAppleCMSNumericIDs(t *testing.T) {
	body := []byte(`{
		"code":1,"page":1,"pagecount":1,"total":1,
		"list":[{
			"vod_id":120334,"type_id":29,"vod_name":"动态漫","vod_play_url":"第1集$http://a.m3u8"
		}],
		"class":[{"type_id":29,"type_pid":4,"type_name":"国产动漫"}]
	}`)
	page, err := ParseAppleCMS(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "120334", page.Items[0].ExternalID)
	require.Equal(t, int64(29), page.Items[0].ExternalCategoryID)
	require.Len(t, page.Classes, 1)
	require.Equal(t, int64(29), page.Classes[0].TypeID)
	require.Equal(t, int64(4), page.Classes[0].TypePID)
	require.Equal(t, "国产动漫", page.Classes[0].TypeName)
}

func TestParseDefaultJSON(t *testing.T) {
	body := []byte(`{
		"code":0,"page":1,"page_count":1,"total":1,
		"list":[{
			"title":"系统格式片","year":2023,"category_id":12,
			"play_urls":[{"episode":1,"title":"全集","url":"http://c.mp4","format":"mp4"}]
		}]
	}`)
	page, err := ParseDefaultJSON(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "系统格式片", page.Items[0].Title)
	require.Equal(t, int64(12), page.Items[0].ExternalCategoryID)
	require.Equal(t, "mp4", page.Items[0].Episodes[0].Format)
}

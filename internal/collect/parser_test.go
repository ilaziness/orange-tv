package collect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAppleCMSList(t *testing.T) {
	body := []byte(`{
		"code":1,"page":1,"pagecount":2,"total":100,
		"list":[
			{"vod_id":123,"vod_time":"2025-07-20 10:30:00"},
			{"vod_id":456,"vod_time":"2025-07-19 12:00:00"}
		],
		"class":[
			{"type_id":1,"type_pid":0,"type_name":"电影片"},
			{"type_id":"6","type_pid":"1","type_name":"动作片"}
		]
	}`)
	lp, err := ParseAppleCMSList(body)
	require.NoError(t, err)
	require.Equal(t, 1, lp.Page)
	require.Equal(t, 2, lp.PageCount)
	require.Equal(t, 100, lp.Total)
	require.Len(t, lp.VodIDs, 2)
	require.Equal(t, int64(123), lp.VodIDs[0])
	require.Equal(t, int64(456), lp.VodIDs[1])
	require.Len(t, lp.VodTimes, 2)
	require.Equal(t, "2025-07-20 10:30:00", lp.VodTimes[0])

	require.Len(t, lp.Classes, 2)
	require.Equal(t, int64(1), lp.Classes[0].TypeID)
	require.Equal(t, "电影片", lp.Classes[0].TypeName)
	require.Equal(t, int64(6), lp.Classes[1].TypeID)
}

func TestParseAppleCMSDetail(t *testing.T) {
	body := []byte(`{
		"code":1,"page":1,"pagecount":1,"total":1,
		"list":[{
			"vod_id":"123","type_id":"1","vod_name":"测试影片","vod_sub":"副标题",
			"vod_content":"<p>这是描述</p>","vod_pic":"http://x/a.jpg","vod_actor":"演员1,演员2",
			"vod_director":"导演A","vod_year":"2024","vod_area":"美国","vod_lang":"英语",
			"vod_duration":"120","douban_score":"8.5","vod_remarks":"更新至第10集",
			"vod_pubdate":"2024-01-01","vod_tag":"高清,独家","vod_class":"动作,科幻",
			"vod_play_url":"第1集$http://a.m3u8#第2集$http://b.m3u8$$$第1集$http://c.mp4"
		}],
		"class":[
			{"type_id":1,"type_pid":0,"type_name":"电影片"}
		]
	}`)
	page, err := ParseAppleCMSDetail(body)
	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Len(t, page.Items, 1)
	it := page.Items[0]
	require.Equal(t, "测试影片", it.Title)
	require.Equal(t, int32(2024), it.Year)
	require.Equal(t, int64(1), it.ExternalCategoryID)
	require.Equal(t, "这是描述", it.Description)
	require.Equal(t, "更新至第10集", it.Remarks)
	// only m3u8 group should be parsed, mp4 group should be skipped
	require.Len(t, it.Episodes, 2)
	require.Equal(t, "http://a.m3u8", it.Episodes[0].URL)
	require.Equal(t, "hls", it.Episodes[0].Format)
	require.Equal(t, []string{"导演A"}, it.Directors)
	require.Equal(t, []string{"演员1", "演员2"}, it.Actors)
	require.Equal(t, []string{"高清", "独家", "动作", "科幻"}, it.Tags)
	require.Equal(t, "http://x/a.jpg", it.Cover)
	require.Equal(t, "2024-01-01", it.ReleaseDate)

	require.Len(t, page.Classes, 1)
	require.Equal(t, int64(1), page.Classes[0].TypeID)
}

func TestParseAppleCMSDetailNumericIDs(t *testing.T) {
	body := []byte(`{
		"code":1,"page":1,"pagecount":1,"total":1,
		"list":[{
			"vod_id":120334,"type_id":29,"vod_name":"动态漫","vod_play_url":"第1集$http://a.m3u8"
		}],
		"class":[{"type_id":29,"type_pid":4,"type_name":"国产动漫"}]
	}`)
	page, err := ParseAppleCMSDetail(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "120334", page.Items[0].ExternalID)
	require.Equal(t, int64(29), page.Items[0].ExternalCategoryID)
	require.Len(t, page.Classes, 1)
	require.Equal(t, int64(29), page.Classes[0].TypeID)
}

func TestParseApplePlayURLsM3u8Only(t *testing.T) {
	// group without m3u8 should be skipped entirely
	eps := parseApplePlayURLs("第1集$http://a.mp4#第2集$http://b.mp4$$$第1集$http://a.m3u8#第2集$http://b.m3u8")
	require.Len(t, eps, 2)
	require.Equal(t, "http://a.m3u8", eps[0].URL)
	require.Equal(t, "http://b.m3u8", eps[1].URL)
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

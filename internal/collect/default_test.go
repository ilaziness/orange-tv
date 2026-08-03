package collect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDefaultList(t *testing.T) {
	body := []byte(`{
		"code":0,"message":"success",
		"data":{
			"list":[
				{"id":101,"title":"影片A","category_id":5,"created_at":"2025-08-01 10:00:00"},
				{"id":102,"title":"影片B","category_id":6,"created_at":"2025-08-02 11:30:00"}
			],
			"total":2,"page":1,"page_size":50,"total_pages":1
		}
	}`)
	lp, err := parseDefaultList(body)
	require.NoError(t, err)
	require.Equal(t, 1, lp.Page)
	require.Equal(t, 1, lp.PageCount)
	require.Equal(t, 2, lp.Total)
	require.Len(t, lp.IDs, 2)
	require.Equal(t, uint32(101), lp.IDs[0])
	require.Equal(t, uint32(102), lp.IDs[1])
	require.Len(t, lp.Times, 2)
	require.Equal(t, "2025-08-01 10:00:00", lp.Times[0])
}

func TestParseDefaultListEmpty(t *testing.T) {
	body := []byte(`{"code":0,"message":"success","data":{"list":[],"total":0,"page":1,"page_size":50,"total_pages":0}}`)
	lp, err := parseDefaultList(body)
	require.NoError(t, err)
	require.Empty(t, lp.IDs)
	require.Equal(t, 1, lp.PageCount) // normalized to >=1
}

func TestParseDefaultListErrorCode(t *testing.T) {
	body := []byte(`{"code":1,"message":"disabled","data":null}`)
	_, err := parseDefaultList(body)
	require.Error(t, err)
}

func TestParseDefaultDetailTakesFirstSourceOnly(t *testing.T) {
	body := []byte(`{
		"code":0,"message":"success",
		"data":[{
			"id":101,"title":"测试影片","subtitle":"副标题","cover":"http://x/c.jpg",
			"category_id":5,"year":2024,"release_date":"2024-01-01","region":"美国",
			"language":"英语","description":"描述","directors":["导演A"],"actors":["演员1","演员2"],
			"sources":[
				{"id":1,"name":"线路1","episodes":[
					{"episode":1,"title":"第1集","url":"http://a.m3u8"},
					{"episode":2,"title":"第2集","url":"http://b.m3u8"}
				]},
				{"id":2,"name":"线路2","episodes":[
					{"episode":1,"title":"第1集","url":"http://c.mp4"}
				]}
			],
			"created_at":"2025-08-01 10:00:00"
		}]
	}`)
	page, err := parseDefaultDetail(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	it := page.Items[0]
	require.Equal(t, "测试影片", it.Title)
	require.Equal(t, "101", it.ExternalID)
	require.Equal(t, uint32(5), it.ExternalCategoryID)
	require.Equal(t, int32(2024), it.Year)
	require.Equal(t, []string{"导演A"}, it.Directors)
	require.Equal(t, []string{"演员1", "演员2"}, it.Actors)
	require.Equal(t, "2025-08-01 10:00:00", it.VodTime)
	// only first source's episodes
	require.Len(t, it.Episodes, 2)
	require.Equal(t, int32(1), it.Episodes[0].Number)
	require.Equal(t, "http://a.m3u8", it.Episodes[0].URL)
	require.Equal(t, "hls", it.Episodes[0].Format)
	require.Equal(t, int32(2), it.Episodes[1].Number)
	require.Equal(t, "http://b.m3u8", it.Episodes[1].URL)
}

func TestParseDefaultDetailEmptySources(t *testing.T) {
	body := []byte(`{
		"code":0,"message":"success",
		"data":[{"id":101,"title":"无线路影片","category_id":5,"sources":[]}]
	}`)
	page, err := parseDefaultDetail(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Empty(t, page.Items[0].Episodes)
}

func TestParseDefaultDetailSkipsEmptyTitle(t *testing.T) {
	body := []byte(`{
		"code":0,"message":"success",
		"data":[
			{"id":101,"title":"","category_id":5,"sources":[]},
			{"id":102,"title":"有效影片","category_id":5,"sources":[]}
		]
	}`)
	page, err := parseDefaultDetail(body)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "有效影片", page.Items[0].Title)
}

func TestParseDefaultCategories(t *testing.T) {
	body := []byte(`{
		"code":0,"message":"success",
		"data":[
			{"id":1,"name":"电影","parent_id":0},
			{"id":6,"name":"动作片","parent_id":1}
		]
	}`)
	cats, err := parseDefaultCategories(body)
	require.NoError(t, err)
	require.Len(t, cats, 2)
	require.Equal(t, uint32(1), cats[0].ID)
	require.Equal(t, "电影", cats[0].Name)
	require.Equal(t, uint32(6), cats[1].ID)
	require.Equal(t, uint32(1), cats[1].ParentID)
}

func TestParseDefaultCategoriesSkipsInvalidID(t *testing.T) {
	body := []byte(`{"code":0,"message":"success","data":[{"id":0,"name":"无效","parent_id":0},{"id":5,"name":"有效","parent_id":0}]}`)
	cats, err := parseDefaultCategories(body)
	require.NoError(t, err)
	require.Len(t, cats, 1)
	require.Equal(t, uint32(5), cats[0].ID)
}

func TestDefaultEndpointURL(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		sub     string
		wantErr bool
		want    string
	}{
		{"normal", "https://example.com/api/open/v1", "videos", false, "https://example.com/api/open/v1/videos"},
		{"trailing slash", "https://example.com/api/open/v1/", "videos", false, "https://example.com/api/open/v1/videos"},
		{"sub with slash", "https://example.com/api/open/v1", "/videos/detail", false, "https://example.com/api/open/v1/videos/detail"},
		{"empty base", "", "videos", true, ""},
		{"invalid scheme", "ftp://x", "videos", true, ""},
		{"with query params", "https://example.com/api/open/v1?key=xxx", "videos", false, "https://example.com/api/open/v1/videos?key=xxx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := defaultEndpointURL(tt.base, tt.sub)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, u.String())
		})
	}
}

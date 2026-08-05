package collect

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/require"
)

func TestAppleCMSFetchListUsesListMode(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"class":[{"type_id":1,"type_pid":0,"type_name":"电影"}],"list":[]}`))
	}))
	t.Cleanup(srv.Close)

	f := NewFetcher(nil)
	c := &appleCMSCollector{fetcher: f, log: nil}
	source := &model.CollectSources{
		Type:       constant.CollectTypeAppleCMS,
		CollectURL: srv.URL + "/api.php",
		APIKey:     "k1",
	}
	_, err := c.FetchListPage(t.Context(), source, 1, "today")
	require.NoError(t, err)

	u, err := url.Parse(gotURL)
	require.NoError(t, err)
	// /provide/vod path is appended by endpointURL to the base address
	require.Equal(t, "/api.php/provide/vod", u.Path)
	q := u.Query()
	require.Equal(t, "list", q.Get("ac"))
	require.Equal(t, "1", q.Get("pg"))
	require.Equal(t, "50", q.Get("limit"))
	require.Equal(t, "24", q.Get("h"))
	require.Equal(t, "k1", q.Get("key"))
}

func TestAppleCMSFetchDetailUsesDetailMode(t *testing.T) {
	var gotPath string
	var gotAC string
	var gotIDs string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAC = r.URL.Query().Get("ac")
		gotIDs = r.URL.Query().Get("ids")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":1,"list":[]}`))
	}))
	t.Cleanup(srv.Close)

	f := NewFetcher(nil)
	c := &appleCMSCollector{fetcher: f, log: nil}
	source := &model.CollectSources{
		Type:       constant.CollectTypeAppleCMS,
		CollectURL: srv.URL + "/api.php",
	}
	_, err := c.FetchDetail(t.Context(), source, []uint32{101, 202, 303})
	require.NoError(t, err)
	require.Equal(t, "/api.php/provide/vod", gotPath)
	require.Equal(t, "detail", gotAC)
	require.Equal(t, "101,202,303", gotIDs)
}

func TestEndpointURL(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		wantErr bool
		want    string
	}{
		{"host only", "https://example.com", false, "https://example.com/provide/vod"},
		{"host only trailing slash", "https://example.com/", false, "https://example.com/provide/vod"},
		{"with api.php", "https://example.com/api.php", false, "https://example.com/api.php/provide/vod"},
		{"with api.php trailing slash", "https://example.com/api.php/", false, "https://example.com/api.php/provide/vod"},
		{"with query params", "https://example.com/api.php?key=xxx", false, "https://example.com/api.php/provide/vod?key=xxx"},
		{"empty", "", true, ""},
		{"whitespace only", "   ", true, ""},
		{"invalid scheme", "ftp://x", true, ""},
		{"no scheme", "example.com/api.php", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := endpointURL(tt.base)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, u.String())
		})
	}
}

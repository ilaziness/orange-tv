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
		CollectURL: srv.URL + "/api.php/provide/vod/?ac=detail&pg=2&h=24",
		APIKey:     "k1",
	}
	_, err := c.FetchListPage(t.Context(), source, 1, "today")
	require.NoError(t, err)

	u, err := url.Parse(gotURL)
	require.NoError(t, err)
	q := u.Query()
	require.Equal(t, "list", q.Get("ac"))
	require.Equal(t, "1", q.Get("pg"))
	require.Equal(t, "50", q.Get("limit"))
	require.Equal(t, "k1", q.Get("key"))
}

func TestAppleCMSFetchDetailUsesDetailMode(t *testing.T) {
	var gotAC string
	var gotIDs string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		CollectURL: srv.URL + "/api.php/provide/vod/",
	}
	_, err := c.FetchDetail(t.Context(), source, []int64{101, 202, 303})
	require.NoError(t, err)
	require.Equal(t, "detail", gotAC)
	require.Equal(t, "101,202,303", gotIDs)
}

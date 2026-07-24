package collect

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchListUsesListMode(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"class":[{"type_id":1,"type_pid":0,"type_name":"电影"}],"list":[]}`))
	}))
	t.Cleanup(srv.Close)

	f := NewFetcher(nil)
	body, err := f.FetchList(t.Context(), srv.URL+"/api.php/provide/vod/?ac=detail&pg=2&h=24", "k1", 1, "today")
	require.NoError(t, err)
	require.Contains(t, string(body), `"class"`)

	u, err := url.Parse(gotURL)
	require.NoError(t, err)
	q := u.Query()
	require.Equal(t, "list", q.Get("ac"))
	require.Equal(t, "1", q.Get("pg"))
	require.Equal(t, "50", q.Get("limit"))
	require.Equal(t, "k1", q.Get("key"))
}

func TestFetchDetailUsesDetailMode(t *testing.T) {
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
	_, err := f.FetchDetail(t.Context(), srv.URL+"/api.php/provide/vod/", "", []int64{101, 202, 303})
	require.NoError(t, err)
	require.Equal(t, "detail", gotAC)
	require.Equal(t, "101,202,303", gotIDs)
}

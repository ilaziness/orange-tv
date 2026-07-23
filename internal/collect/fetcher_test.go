package collect

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchAppleCMSCategoriesUsesListMode(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"class":[{"type_id":1,"type_pid":0,"type_name":"电影"}],"list":[]}`))
	}))
	t.Cleanup(srv.Close)

	f := NewFetcher()
	body, err := f.FetchAppleCMSCategories(t.Context(), srv.URL+"/api.php/provide/vod/?ac=detail&pg=2&h=24", "k1")
	require.NoError(t, err)
	require.Contains(t, string(body), `"class"`)

	u, err := url.Parse(gotURL)
	require.NoError(t, err)
	q := u.Query()
	require.Equal(t, "list", q.Get("ac"))
	require.Equal(t, "k1", q.Get("key"))
	require.Empty(t, q.Get("h"))
}

func TestFetchPageDefaultsAppleDetail(t *testing.T) {
	var gotAC string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAC = r.URL.Query().Get("ac")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":1,"list":[]}`))
	}))
	t.Cleanup(srv.Close)

	f := NewFetcher()
	_, err := f.FetchPage(t.Context(), srv.URL+"/api.php/provide/vod/", "", 1, true, "")
	require.NoError(t, err)
	require.Equal(t, "detail", gotAC)
}

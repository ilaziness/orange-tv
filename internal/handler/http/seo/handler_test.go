package seo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	seosvc "github.com/ilaziness/orange-tv/internal/service/seo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseSitemapVideoPage(t *testing.T) {
	t.Parallel()
	n, ok := parseSitemapVideoPage("videos-41.xml")
	require.True(t, ok)
	require.Equal(t, 41, n)
	_, ok = parseSitemapVideoPage("41.xml")
	require.False(t, ok)
	_, ok = parseSitemapVideoPage("videos-0.xml")
	require.False(t, ok)
	_, ok = parseSitemapVideoPage("videos-abc.xml")
	require.False(t, ok)
	_, ok = parseSitemapVideoPage("pages.xml")
	require.False(t, ok)
}

func TestSitemapVideos_routeParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(seosvc.StubService{}, zap.NewNop())
	r := gin.New()
	r.GET("/sitemaps/pages.xml", func(c *gin.Context) { c.String(http.StatusOK, "pages") })
	r.GET("/sitemaps/:name", h.SitemapVideos)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sitemaps/videos-41.xml", nil)
	r.ServeHTTP(w, req)
	// Stub always 404 for sitemap videos, but must reach service (not fail param parse).
	require.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/sitemaps/pages.xml", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "pages", w.Body.String())
}

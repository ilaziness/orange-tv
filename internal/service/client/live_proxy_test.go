package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteM3U8_AbsoluteAndRelativeURLs(t *testing.T) {
	body := `#EXTM3U
#EXT-X-VERSION:3
#EXTINF:-1 tvg-id="1",Channel 1
http://example.com/segment1.ts
#EXTINF:-1,Channel 2
relative/segment2.ts
http://another.host/playlist.m3u8
#EXT-X-ENDLIST
`
	rewritten := rewriteM3U8([]byte(body), "http://origin.example/live.m3u8", 42)
	lines := strings.Split(rewritten, "\n")

	require.Contains(t, rewritten, "#EXTM3U")
	assert.True(t, strings.HasPrefix(lines[3], "/api/client/v1/live/play/42?u="))
	assert.True(t, strings.HasPrefix(lines[5], "/api/client/v1/live/play/42?u="))
	assert.True(t, strings.HasPrefix(lines[6], "/api/client/v1/live/play/42?u="))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		assert.True(t, strings.HasPrefix(line, "/api/client/v1/live/play/42?u="), "unexpected line: %s", line)
	}
}

func TestRewriteM3U8_KeyAndMapURIs(t *testing.T) {
	body := `#EXTM3U
#EXT-X-KEY:METHOD=AES-128,URI="../key.bin"
#EXT-X-MAP:URI="init.mp4"
#EXTINF:-1,Channel
segment.ts
`
	rewritten := rewriteM3U8([]byte(body), "http://origin.example/path/playlist.m3u8", 7)

	assert.Contains(t, rewritten, "#EXT-X-KEY:METHOD=AES-128,URI=\"/api/client/v1/live/play/7?u=")
	assert.Contains(t, rewritten, "#EXT-X-MAP:URI=\"/api/client/v1/live/play/7?u=")
	assert.Contains(t, rewritten, "/api/client/v1/live/play/7?u=")
}

func TestSegmentContentType(t *testing.T) {
	cases := []struct {
		url        string
		upstreamCT string
		peek       []byte
		want       string
	}{
		{"http://x/a.ts", "application/octet-stream", nil, "video/mp2t"},
		{"http://x/a.aac", "application/octet-stream", nil, "audio/aac"},
		{"http://x/a.m4s", "application/octet-stream", nil, "video/mp4"},
		{"http://x/a.mp4", "application/octet-stream", nil, "video/mp4"},
		{"http://x/a.key", "application/octet-stream", nil, "application/octet-stream"},
		{"http://x/a.m3u8", "application/octet-stream", nil, "application/vnd.apple.mpegurl"},
		{"http://x/unknown", "video/mp2t", nil, "video/mp2t"},
		{"http://x/unknown", "application/octet-stream", nil, "video/mp2t"},
		{"http://x/unknown", "application/octet-stream", []byte{0x47, 0x00}, "video/mp2t"},
		{"http://x/unknown", "application/octet-stream", []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p'}, "video/mp4"},
		{"http://x/unknown", "application/octet-stream", []byte{0xFF, 0xF1, 0x00}, "audio/aac"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, segmentContentType(tc.url, tc.upstreamCT, tc.peek), "url=%s", tc.url)
	}
}

func TestEncodeDecodeSegmentURL(t *testing.T) {
	input := "http://example.com/path?foo=bar&baz=qux"
	encoded := encodeSegmentURL(input)
	decoded, err := decodeSegmentURL(encoded)
	require.NoError(t, err)
	assert.Equal(t, input, decoded)

	// URL-safe base64 must not contain '+' or '/'.
	assert.NotContains(t, encoded, "+")
	assert.NotContains(t, encoded, "/")
}

func TestDecodeSegmentURL_InvalidBase64(t *testing.T) {
	_, err := decodeSegmentURL("!!!not-base64!!!")
	assert.Error(t, err)
}

func TestLiveProxyService_isDomainAllowed_CacheConcurrency(t *testing.T) {
	called := 0
	var mu sync.Mutex
	svc := &mockLiveService{
		allowedDomains: map[string]struct{}{
			"allowed.example": {},
		},
		domainsHook: func() {
			mu.Lock()
			called++
			mu.Unlock()
		},
	}
	proxy := NewLiveProxyService(svc, nil)
	lps := proxy.(*liveProxyService)
	lps.domainsTTL = 100 * time.Millisecond

	// First call hits the service.
	assert.True(t, lps.isDomainAllowed(context.Background(), "http://allowed.example/segment.ts"))
	mu.Lock()
	assert.Equal(t, 1, called)
	mu.Unlock()

	// Subsequent calls use the in-memory cache.
	for i := 0; i < 10; i++ {
		assert.True(t, lps.isDomainAllowed(context.Background(), "http://allowed.example/segment.ts"))
	}
	mu.Lock()
	assert.Equal(t, 1, called)
	mu.Unlock()

	// Disallowed domain should not trigger a service call when cache is warm.
	assert.False(t, lps.isDomainAllowed(context.Background(), "http://evil.example/segment.ts"))
	mu.Lock()
	assert.Equal(t, 1, called)
	mu.Unlock()

	// Wait for cache expiry and verify a refresh happens.
	time.Sleep(150 * time.Millisecond)
	assert.True(t, lps.isDomainAllowed(context.Background(), "http://allowed.example/segment.ts"))
	mu.Lock()
	assert.Equal(t, 2, called)
	mu.Unlock()
}

func TestLiveProxyService_Proxy_Options(t *testing.T) {
	svc := &mockLiveService{}
	proxy := NewLiveProxyService(svc, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/live/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	err := proxy.Proxy(c, 1, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

type mockLiveService struct {
	allowedDomains map[string]struct{}
	domainsHook    func()
}

func (m *mockLiveService) List(ctx context.Context, req *clientdto.LiveListRequest) ([]clientdto.LiveChannelItem, int, error) {
	return nil, 0, nil
}

func (m *mockLiveService) GetStreamURL(ctx context.Context, id uint32) (string, error) {
	return "", nil
}

func (m *mockLiveService) AllowedStreamDomains(ctx context.Context) (map[string]struct{}, error) {
	if m.domainsHook != nil {
		m.domainsHook()
	}
	return m.allowedDomains, nil
}

func TestLiveProxyService_proxyURL_PlaylistWithoutExtension(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:-1\nsegment.ts\n"))
	}))
	defer upstream.Close()

	svc := &mockLiveService{
		allowedDomains: map[string]struct{}{
			mustParseHost(upstream.URL): {},
		},
	}
	proxy := NewLiveProxyService(svc, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/live/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	segURL := encodeSegmentURL(upstream.URL + "/playlist")
	err := proxy.Proxy(c, 1, segURL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.apple.mpegurl", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "/api/client/v1/live/play/1?u=")
}

func TestLiveProxyService_proxyURL_SegmentWithoutExtension(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("binarysegmentdata"))
	}))
	defer upstream.Close()

	svc := &mockLiveService{
		allowedDomains: map[string]struct{}{
			mustParseHost(upstream.URL): {},
		},
	}
	proxy := NewLiveProxyService(svc, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/live/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	segURL := encodeSegmentURL(upstream.URL + "/segment")
	err := proxy.Proxy(c, 1, segURL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "video/mp2t", w.Header().Get("Content-Type"))
	assert.Equal(t, "binarysegmentdata", w.Body.String())
}

func TestLiveProxyService_proxyURL_PlaylistDetectedByBodyPeek(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Misleading Content-Type and no .m3u8 extension; body peek must detect the playlist.
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:-1\nsegment.ts\n"))
	}))
	defer upstream.Close()

	svc := &mockLiveService{
		allowedDomains: map[string]struct{}{
			mustParseHost(upstream.URL): {},
		},
	}
	proxy := NewLiveProxyService(svc, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/live/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	segURL := encodeSegmentURL(upstream.URL + "/resource")
	err := proxy.Proxy(c, 1, segURL)
	require.NoError(t, err)
	assert.Equal(t, "application/vnd.apple.mpegurl", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "#EXTM3U")
}

func TestLiveProxyService_proxyURL_SegmentWithoutExtensionDetectedByBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No extension and octet-stream Content-Type; body sniff should identify MPEG-TS.
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{0x47, 0x00, 0x00, 0x00, 0x00})
	}))
	defer upstream.Close()

	svc := &mockLiveService{
		allowedDomains: map[string]struct{}{
			mustParseHost(upstream.URL): {},
		},
	}
	proxy := NewLiveProxyService(svc, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/live/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	segURL := encodeSegmentURL(upstream.URL + "/segment")
	err := proxy.Proxy(c, 1, segURL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "video/mp2t", w.Header().Get("Content-Type"))
}

func mustParseHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u.Host
}

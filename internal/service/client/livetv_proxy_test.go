package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestProxy(t *testing.T) *liveTVProxyService {
	t.Helper()
	proxy := NewLiveTVProxyService(&mockLiveTVService{}, nil)
	return proxy.(*liveTVProxyService)
}

// newTestProxyWithUpstream creates a proxy whose HTTP client connects to
// httptest servers (which run on 127.0.0.1) by bypassing the private-IP
// DialContext check that production uses for SSRF defense.
func newTestProxyWithUpstream(t *testing.T) *liveTVProxyService {
	t.Helper()
	proxy := NewLiveTVProxyService(&mockLiveTVService{}, nil)
	lps := proxy.(*liveTVProxyService)
	lps.httpc = &http.Client{
		Timeout: 10 * time.Second,
	}
	return lps
}

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
	lps := newTestProxy(t)
	rewritten := lps.rewriteM3U8([]byte(body), "http://origin.example/live.m3u8", 42)
	lines := strings.Split(rewritten, "\n")

	require.Contains(t, rewritten, "#EXTM3U")
	assert.True(t, strings.HasPrefix(lines[3], "/api/client/v1/livetv/play/42?u="))
	assert.True(t, strings.HasPrefix(lines[5], "/api/client/v1/livetv/play/42?u="))
	assert.True(t, strings.HasPrefix(lines[6], "/api/client/v1/livetv/play/42?u="))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		assert.True(t, strings.HasPrefix(line, "/api/client/v1/livetv/play/42?u="), "unexpected line: %s", line)
	}
}

func TestRewriteM3U8_KeyAndMapURIs(t *testing.T) {
	body := `#EXTM3U
#EXT-X-KEY:METHOD=AES-128,URI="../key.bin"
#EXT-X-MAP:URI="init.mp4"
#EXTINF:-1,Channel
segment.ts
`
	lps := newTestProxy(t)
	rewritten := lps.rewriteM3U8([]byte(body), "http://origin.example/path/playlist.m3u8", 7)

	assert.Contains(t, rewritten, "#EXT-X-KEY:METHOD=AES-128,URI=\"/api/client/v1/livetv/play/7?u=")
	assert.Contains(t, rewritten, "#EXT-X-MAP:URI=\"/api/client/v1/livetv/play/7?u=")
	assert.Contains(t, rewritten, "/api/client/v1/livetv/play/7?u=")
}

func TestRewriteM3U8_PreservesQueryParams(t *testing.T) {
	body := `#EXTM3U
#EXTINF:-1,Channel
segment.ts
`
	lps := newTestProxy(t)
	// Base URL has a token query param; the segment URL is relative and has no query.
	rewritten := lps.rewriteM3U8([]byte(body), "http://origin.example/live.m3u8?token=abc123", 1)
	lines := strings.Split(rewritten, "\n")

	var segLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "/api/client/v1/livetv/play/1?u=") {
			segLine = line
			break
		}
	}
	require.NotEmpty(t, segLine, "segment line not found")

	encoded := strings.TrimPrefix(segLine, "/api/client/v1/livetv/play/1?u=")
	decoded, err := lps.decodeSegmentURL(encoded)
	require.NoError(t, err)
	// The resolved segment URL should preserve the base query token.
	assert.Contains(t, decoded, "token=abc123")
	assert.Contains(t, decoded, "segment.ts")
}

func TestSegmentContentType(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		upstreamCT string
		peek       []byte
		want       string
	}{
		{"ts extension", "http://x/a.ts", "application/octet-stream", nil, "video/mp2t"},
		{"aac extension", "http://x/a.aac", "application/octet-stream", nil, "audio/aac"},
		{"m4s extension", "http://x/a.m4s", "application/octet-stream", nil, "video/mp4"},
		{"mp4 extension", "http://x/a.mp4", "application/octet-stream", nil, "video/mp4"},
		{"key extension", "http://x/a.key", "application/octet-stream", nil, "application/octet-stream"},
		{"m3u8 extension", "http://x/a.m3u8", "application/octet-stream", nil, "application/vnd.apple.mpegurl"},
		{"unknown with concrete CT", "http://x/unknown", "video/mp2t", nil, "video/mp2t"},
		{"unknown with octet-stream", "http://x/unknown", "application/octet-stream", nil, "video/mp2t"},
		{"body sniff mpegts", "http://x/unknown", "application/octet-stream", []byte{0x47, 0x00}, "video/mp2t"},
		{"body sniff fmp4", "http://x/unknown", "application/octet-stream", []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p'}, "video/mp4"},
		{"body sniff aac", "http://x/unknown", "application/octet-stream", []byte{0xFF, 0xF1, 0x00}, "audio/aac"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, segmentContentType(tc.url, tc.upstreamCT, tc.peek))
		})
	}
}

func TestEncodeDecodeSegmentURL(t *testing.T) {
	lps := newTestProxy(t)
	input := "http://example.com/path?foo=bar&baz=qux"
	encoded := lps.encodeSegmentURL(input)
	decoded, err := lps.decodeSegmentURL(encoded)
	require.NoError(t, err)
	assert.Equal(t, input, decoded)

	// URL-safe base64 (RawURLEncoding) must not contain '+', '/', or '='.
	assert.NotContains(t, encoded, "+")
	assert.NotContains(t, encoded, "/")
	assert.NotContains(t, encoded, "=")
}

func TestDecodeSegmentURL_InvalidBase64(t *testing.T) {
	lps := newTestProxy(t)
	_, err := lps.decodeSegmentURL("!!!not-base64!!!")
	assert.Error(t, err)
}

func TestDecodeSegmentURL_ForgedHMAC(t *testing.T) {
	lps := newTestProxy(t)
	// Encode with one proxy instance, tamper the signature, decode with another.
	encoded := lps.encodeSegmentURL("http://example.com/segment.ts")
	parts := strings.SplitN(encoded, ".", 2)
	require.Len(t, parts, 2)
	// Flip a character in the signature portion to forge it.
	forged := parts[0] + "." + "AAAAAAAAAAA"
	_, err := lps.decodeSegmentURL(forged)
	assert.Error(t, err)
}

func TestDecodeSegmentURL_MissingSignature(t *testing.T) {
	lps := newTestProxy(t)
	_, err := lps.decodeSegmentURL("aGVsbG8")
	assert.Error(t, err)
}

func TestIsPrivateURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want bool
	}{
		{"loopback IPv4", "http://127.0.0.1/stream.ts", true},
		{"private 10.x", "http://10.0.0.1/stream.ts", true},
		{"private 192.168.x", "http://192.168.1.1/stream.ts", true},
		{"private 172.16.x", "http://172.16.0.1/stream.ts", true},
		{"link-local", "http://169.254.1.1/stream.ts", true},
		{"localhost", "http://localhost/stream.ts", true},
		{"public domain", "http://example.com/stream.ts", false},
		{"public IPv4", "http://8.8.8.8/stream.ts", false},
		{"public IPv4 with port", "http://222.169.85.8:9901/stream.ts", false},
		{"invalid URL", "not-a-url", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isPrivateURL(tc.url))
		})
	}
}

func TestLooksLikeM3U8_BodyOverContentType(t *testing.T) {
	// Content-Type claims m3u8 but body is HTML error page — should NOT be treated as m3u8.
	htmlBody := []byte("<html><body>403 Forbidden</body></html>")
	assert.False(t, looksLikeM3U8(htmlBody))

	// Body starts with #EXTM3U — should be treated as m3u8 regardless of Content-Type.
	m3u8Body := []byte("#EXTM3U\n#EXTINF:-1\nsegment.ts\n")
	assert.True(t, looksLikeM3U8(m3u8Body))
}

type mockLiveTVService struct {
	streamURL string
	streamErr error
}

func (m *mockLiveTVService) List(ctx context.Context, req *clientdto.LiveTVListRequest) ([]clientdto.LiveTVChannelItem, int, error) {
	return nil, 0, nil
}

func (m *mockLiveTVService) GetStreamURL(ctx context.Context, id uint32) (string, error) {
	return m.streamURL, m.streamErr
}

func TestLiveProxyService_proxyURL_PlaylistWithoutExtension(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:-1\nsegment.ts\n"))
	}))
	defer upstream.Close()

	lps := newTestProxyWithUpstream(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livetv/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	err := lps.proxyURL(c, 1, upstream.URL+"/playlist")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.apple.mpegurl", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "/api/client/v1/livetv/play/1?u=")
}

func TestLiveProxyService_proxyURL_SegmentWithoutExtension(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("binarysegmentdata"))
	}))
	defer upstream.Close()

	lps := newTestProxyWithUpstream(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livetv/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	err := lps.proxyURL(c, 1, upstream.URL+"/segment")
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

	lps := newTestProxyWithUpstream(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livetv/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	err := lps.proxyURL(c, 1, upstream.URL+"/resource")
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

	lps := newTestProxyWithUpstream(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livetv/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	err := lps.proxyURL(c, 1, upstream.URL+"/segment")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "video/mp2t", w.Header().Get("Content-Type"))
}

func TestLiveProxyService_proxyURL_UpstreamError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer upstream.Close()

	lps := newTestProxyWithUpstream(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livetv/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	err := lps.proxyURL(c, 1, upstream.URL+"/segment")
	assert.Error(t, err)
}

func TestLiveProxyService_PrivateIPBlocked(t *testing.T) {
	svc := &mockLiveTVService{}
	proxy := NewLiveTVProxyService(svc, nil)
	lps := proxy.(*liveTVProxyService)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livetv/play/1", nil)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	segURL := lps.encodeSegmentURL("http://127.0.0.1/secret/stream.ts")
	err := proxy.Proxy(c, 1, segURL)
	assert.Error(t, err)
}

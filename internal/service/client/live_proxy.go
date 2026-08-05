package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"go.uber.org/zap"
)

// LiveProxyService proxies IPTV live streams.
type LiveProxyService interface {
	// Proxy handles live stream proxy for the given channel.
	// If segURL is empty, it proxies the master m3u8 playlist for channelID.
	// If segURL is non-empty, it proxies the decoded segment/sub-playlist URL.
	// The method writes the HTTP response directly into gin.Context and returns
	// an error only when the request cannot be fulfilled before response headers.
	Proxy(c *gin.Context, channelID uint32, segURL string) error
}

type liveProxyService struct {
	svc             LiveService
	log             *zap.Logger
	httpc           *http.Client
	domainsMutex    sync.RWMutex
	domainsTTL      time.Duration
	domainsLoadedAt time.Time
	cachedDomains   map[string]struct{}
}

// NewLiveProxyService creates a client LiveProxyService.
// httpc is configured without a global timeout so long-running stream
// transfers are not interrupted; connect/response-header timeouts are
// controlled by Transport.
func NewLiveProxyService(svc LiveService, log *zap.Logger) LiveProxyService {
	if log == nil {
		log = zap.NewNop()
	}
	return &liveProxyService{
		svc: svc,
		log: log,
		httpc: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).DialContext,
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},
		domainsTTL: 1 * time.Minute,
	}
}

// Proxy handles live stream proxy.
func (s *liveProxyService) Proxy(c *gin.Context, channelID uint32, segURL string) error {
	s.writeCORS(c)
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return nil
	}

	var realURL string
	if segURL != "" {
		decoded, err := decodeSegmentURL(segURL)
		if err != nil {
			s.log.Debug("[LIVE-PROXY] invalid segment url", zap.String("encoded", segURL), zap.Error(err))
			return errcode.WithMessage(errcode.ParamError, "无效的分片地址")
		}
		if !s.isDomainAllowed(c.Request.Context(), decoded) {
			s.log.Warn("[LIVE-PROXY] segment url not in allowed domains", zap.String("url", decoded))
			return errcode.WithMessage(errcode.ParamError, "分片地址不在允许范围内")
		}
		realURL = decoded
		s.log.Debug("[LIVE-PROXY] resource request", zap.String("url", realURL))
	} else {
		streamURL, err := s.svc.GetStreamURL(c.Request.Context(), channelID)
		if err != nil {
			s.log.Error("[LIVE-PROXY] get stream url failed", zap.Uint32("channel_id", channelID), zap.Error(err))
			return err
		}
		realURL = streamURL
		s.log.Debug("[LIVE-PROXY] master playlist request", zap.Uint32("channel_id", channelID), zap.String("url", realURL))
	}

	return s.proxyURL(c, channelID, realURL)
}

// proxyURL fetches a URL and either rewrites an m3u8 playlist or streams a media segment.
// Detection is based on Content-Type, URL extension, and a small body peek, so it works
// for segment URLs without a .ts extension and for playlist URLs without a .m3u8 extension.
func (s *liveProxyService) proxyURL(c *gin.Context, channelID uint32, realURL string) error {
	resp, err := s.fetchResp(c.Request.Context(), realURL)
	if err != nil {
		s.log.Error("[LIVE-PROXY] fetch failed", zap.String("url", realURL), zap.Error(err))
		return errcode.WithMessage(errcode.LiveSyncFailed, "拉取直播流失败")
	}
	defer func() { _ = resp.Body.Close() }()

	contentType := resp.Header.Get("Content-Type")
	finalURL := resp.Request.URL.String()

	peek := make([]byte, 512)
	n, readErr := resp.Body.Read(peek)
	if n == 0 {
		if readErr == nil {
			readErr = io.EOF
		}
		s.log.Error("[LIVE-PROXY] empty upstream response", zap.String("url", realURL), zap.Error(readErr))
		return errcode.WithMessage(errcode.LiveSyncFailed, "读取直播流失败")
	}
	peek = peek[:n]

	// Treat as m3u8 when Content-Type, extension, or body strongly suggests it.
	if isM3U8ContentType(contentType) || isM3U8URL(realURL) || looksLikeM3U8(peek) {
		rest, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			s.log.Error("[LIVE-PROXY] read playlist failed", zap.String("url", realURL), zap.Error(readErr))
			return errcode.WithMessage(errcode.LiveSyncFailed, "读取播放列表失败")
		}
		body := append(peek, rest...)

		if isM3U8(body, contentType) {
			rewritten := rewriteM3U8(body, finalURL, channelID)
			c.Header("Cache-Control", "no-cache")
			c.Data(http.StatusOK, "application/vnd.apple.mpegurl", []byte(rewritten))
			return nil
		}

		// Content-Type/extension/body claimed m3u8 but it is not; pass through as segment.
		ct := segmentContentType(finalURL, contentType, body)
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, ct, body)
		return nil
	}

	// Stream as a media segment without loading large files into memory.
	ct := segmentContentType(finalURL, contentType, peek)
	s.log.Debug("[LIVE-PROXY] streaming segment",
		zap.String("url", realURL),
		zap.String("final_url", finalURL),
		zap.String("content_type", contentType),
		zap.String("final_ct", ct),
	)
	c.Header("Cache-Control", "no-cache")
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		c.Header("Content-Length", cl)
	}
	c.DataFromReader(http.StatusOK, resp.ContentLength, ct, io.MultiReader(bytes.NewReader(peek), resp.Body), nil)
	return nil
}

// fetchResp sends a GET request using the client request context.
func (s *liveProxyService) fetchResp(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "orange-tv-live-proxy/1.0")
	return s.httpc.Do(req)
}

// writeCORS writes cross-origin response headers.
func (s *liveProxyService) writeCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "*")
}

// isDomainAllowed checks whether the target URL host is in the online channel domain whitelist.
func (s *liveProxyService) isDomainAllowed(ctx context.Context, rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	domains, err := s.getDomains(ctx)
	if err != nil {
		return false
	}
	_, ok := domains[u.Host]
	return ok
}

// getDomains fetches the domain whitelist with a short in-memory cache.
func (s *liveProxyService) getDomains(ctx context.Context) (map[string]struct{}, error) {
	s.domainsMutex.RLock()
	if s.cachedDomains != nil && time.Since(s.domainsLoadedAt) < s.domainsTTL {
		defer s.domainsMutex.RUnlock()
		return s.cachedDomains, nil
	}
	s.domainsMutex.RUnlock()

	s.domainsMutex.Lock()
	defer s.domainsMutex.Unlock()
	// Double-check after acquiring write lock.
	if s.cachedDomains != nil && time.Since(s.domainsLoadedAt) < s.domainsTTL {
		return s.cachedDomains, nil
	}
	domains, err := s.svc.AllowedStreamDomains(ctx)
	if err != nil {
		return nil, err
	}
	s.cachedDomains = domains
	s.domainsLoadedAt = time.Now()
	return domains, nil
}

// rewriteM3U8 rewrites m3u8 content URLs so they go through the proxy endpoint.
func rewriteM3U8(body []byte, baseURL string, channelID uint32) string {
	text := string(body)
	base, _ := url.Parse(baseURL)
	lines := strings.Split(text, "\n")
	prefix := fmt.Sprintf("/api/client/v1/live/play/%d?u=", channelID)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			rewritten := rewriteTagURI(line, base, prefix)
			if rewritten != line {
				lines[i] = rewritten
			}
			continue
		}
		abs := resolveURL(base, trimmed)
		if abs == "" {
			continue
		}
		lines[i] = prefix + encodeSegmentURL(abs)
	}
	return strings.Join(lines, "\n")
}

// rewriteTagURI rewrites URI="..." portions of m3u8 tags.
func rewriteTagURI(line string, base *url.URL, prefix string) string {
	idx := strings.Index(line, "URI=\"")
	if idx < 0 {
		return line
	}
	end := strings.Index(line[idx+5:], "\"")
	if end < 0 {
		return line
	}
	uri := line[idx+5 : idx+5+end]
	abs := resolveURL(base, uri)
	if abs == "" {
		return line
	}
	return line[:idx+5] + prefix + encodeSegmentURL(abs) + line[idx+5+end:]
}

// resolveURL resolves a relative URL against a base URL.
func resolveURL(base *url.URL, ref string) string {
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	if base == nil {
		return u.String()
	}
	return base.ResolveReference(u).String()
}

// encodeSegmentURL encodes a real URL with URL-safe base64.
func encodeSegmentURL(rawURL string) string {
	return base64.URLEncoding.EncodeToString([]byte(rawURL))
}

// decodeSegmentURL decodes a base64 segment URL.
func decodeSegmentURL(encoded string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// isM3U8ContentType checks whether the Content-Type header indicates an m3u8 playlist.
func isM3U8ContentType(contentType string) bool {
	lower := strings.ToLower(contentType)
	return strings.Contains(lower, "mpegurl") || strings.Contains(lower, "m3u")
}

// looksLikeM3U8 checks whether the first bytes of the body look like an m3u8 playlist.
func looksLikeM3U8(data []byte) bool {
	return strings.HasPrefix(strings.TrimSpace(string(data)), "#EXTM3U")
}

// isM3U8 determines whether the response body is an m3u8 playlist.
func isM3U8(body []byte, contentType string) bool {
	if isM3U8ContentType(contentType) {
		return true
	}
	return looksLikeM3U8(body)
}

// isM3U8URL guesses whether a URL points to an m3u8 playlist by extension.
func isM3U8URL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	ext := strings.ToLower(path.Ext(u.Path))
	return ext == ".m3u8" || ext == ".m3u"
}

// segmentContentType infers the correct Content-Type for a segment from its URL extension
// or, for extensionless/octet-stream responses, by sniffing the first bytes of the body.
// Many IPTV sources return application/octet-stream for ts segments, which players cannot
// recognize, so the type is corrected here.
func segmentContentType(rawURL, upstreamCT string, peek []byte) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return upstreamCT
	}
	ext := strings.ToLower(path.Ext(u.Path))
	switch ext {
	case ".ts":
		return "video/mp2t"
	case ".aac":
		return "audio/aac"
	case ".m4s", ".mp4":
		return "video/mp4"
	case ".flv":
		return "video/x-flv"
	case ".key":
		// Encryption keys stay binary.
		return "application/octet-stream"
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	}

	// No recognized extension: trust upstream if it gives a concrete type.
	lower := strings.ToLower(upstreamCT)
	if upstreamCT != "" && !strings.Contains(lower, "octet-stream") {
		return upstreamCT
	}

	// Fall back to body sniffing for extensionless or octet-stream responses.
	if len(peek) > 0 {
		// MPEG-TS sync byte.
		if peek[0] == 0x47 {
			return "video/mp2t"
		}
		// fMP4 'ftyp' box signature (4-byte size followed by 'ftyp').
		if len(peek) >= 8 && string(peek[4:8]) == "ftyp" {
			return "video/mp4"
		}
		// AAC ADTS syncword: first 12 bits are 0xFFF.
		if peek[0] == 0xFF && len(peek) > 1 && (peek[1]&0xF0) == 0xF0 {
			return "audio/aac"
		}
	}
	return "video/mp2t"
}

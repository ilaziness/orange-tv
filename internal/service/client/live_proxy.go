package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"go.uber.org/zap"
)

// browserUserAgent is a common desktop browser User-Agent string used when
// fetching upstream IPTV sources. Many sources reject non-browser UAs with 403.
const browserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

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
	svc        LiveService
	log        *zap.Logger
	httpc      *http.Client
	hmacSecret []byte
}

// NewLiveProxyService creates a client LiveProxyService.
// httpc is configured without a global timeout so long-running stream
// transfers are not interrupted; connect/response-header timeouts are
// controlled by Transport. TLS certificate verification is skipped to support
// IPTV sources on non-standard ports with self-signed certificates.
func NewLiveProxyService(svc LiveService, log *zap.Logger) LiveProxyService {
	if log == nil {
		log = zap.NewNop()
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		// crypto/rand should never fail on supported platforms; fall back to a
		// fixed but unpredictable-enough value to keep the service running.
		secret = []byte("orange-tv-live-proxy-fallback-key")
	}
	return &liveProxyService{
		svc:        svc,
		log:        log,
		hmacSecret: secret,
		httpc: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					host, _, err := net.SplitHostPort(addr)
					if err != nil {
						return nil, err
					}
					// Resolve hostname and block connections to private IPs
					// (SSRF defense for hostname-based URLs; IP literals are
					// already blocked in isPrivateURL before reaching here).
					ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
					if err != nil {
						return nil, err
					}
					for _, ipAddr := range ips {
						if isPrivateIP(ipAddr.IP) {
							return nil, fmt.Errorf("connection to private IP %s blocked", ipAddr.IP)
						}
					}
					return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, addr)
				},
				ResponseHeaderTimeout: 10 * time.Second,
				// #nosec G402 -- IPTV sources often use self-signed certs or
				// non-standard ports (e.g. :4430); skipping verification is a
				// deliberate usability trade-off for this read-only proxy.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// Proxy handles live stream proxy.
func (s *liveProxyService) Proxy(c *gin.Context, channelID uint32, segURL string) error {
	var realURL string
	if segURL != "" {
		decoded, err := s.decodeSegmentURL(segURL)
		if err != nil {
			s.log.Debug("[LIVE-PROXY] invalid segment url", zap.String("encoded", segURL), zap.Error(err))
			return errcode.WithMessage(errcode.ParamError, "无效的分片地址")
		}
		if isPrivateURL(decoded) {
			s.log.Warn("[LIVE-PROXY] segment url points to private address", zap.String("url", decoded))
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.log.Error("[LIVE-PROXY] upstream non-2xx",
			zap.String("url", realURL),
			zap.Int("status", resp.StatusCode),
			zap.String("content_type", resp.Header.Get("Content-Type")),
		)
		return errcode.WithMessage(errcode.LiveSyncFailed, fmt.Sprintf("直播源返回错误状态码: %d", resp.StatusCode))
	}

	contentType := resp.Header.Get("Content-Type")
	finalURL := resp.Request.URL.String()

	peek := make([]byte, 512)
	n, readErr := io.ReadFull(resp.Body, peek)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		s.log.Error("[LIVE-PROXY] read upstream body failed", zap.String("url", realURL), zap.Error(readErr))
		return errcode.WithMessage(errcode.LiveSyncFailed, "读取直播流失败")
	}
	if n == 0 {
		s.log.Error("[LIVE-PROXY] empty upstream response", zap.String("url", realURL))
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

		if looksLikeM3U8(body) {
			rewritten := s.rewriteM3U8(body, finalURL, channelID)
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
	// resp.ContentLength is the full body length; since we prepend peek back via
	// MultiReader, the total equals resp.ContentLength. When upstream uses chunked
	// encoding, resp.ContentLength is -1 and DataFromReader uses chunked transfer.
	ct := segmentContentType(finalURL, contentType, peek)
	s.log.Debug("[LIVE-PROXY] streaming segment",
		zap.String("url", realURL),
		zap.String("final_url", finalURL),
		zap.String("content_type", contentType),
		zap.String("final_ct", ct),
	)
	c.Header("Cache-Control", "no-cache")
	c.DataFromReader(http.StatusOK, resp.ContentLength, ct, io.MultiReader(bytes.NewReader(peek), resp.Body), nil)
	return nil
}

// fetchResp sends a GET request using the client request context.
// It sets a browser User-Agent and a Referer matching the upstream origin
// to bypass common hotlink protection on IPTV sources.
func (s *liveProxyService) fetchResp(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "*/*")
	if u, perr := url.Parse(rawURL); perr == nil && u.Host != "" {
		// Set Referer to the upstream origin — browsers send this on same-origin
		// requests and many IPTV CDNs require it for hotlink protection.
		// Origin is intentionally omitted: browsers don't send it on simple GET
		// requests, and including it may cause some servers to reject the request.
		req.Header.Set("Referer", u.Scheme+"://"+u.Host+"/")
	}
	return s.httpc.Do(req)
}

// rewriteM3U8 rewrites m3u8 content URLs so they go through the proxy endpoint.
func (s *liveProxyService) rewriteM3U8(body []byte, baseURL string, channelID uint32) string {
	text := string(body)
	base, _ := url.Parse(baseURL)
	baseQuery := ""
	if base != nil && base.RawQuery != "" {
		baseQuery = base.RawQuery
	}
	lines := strings.Split(text, "\n")
	prefix := fmt.Sprintf("/api/client/v1/live/play/%d?u=", channelID)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			rewritten := s.rewriteTagURI(line, base, prefix, baseQuery)
			if rewritten != line {
				lines[i] = rewritten
			}
			continue
		}
		abs := resolveURL(base, trimmed, baseQuery)
		if abs == "" {
			continue
		}
		lines[i] = prefix + s.encodeSegmentURL(abs)
	}
	return strings.Join(lines, "\n")
}

// rewriteTagURI rewrites URI="..." portions of m3u8 tags.
func (s *liveProxyService) rewriteTagURI(line string, base *url.URL, prefix, baseQuery string) string {
	idx := strings.Index(line, "URI=\"")
	if idx < 0 {
		return line
	}
	end := strings.Index(line[idx+5:], "\"")
	if end < 0 {
		return line
	}
	uri := line[idx+5 : idx+5+end]
	abs := resolveURL(base, uri, baseQuery)
	if abs == "" {
		return line
	}
	return line[:idx+5] + prefix + s.encodeSegmentURL(abs) + line[idx+5+end:]
}

// resolveURL resolves a relative URL against a base URL.
// If the resolved URL has no query string but the base URL does, the base
// query is appended so that CDN tokens are preserved across redirects.
func resolveURL(base *url.URL, ref, baseQuery string) string {
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
	resolved := base.ResolveReference(u)
	// Preserve base query params (e.g. CDN tokens) only when the segment URL
	// itself has no query string, to avoid overwriting segment-specific params.
	if resolved.RawQuery == "" && baseQuery != "" {
		resolved.RawQuery = baseQuery
	}
	return resolved.String()
}

// encodeSegmentURL encodes a real URL with URL-safe base64 (no padding) and
// appends an HMAC-SHA256 signature so that only URLs generated by this proxy
// instance are accepted on decode.
func (s *liveProxyService) encodeSegmentURL(rawURL string) string {
	mac := hmac.New(sha256.New, s.hmacSecret)
	_, _ = mac.Write([]byte(rawURL))
	sig := mac.Sum(nil)[:8]
	return base64.RawURLEncoding.EncodeToString([]byte(rawURL)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// decodeSegmentURL decodes a base64 segment URL and verifies its HMAC signature.
func (s *liveProxyService) decodeSegmentURL(encoded string) (string, error) {
	parts := strings.SplitN(encoded, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid segment url format")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, s.hmacSecret)
	_, _ = mac.Write(raw)
	expected := mac.Sum(nil)[:8]
	if !hmac.Equal(sig, expected) {
		return "", fmt.Errorf("invalid segment url signature")
	}
	return string(raw), nil
}

// isPrivateURL reports whether the URL points to a private/loopback IP literal.
// Used as SSRF defense so that signed segment URLs cannot target internal services.
// Only IP literals are checked here (fast path, no DNS lookup). Hostname-based
// private IP detection is handled in DialContext at connection time to avoid
// blocking DNS resolution on every segment request.
func isPrivateURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return true
	}
	hostname := u.Hostname()
	if hostname == "localhost" {
		return true
	}
	ip := net.ParseIP(hostname)
	if ip == nil {
		// Hostname: defer to DialContext check at connection time.
		return false
	}
	return isPrivateIP(ip)
}

// isPrivateIP reports whether an IP address is private, loopback, link-local, or unspecified.
func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
}

// isM3U8ContentType checks whether the Content-Type header indicates an m3u8 playlist.
func isM3U8ContentType(contentType string) bool {
	lower := strings.ToLower(contentType)
	return strings.Contains(lower, "mpegurl") || strings.Contains(lower, "m3u")
}

// looksLikeM3U8 checks whether the first bytes of the body look like an m3u8 playlist.
// Body content takes priority over Content-Type: if the body does not start
// with #EXTM3U, it is not treated as a playlist even when the Content-Type
// claims otherwise (some misconfigured servers return the wrong type).
func looksLikeM3U8(data []byte) bool {
	return strings.HasPrefix(strings.TrimSpace(string(data)), "#EXTM3U")
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

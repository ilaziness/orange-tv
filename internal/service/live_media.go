package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
)

// ErrLiveResumeFailed means the cloud provider did not clear a persistent forbid.
var ErrLiveResumeFailed = errors.New("live resume failed")

// LiveReForbidError is a failed start whose follow-up forbid also failed.
// Unwrap returns the original start error.
type LiveReForbidError struct {
	Mark error
	Kick error
}

func (e *LiveReForbidError) Error() string {
	return fmt.Sprintf("%v; re-forbid: %v", e.Mark, e.Kick)
}

func (e *LiveReForbidError) Unwrap() error { return e.Mark }

// ResumeForStart clears a persistent forbid, then runs mark (the living transaction).
// If mark fails because the room is already living, the stream is left resumed.
// Any other mark failure puts the forbid back so an idle room cannot keep publishing.
func ResumeForStart(ctx context.Context, client *http.Client, cfg LiveStreamConfig, streamName string, mark func() error) error {
	if err := ResumeLiveStream(ctx, client, cfg, streamName); err != nil {
		return fmt.Errorf("%w: %v", ErrLiveResumeFailed, err)
	}
	err := mark()
	if err == nil {
		return nil
	}
	if code, ok := errcode.As(err); ok && code.Code == errcode.LiveRoomLiving.Code {
		return err
	}
	if kickErr := KickLiveStream(ctx, client, cfg, streamName); kickErr != nil {
		return &LiveReForbidError{Mark: err, Kick: kickErr}
	}
	return err
}

// NewLiveHTTPClient is the client used for SRS kick and cloud forbid/resume.
func NewLiveHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// ResumeLiveStream clears a persistent cloud forbid before a room is marked living.
// Self-hosted SRS has nothing to resume. Failure must abort start.
func ResumeLiveStream(ctx context.Context, client *http.Client, cfg LiveStreamConfig, streamName string) error {
	streamName = strings.TrimSpace(streamName)
	switch cfg.ProviderName() {
	case constant.LiveProviderSelf:
		return nil
	case constant.LiveProviderAliyun:
		return aliyunCall(ctx, client, cfg.Aliyun, "ResumeLiveStream", map[string]string{
			"StreamName": streamName,
		})
	case constant.LiveProviderQiniu:
		return qiniuStreamDisabled(ctx, client, cfg.Qiniu, streamName, 0)
	default:
		return fmt.Errorf("resume: unknown provider %s", cfg.ProviderName())
	}
}

// KickLiveStream disconnects the publisher after the room is already idle.
// Self-hosted kick is skipped when auth is off. Cloud forbid is persistent.
func KickLiveStream(ctx context.Context, client *http.Client, cfg LiveStreamConfig, streamName string) error {
	streamName = strings.TrimSpace(streamName)
	switch cfg.ProviderName() {
	case constant.LiveProviderSelf:
		if !cfg.AuthEnabled {
			return nil
		}
		return kickSRS(ctx, client, cfg, streamName)
	case constant.LiveProviderAliyun:
		return aliyunCall(ctx, client, cfg.Aliyun, "ForbidLiveStream", map[string]string{
			"StreamName": streamName,
			"Oneshot":    "no",
		})
	case constant.LiveProviderQiniu:
		return qiniuStreamDisabled(ctx, client, cfg.Qiniu, streamName, -1)
	default:
		return fmt.Errorf("kick: unknown provider %s", cfg.ProviderName())
	}
}

type srsClient struct {
	ID      string `json:"id"`
	Stream  string `json:"stream"`
	URL     string `json:"url"`
	Type    string `json:"type"`
	Publish bool   `json:"publish"`
}

type srsClientsResponse struct {
	Code    int         `json:"code"`
	Clients []srsClient `json:"clients"`
}

const (
	srsClientPageSize = 100
	srsClientMaxPages = 50
)

func kickSRS(ctx context.Context, client *http.Client, cfg LiveStreamConfig, streamName string) error {
	if client == nil {
		return fmt.Errorf("srs kick: http client is nil")
	}
	app, err := pushApp(cfg.PushServer)
	if err != nil || app == "" {
		return fmt.Errorf("srs kick: push app: %w", err)
	}
	base := strings.TrimRight(cfg.SRSAPIBase, "/")
	var kickErr error
	for page := 0; page < srsClientMaxPages; page++ {
		start := page * srsClientPageSize
		listed, err := srsListClients(ctx, client, base, start, srsClientPageSize)
		if err != nil {
			return err
		}
		for _, c := range listed.Clients {
			if !srsPublishing(c, app, streamName) {
				continue
			}
			if err := srsDisconnect(ctx, client, base, c.ID); err != nil && kickErr == nil {
				kickErr = err
			}
		}
		if len(listed.Clients) < srsClientPageSize {
			break
		}
	}
	return kickErr
}

func srsListClients(ctx context.Context, client *http.Client, base string, start, count int) (srsClientsResponse, error) {
	raw := fmt.Sprintf("%s/api/v1/clients?start=%d&count=%d", base, start, count)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return srsClientsResponse{}, fmt.Errorf("srs list clients: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return srsClientsResponse{}, fmt.Errorf("srs list clients: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return srsClientsResponse{}, fmt.Errorf("srs list clients: status %d body %s", resp.StatusCode, truncate(body))
	}
	var listed srsClientsResponse
	if err := json.Unmarshal(body, &listed); err != nil {
		return srsClientsResponse{}, fmt.Errorf("srs list clients: %w", err)
	}
	if listed.Code != 0 {
		return srsClientsResponse{}, fmt.Errorf("srs list clients: code %d", listed.Code)
	}
	return listed, nil
}

func srsPublishing(c srsClient, app, streamName string) bool {
	if c.ID == "" {
		return false
	}
	publishing := c.Publish || strings.Contains(strings.ToLower(c.Type), "publish")
	if !publishing {
		return false
	}
	if c.Stream == streamName && srsApp(c.URL) == app {
		return true
	}
	path := strings.Trim(c.URL, "/")
	return path == app+"/"+streamName
}

func srsApp(rawURL string) string {
	u, err := url.Parse(rawURL)
	path := rawURL
	if err == nil && u.Path != "" {
		path = u.Path
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func srsDisconnect(ctx context.Context, client *http.Client, base, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, base+"/api/v1/clients/"+url.PathEscape(id), nil)
	if err != nil {
		return fmt.Errorf("srs disconnect: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("srs disconnect %s: %w", id, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("srs disconnect %s: status %d body %s", id, resp.StatusCode, string(body))
	}
	return nil
}

func truncate(b []byte) string {
	if len(b) > 512 {
		return string(b[:512])
	}
	return string(b)
}

package collect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Fetcher downloads collect pages via HTTP. Format-specific URL building and
// query params live in the corresponding Collector implementation; Fetcher only
// provides the shared HTTP client and request execution.
type Fetcher struct {
	client *http.Client
	log    *zap.Logger
}

// NewFetcher creates a Fetcher with default timeouts.
func NewFetcher(log *zap.Logger) *Fetcher {
	if log == nil {
		log = zap.NewNop()
	}
	return &Fetcher{
		client: &http.Client{Timeout: 30 * time.Second},
		log:    log,
	}
}

// doGet executes a GET request and returns the response body.
// apiKey, when non-empty, is sent both as the X-API-Key header and is expected
// to already be embedded in rawURL by the caller if needed.
func (f *Fetcher) doGet(ctx context.Context, rawURL, apiKey string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		f.log.Error("fetcher: create request failed", zap.String("url", rawURL), zap.Error(err))
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "orange-tv-collect/1.0")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		f.log.Error("fetcher: http request failed", zap.String("url", rawURL), zap.Error(err))
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		f.log.Error("fetcher: read body failed", zap.String("url", rawURL), zap.Error(err))
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		f.log.Error("fetcher: unexpected status", zap.String("url", rawURL), zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return body, nil
}

package collect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Fetcher downloads collect pages.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a Fetcher with default timeouts.
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchPage GET the collect URL with page params and optional API key.
func (f *Fetcher) FetchPage(ctx context.Context, baseURL, apiKey string, page int, isApple bool) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse collect url: %w", err)
	}
	q := u.Query()
	if page < 1 {
		page = 1
	}
	if isApple {
		if q.Get("ac") == "" {
			q.Set("ac", "detail")
		}
		q.Set("pg", strconv.Itoa(page))
	} else {
		q.Set("page", strconv.Itoa(page))
	}
	if apiKey != "" {
		if q.Get("key") == "" && q.Get("api_key") == "" {
			q.Set("key", apiKey)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "orange-tv-collect/1.0")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return body, nil
}

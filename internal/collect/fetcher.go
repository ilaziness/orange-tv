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
// dataRange filters by time for Apple CMS (today/last1d/last3d/last1w/last1m/all).
// Apple CMS content collection uses ac=detail (full vod fields); category class is usually absent.
func (f *Fetcher) FetchPage(ctx context.Context, baseURL, apiKey string, page int, isApple bool, dataRange string) ([]byte, error) {
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
		if h := dataRangeToHours(dataRange); h > 0 {
			q.Set("h", strconv.Itoa(h))
		}
	} else {
		q.Set("page", strconv.Itoa(page))
	}
	if apiKey != "" {
		if q.Get("key") == "" && q.Get("api_key") == "" {
			q.Set("key", apiKey)
		}
	}
	u.RawQuery = q.Encode()
	return f.doGet(ctx, u.String(), apiKey)
}

// FetchAppleCMSCategories GET Apple CMS list API (ac=list) which includes the class field.
// Many providers omit class when ac=detail; binding remote categories must use list mode.
func (f *Fetcher) FetchAppleCMSCategories(ctx context.Context, baseURL, apiKey string) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse collect url: %w", err)
	}
	q := u.Query()
	q.Set("ac", "list")
	// drop detail-only filters that some sources reject on list endpoints
	q.Del("h")
	q.Del("ids")
	q.Del("t")
	if apiKey != "" {
		if q.Get("key") == "" && q.Get("api_key") == "" {
			q.Set("key", apiKey)
		}
	}
	u.RawQuery = q.Encode()
	return f.doGet(ctx, u.String(), apiKey)
}

func (f *Fetcher) doGet(ctx context.Context, rawURL, apiKey string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
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

// dataRangeToHours converts a data range string to hours for Apple CMS h parameter.
// Returns 0 for "all" or empty (no filter).
func dataRangeToHours(dataRange string) int {
	switch strings.TrimSpace(dataRange) {
	case "today":
		return 24
	case "last1d":
		return 24
	case "last3d":
		return 72
	case "last1w":
		return 168
	case "last1m":
		return 720
	default:
		return 0
	}
}

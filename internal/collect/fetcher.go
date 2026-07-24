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

// FetchList GET the collect URL with ac=list mode to retrieve vod_id list and class info.
// dataRange filters by time via the h parameter (today/last1d/last3d/last1w/last1m/all).
// limit=50 is set to get more items per page (default is 20).
func (f *Fetcher) FetchList(ctx context.Context, baseURL, apiKey string, page int, dataRange string) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse collect url: %w", err)
	}
	q := u.Query()
	if page < 1 {
		page = 1
	}
	q.Set("ac", "list")
	q.Set("pg", strconv.Itoa(page))
	q.Set("limit", "50")
	if h := dataRangeToHours(dataRange); h > 0 {
		q.Set("h", strconv.Itoa(h))
	}
	if apiKey != "" {
		if q.Get("key") == "" && q.Get("api_key") == "" {
			q.Set("key", apiKey)
		}
	}
	u.RawQuery = q.Encode()
	return f.doGet(ctx, u.String(), apiKey)
}

// FetchDetail GET the collect URL with ac=detail mode to retrieve full vod info for given IDs.
// The caller should batch IDs (max 25 per call) to avoid URL length limits.
func (f *Fetcher) FetchDetail(ctx context.Context, baseURL, apiKey string, ids []int64) ([]byte, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse collect url: %w", err)
	}
	q := u.Query()
	q.Set("ac", "detail")
	// remove list-only params that detail endpoints don't need
	q.Del("pg")
	q.Del("h")
	q.Del("limit")
	if len(ids) > 0 {
		parts := make([]string, len(ids))
		for i, id := range ids {
			parts[i] = strconv.FormatInt(id, 10)
		}
		q.Set("ids", strings.Join(parts, ","))
	}
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

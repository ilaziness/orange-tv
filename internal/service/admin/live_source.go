package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const liveSourceURL = "https://live.zbds.top/tv/iptv4.txt"

// LiveSourceEntry represents a parsed entry from a live source.
type LiveSourceEntry struct {
	Category  string
	Name      string
	StreamURL string
	SortOrder uint32
}

// defaultLiveSourceFetcher fetches and parses the default iptv4.txt format.
// Future implementations can follow the same pattern for different URLs or formats.
type defaultLiveSourceFetcher struct {
	url string
}

func (f *defaultLiveSourceFetcher) Fetch(ctx context.Context) ([]LiveSourceEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch live source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch live source: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read live source body: %w", err)
	}

	return ParseLiveSource(string(body)), nil
}

// ParseLiveSource parses the iptv4.txt format into LiveSourceEntry slice.
// Format: categories are marked by "XXX,#genre#", channels are "name,url".
// The "更新时间" category and its entries are ignored.
// Same-name channels within a category get suffixes: first uses original name,
// second gets "(2)", third gets "(3)", etc.
func ParseLiveSource(raw string) []LiveSourceEntry {
	var entries []LiveSourceEntry
	category := ""
	sortOrder := uint32(0)
	nameCountInCategory := make(map[string]int)

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for category marker: "XXX,#genre#"
		if strings.HasSuffix(line, ",#genre#") {
			category = strings.TrimSuffix(line, ",#genre#")
			nameCountInCategory = make(map[string]int)
			continue
		}

		// Skip entries before any category is set
		if category == "" {
			continue
		}

		// Ignore "更新时间" category
		if category == "更新时间" {
			continue
		}

		// Parse "name,url"
		commaIdx := strings.Index(line, ",")
		if commaIdx <= 0 || commaIdx >= len(line)-1 {
			continue
		}

		name := strings.TrimSpace(line[:commaIdx])
		streamURL := strings.TrimSpace(line[commaIdx+1:])
		if name == "" || streamURL == "" {
			continue
		}

		// Add suffix for duplicate names within the same category
		nameCountInCategory[name]++
		if count := nameCountInCategory[name]; count > 1 {
			name = fmt.Sprintf("%s(%d)", name, count)
		}

		sortOrder++
		entries = append(entries, LiveSourceEntry{
			Category:  category,
			Name:      name,
			StreamURL: streamURL,
			SortOrder: sortOrder,
		})
	}

	return entries
}

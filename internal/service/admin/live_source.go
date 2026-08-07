package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sherif-fanous/m3u"
)

// LiveSourceEntry represents a parsed entry from a live source.
type LiveSourceEntry struct {
	Category  string
	Name      string
	StreamURL string
	Logo      string
	SortOrder uint32
}

// LiveSourceParser parses raw live source content into entries.
type LiveSourceParser interface {
	Parse(raw string) []LiveSourceEntry
}

// fetchLiveSource fetches raw content from the given live source URL.
func fetchLiveSource(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch live source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch live source: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read live source body: %w", err)
	}

	return string(body), nil
}

// selectParser selects a parser based on the URL file extension.
func selectParser(sourceURL string) (LiveSourceParser, error) {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid source URL: %w", err)
	}
	path := strings.ToLower(parsed.Path)
	switch {
	case strings.HasSuffix(path, ".m3u"):
		return &m3uParser{}, nil
	case strings.HasSuffix(path, ".txt"):
		return &txtParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported live source format, only .txt and .m3u are supported")
	}
}

// --- txt parser ---

type txtParser struct{}

func (p *txtParser) Parse(raw string) []LiveSourceEntry {
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

		// Filter out "支持作者" entries
		if strings.Contains(name, "支持作者") {
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

// --- m3u parser (using github.com/sherif-fanous/m3u) ---

// m3uFilteredGroupKeywords lists group-title keywords to filter out (promotion/non-content).
var m3uFilteredGroupKeywords = []string{
	"TG频道",
	"CloudFlare",
	"更新时间",
}

type m3uParser struct{}

func (p *m3uParser) Parse(raw string) []LiveSourceEntry {
	playlist, err := m3u.Unmarshal([]byte(raw))
	if err != nil {
		return nil
	}

	var entries []LiveSourceEntry
	sortOrder := uint32(0)
	nameCountInCategory := make(map[string]int)

	for _, track := range playlist.Tracks {
		name := strings.TrimSpace(track.Name)
		if name == "" {
			continue
		}

		// Filter out "支持作者" entries
		if strings.Contains(name, "支持作者") {
			continue
		}

		// Extract group-title as category
		category := ""
		if track.GroupTitle != nil {
			category = strings.TrimSpace(*track.GroupTitle)
		}
		if category == "" {
			continue
		}

		// Filter by group-title keywords
		if isFilteredGroupTitle(category) {
			continue
		}

		// Extract stream URL
		streamURL := ""
		if track.URL != nil {
			streamURL = track.URL.String()
		}
		if streamURL == "" {
			continue
		}

		// Extract logo
		logo := ""
		if track.TVGLogo != nil {
			logo = track.TVGLogo.String()
		}

		// Add suffix for duplicate names within the same category
		// Use composite key to ensure dedup is per-category, not global
		dedupeKey := category + "\x00" + name
		nameCountInCategory[dedupeKey]++
		if count := nameCountInCategory[dedupeKey]; count > 1 {
			name = fmt.Sprintf("%s(%d)", name, count)
		}

		sortOrder++
		entries = append(entries, LiveSourceEntry{
			Category:  category,
			Name:      name,
			StreamURL: streamURL,
			Logo:      logo,
			SortOrder: sortOrder,
		})
	}

	return entries
}

// isFilteredGroupTitle checks if the group-title contains any filtered keyword.
func isFilteredGroupTitle(groupTitle string) bool {
	for _, kw := range m3uFilteredGroupKeywords {
		if strings.Contains(groupTitle, kw) {
			return true
		}
	}
	return false
}

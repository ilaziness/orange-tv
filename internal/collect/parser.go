package collect

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseDefaultJSON parses the system default collect JSON format.
// Expected shape:
//
//	{
//	  "code": 0,
//	  "page": 1,
//	  "page_count": 1,
//	  "total": 1,
//	  "list": [{ "title": "...", "play_urls": [{"episode":1,"title":"","url":"","format":"hls"}] }]
//	}
func ParseDefaultJSON(body []byte) (*Page, error) {
	var raw struct {
		Code      int             `json:"code"`
		Page      int             `json:"page"`
		PageCount int             `json:"page_count"`
		Total     int             `json:"total"`
		List      json.RawMessage `json:"list"`
		Data      struct {
			List      json.RawMessage `json:"list"`
			Page      int             `json:"page"`
			PageCount int             `json:"page_count"`
			Total     int             `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("default json unmarshal: %w", err)
	}
	listRaw := raw.List
	page, pageCount, total := raw.Page, raw.PageCount, raw.Total
	if len(listRaw) == 0 && len(raw.Data.List) > 0 {
		listRaw = raw.Data.List
		page, pageCount, total = raw.Data.Page, raw.Data.PageCount, raw.Data.Total
	}
	if page < 1 {
		page = 1
	}
	var rows []defaultItem
	if len(listRaw) > 0 {
		if err := json.Unmarshal(listRaw, &rows); err != nil {
			return nil, fmt.Errorf("default list unmarshal: %w", err)
		}
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		it := Item{
			ExternalID:  strings.TrimSpace(row.ID),
			Title:       strings.TrimSpace(row.Title),
			Subtitle:    strings.TrimSpace(row.Subtitle),
			Description: strings.TrimSpace(row.Description),
			Cover:       strings.TrimSpace(row.Cover),
			Poster:      strings.TrimSpace(row.Poster),
			Year:        row.Year,
			Region:      strings.TrimSpace(row.Region),
			Language:    strings.TrimSpace(row.Language),
			Duration:    row.Duration,
			Rating:      row.Rating,
			ReleaseDate: strings.TrimSpace(row.ReleaseDate),
			CategoryKey: firstNonEmpty(row.Category, strconv.FormatInt(row.CategoryID, 10)),
			Directors:   splitNames(row.Directors),
			Actors:      splitNames(row.Actors),
			Tags:        splitNames(row.Tags),
		}
		if it.Title == "" {
			continue
		}
		if it.Poster == "" {
			it.Poster = it.Cover
		}
		for _, ep := range row.PlayURLs {
			url := strings.TrimSpace(ep.URL)
			if url == "" {
				continue
			}
			num := ep.Episode
			if num <= 0 {
				num = 1
			}
			format := strings.ToLower(strings.TrimSpace(ep.Format))
			if format == "" {
				format = guessFormat(url)
			}
			it.Episodes = append(it.Episodes, Episode{
				Number:  num,
				Title:   strings.TrimSpace(ep.Title),
				URL:     url,
				Quality: strings.TrimSpace(ep.Quality),
				Format:  format,
			})
		}
		items = append(items, it)
	}
	if pageCount < 1 {
		pageCount = 1
	}
	if total < len(items) {
		total = len(items)
	}
	return &Page{Page: page, PageCount: pageCount, Total: total, Items: items}, nil
}

type defaultItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Subtitle    string           `json:"subtitle"`
	Description string           `json:"description"`
	Cover       string           `json:"cover"`
	Poster      string           `json:"poster"`
	Year        int32            `json:"year"`
	Region      string           `json:"region"`
	Language    string           `json:"language"`
	Duration    int32            `json:"duration"`
	Rating      float64          `json:"rating"`
	ReleaseDate string           `json:"release_date"`
	Category    string           `json:"category"`
	CategoryID  int64            `json:"category_id"`
	Directors   string           `json:"directors"`
	Actors      string           `json:"actors"`
	Tags        string           `json:"tags"`
	PlayURLs    []defaultEpisode `json:"play_urls"`
}

type defaultEpisode struct {
	Episode int32  `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Format  string `json:"format"`
}

// ParseAppleCMS parses 苹果CMS list API JSON.
func ParseAppleCMS(body []byte) (*Page, error) {
	var raw struct {
		Code      any             `json:"code"`
		Page      any             `json:"page"`
		PageCount any             `json:"pagecount"`
		Total     any             `json:"total"`
		List      json.RawMessage `json:"list"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("apple cms unmarshal: %w", err)
	}
	page := anyToInt(raw.Page)
	pageCount := anyToInt(raw.PageCount)
	total := anyToInt(raw.Total)
	if page < 1 {
		page = 1
	}
	if pageCount < 1 {
		pageCount = 1
	}
	var rows []appleItem
	if len(raw.List) > 0 && string(raw.List) != "null" {
		if err := json.Unmarshal(raw.List, &rows); err != nil {
			return nil, fmt.Errorf("apple cms list: %w", err)
		}
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		title := strings.TrimSpace(row.VodName)
		if title == "" {
			continue
		}
		it := Item{
			ExternalID:  strings.TrimSpace(row.VodID),
			Title:       title,
			Subtitle:    strings.TrimSpace(row.VodSub),
			Description: firstNonEmpty(strings.TrimSpace(row.Blurb), strings.TrimSpace(row.Content)),
			Cover:       strings.TrimSpace(row.Pic),
			Poster:      strings.TrimSpace(row.Pic),
			Year:        int32(anyToInt(row.VodYear)),
			Region:      strings.TrimSpace(row.VodArea),
			Language:    strings.TrimSpace(row.VodLang),
			Duration:    int32(anyToInt(row.VodDuration)),
			Rating:      anyToFloat(row.DoubanScore),
			ReleaseDate: strings.TrimSpace(row.Pubdate),
			CategoryKey: firstNonEmpty(strings.TrimSpace(row.TypeID), strings.TrimSpace(row.TypeName), strings.TrimSpace(row.Class)),
			Directors:   splitNames(row.Director),
			Actors:      splitNames(row.Actor),
			Tags:        splitNames(firstNonEmpty(row.VodTag, row.Class)),
			Episodes:    parseApplePlayURLs(row.VodPlayURL),
		}
		items = append(items, it)
	}
	if total < len(items) {
		total = len(items)
	}
	return &Page{Page: page, PageCount: pageCount, Total: total, Items: items}, nil
}

type appleItem struct {
	VodID       string `json:"vod_id"`
	TypeID      string `json:"type_id"`
	TypeName    string `json:"type_name"`
	Class       string `json:"class"`
	VodName     string `json:"vod_name"`
	VodSub      string `json:"vod_sub"`
	VodTag      string `json:"vod_tag"`
	Pic         string `json:"pic"`
	Actor       string `json:"actor"`
	Director    string `json:"director"`
	Blurb       string `json:"blurb"`
	Content     string `json:"vod_content"`
	Pubdate     string `json:"pubdate"`
	VodYear     any    `json:"vod_year"`
	VodArea     string `json:"vod_area"`
	VodLang     string `json:"vod_lang"`
	VodDuration any    `json:"vod_duration"`
	DoubanScore any    `json:"douban_score"`
	VodPlayFrom string `json:"vod_play_from"`
	VodPlayURL  string `json:"vod_play_url"`
}

func parseApplePlayURLs(raw string) []Episode {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// multi-source groups are separated by $$$; take all groups
	groups := strings.Split(raw, "$$$")
	var eps []Episode
	seen := map[int32]bool{}
	for _, g := range groups {
		parts := strings.Split(g, "#")
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			title, url := part, part
			if idx := strings.Index(part, "$"); idx >= 0 {
				title = strings.TrimSpace(part[:idx])
				url = strings.TrimSpace(part[idx+1:])
			}
			if url == "" {
				continue
			}
			num := int32(i + 1)
			// try extract episode number from title
			if n := extractEpisodeNumber(title); n > 0 {
				num = n
			}
			if seen[num] {
				// keep first occurrence per episode number
				continue
			}
			seen[num] = true
			eps = append(eps, Episode{
				Number: num,
				Title:  title,
				URL:    url,
				Format: guessFormat(url),
			})
		}
	}
	return eps
}

func extractEpisodeNumber(s string) int32 {
	s = strings.TrimSpace(s)
	// 第12集 / 12
	var digits strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else if digits.Len() > 0 {
			break
		}
	}
	if digits.Len() == 0 {
		return 0
	}
	n, err := strconv.Atoi(digits.String())
	if err != nil || n <= 0 {
		return 0
	}
	return int32(n)
}

func splitNames(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "，", ",")
	s = strings.ReplaceAll(s, "、", ",")
	s = strings.ReplaceAll(s, "/", ",")
	s = strings.ReplaceAll(s, "|", ",")
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func guessFormat(url string) string {
	u := strings.ToLower(url)
	switch {
	case strings.Contains(u, ".m3u8"):
		return "hls"
	case strings.Contains(u, ".mpd"):
		return "dash"
	case strings.Contains(u, ".flv"):
		return "flv"
	case strings.Contains(u, ".mp4"):
		return "mp4"
	default:
		return "hls"
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" && v != "0" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func anyToInt(v any) int {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(t))
		return i
	default:
		s := fmt.Sprint(t)
		i, _ := strconv.Atoi(strings.TrimSpace(s))
		return i
	}
}

func anyToFloat(v any) float64 {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		f, _ := strconv.ParseFloat(fmt.Sprint(t), 64)
		return f
	}
}

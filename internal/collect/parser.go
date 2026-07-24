package collect

import (
	"encoding/json"
	"fmt"
	"regexp"
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
			ExternalID:         strings.TrimSpace(row.ID),
			Title:              strings.TrimSpace(row.Title),
			Subtitle:           strings.TrimSpace(row.Subtitle),
			Description:        strings.TrimSpace(row.Description),
			Cover:              strings.TrimSpace(row.Cover),
			Poster:             strings.TrimSpace(row.Poster),
			Year:               row.Year,
			Region:             strings.TrimSpace(row.Region),
			Language:           strings.TrimSpace(row.Language),
			Duration:           row.Duration,
			Rating:             row.Rating,
			ReleaseDate:        strings.TrimSpace(row.ReleaseDate),
			ExternalCategoryID: row.CategoryID,
			Directors:          splitNames(row.Directors),
			Actors:             splitNames(row.Actors),
			Tags:               splitNames(row.Tags),
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

// ParseAppleCMSList parses an Apple CMS list API response (ac=list).
// It only extracts vod_id list, vod_time list, and class info.
// Detailed fields are obtained separately via the detail API.
func ParseAppleCMSList(body []byte) (*ListPage, error) {
	var raw struct {
		Code      any             `json:"code"`
		Page      any             `json:"page"`
		PageCount any             `json:"pagecount"`
		Total     any             `json:"total"`
		List      json.RawMessage `json:"list"`
		Class     json.RawMessage `json:"class"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("apple cms list unmarshal: %w", err)
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

	var rows []appleListItem
	if len(raw.List) > 0 && string(raw.List) != "null" {
		if err := json.Unmarshal(raw.List, &rows); err != nil {
			return nil, fmt.Errorf("apple cms list rows: %w", err)
		}
	}
	vodIDs := make([]int64, 0, len(rows))
	vodTimes := make([]string, 0, len(rows))
	for _, row := range rows {
		vodIDs = append(vodIDs, int64(anyToInt(row.VodID)))
		vodTimes = append(vodTimes, strings.TrimSpace(row.VodTime))
	}

	var classRows []appleClass
	if len(raw.Class) > 0 && string(raw.Class) != "null" {
		if err := json.Unmarshal(raw.Class, &classRows); err != nil {
			return nil, fmt.Errorf("apple cms class: %w", err)
		}
	}
	classes := make([]AppleCMSClass, 0, len(classRows))
	for _, c := range classRows {
		classes = append(classes, AppleCMSClass{
			TypeID:   int64(anyToInt(c.TypeID)),
			TypeName: strings.TrimSpace(c.TypeName),
			TypePID:  int64(anyToInt(c.TypePID)),
		})
	}

	return &ListPage{Page: page, PageCount: pageCount, Total: total, VodIDs: vodIDs, VodTimes: vodTimes, Classes: classes}, nil
}

// ParseAppleCMSDetail parses an Apple CMS detail API response (ac=detail).
// It extracts full vod fields for each item, including directors, actors, tags, and episodes.
func ParseAppleCMSDetail(body []byte) (*Page, error) {
	var raw struct {
		Code      any             `json:"code"`
		Page      any             `json:"page"`
		PageCount any             `json:"pagecount"`
		Total     any             `json:"total"`
		List      json.RawMessage `json:"list"`
		Class     json.RawMessage `json:"class"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("apple cms detail unmarshal: %w", err)
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
			return nil, fmt.Errorf("apple cms detail list: %w", err)
		}
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		title := strings.TrimSpace(row.VodName)
		if title == "" {
			continue
		}
		it := Item{
			ExternalID:         anyToString(row.VodID),
			Title:              title,
			Subtitle:           strings.TrimSpace(row.VodSub),
			Description:        stripHTMLTags(row.Content),
			Cover:              strings.TrimSpace(row.Pic),
			Poster:             strings.TrimSpace(row.Pic),
			Year:               int32(anyToInt(row.VodYear)),
			Region:             strings.TrimSpace(row.VodArea),
			Language:           strings.TrimSpace(row.VodLang),
			Duration:           int32(anyToInt(row.VodDuration)),
			Rating:             anyToFloat(row.DoubanScore),
			ReleaseDate:        strings.TrimSpace(row.Pubdate),
			ExternalCategoryID: int64(anyToInt(row.TypeID)),
			Directors:          splitNames(row.Director),
			Actors:             splitNames(row.Actor),
			Tags:               splitNames(row.VodTag, row.Class),
			Episodes:           parseApplePlayURLs(row.VodPlayURL),
			VodTime:            strings.TrimSpace(row.VodTime),
			Remarks:            strings.TrimSpace(row.VodRemarks),
		}
		items = append(items, it)
	}
	if total < len(items) {
		total = len(items)
	}

	var classRows []appleClass
	if len(raw.Class) > 0 && string(raw.Class) != "null" {
		if err := json.Unmarshal(raw.Class, &classRows); err != nil {
			return nil, fmt.Errorf("apple cms class: %w", err)
		}
	}
	classes := make([]AppleCMSClass, 0, len(classRows))
	for _, c := range classRows {
		classes = append(classes, AppleCMSClass{
			TypeID:   int64(anyToInt(c.TypeID)),
			TypeName: strings.TrimSpace(c.TypeName),
			TypePID:  int64(anyToInt(c.TypePID)),
		})
	}

	return &Page{Page: page, PageCount: pageCount, Total: total, Items: items, Classes: classes}, nil
}

type appleClass struct {
	TypeID   any    `json:"type_id"`
	TypeName string `json:"type_name"`
	TypePID  any    `json:"type_pid"`
}

// appleListItem is the minimal struct for parsing list API responses.
// Only vod_id and vod_time are needed from the list endpoint.
type appleListItem struct {
	VodID   any    `json:"vod_id"`
	VodTime string `json:"vod_time"`
}

type appleItem struct {
	VodID       any    `json:"vod_id"`
	TypeID      any    `json:"type_id"`
	TypeName    string `json:"type_name"`
	Class       string `json:"vod_class"`
	VodName     string `json:"vod_name"`
	VodSub      string `json:"vod_sub"`
	VodTag      string `json:"vod_tag"`
	Pic         string `json:"vod_pic"`
	Actor       string `json:"vod_actor"`
	Director    string `json:"vod_director"`
	Content     string `json:"vod_content"`
	Pubdate     string `json:"vod_pubdate"`
	VodYear     any    `json:"vod_year"`
	VodArea     string `json:"vod_area"`
	VodLang     string `json:"vod_lang"`
	VodDuration any    `json:"vod_duration"`
	DoubanScore any    `json:"douban_score"`
	VodPlayFrom string `json:"vod_play_from"`
	VodPlayURL  string `json:"vod_play_url"`
	VodTime     string `json:"vod_time"`
	VodRemarks  string `json:"vod_remarks"`
}

func parseApplePlayURLs(raw string) []Episode {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// multi-source groups are separated by $$$; only parse groups containing .m3u8 URLs
	groups := strings.Split(raw, "$$$")
	var eps []Episode
	seen := map[int32]bool{}
	for _, g := range groups {
		if !strings.Contains(g, ".m3u8") {
			continue
		}
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

func splitNames(parts ...string) []string {
	out := make([]string, 0)
	seen := map[string]bool{}
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		s = strings.ReplaceAll(s, "，", ",")
		s = strings.ReplaceAll(s, "、", ",")
		s = strings.ReplaceAll(s, "/", ",")
		s = strings.ReplaceAll(s, "|", ",")
		// split by comma first, then by whitespace for each part
		commaParts := strings.Split(s, ",")
		for _, cp := range commaParts {
			// split each comma-separated part by whitespace to handle space-delimited names
			spaceParts := strings.Fields(cp)
			for _, p := range spaceParts {
				p = strings.TrimSpace(p)
				if p == "" || seen[p] {
					continue
				}
				seen[p] = true
				out = append(out, p)
			}
		}
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

var htmlTagRegexp = regexp.MustCompile(`<[^>]*>`)

// stripHTMLTags removes HTML tags from s and trims whitespace.
func stripHTMLTags(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = htmlTagRegexp.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
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
	case float32:
		return int(t)
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint:
		return int(t)
	case uint32:
		return int(t)
	case uint64:
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

func anyToString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
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

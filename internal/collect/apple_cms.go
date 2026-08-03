package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/utils"
	"go.uber.org/zap"
)

// appleCMSCollector implements Collector for Apple CMS format sources.
type appleCMSCollector struct {
	fetcher *Fetcher
	log     *zap.Logger
}

// FetchListPage GET the collect URL with ac=list mode to retrieve vod_id list and class info.
// dataRange filters by time via the h parameter (today/last1d/last3d/last1w/last1m/all).
// limit=50 is set to get more items per page (default is 20).
func (c *appleCMSCollector) FetchListPage(ctx context.Context, source *model.CollectSources, page int, dataRange string) (*ListPage, error) {
	body, err := c.fetchAppleList(ctx, source.CollectURL, source.APIKey, page, dataRange)
	if err != nil {
		return nil, err
	}
	return parseAppleCMSList(body)
}

// FetchDetail GET the collect URL with ac=detail mode to retrieve full vod info for given IDs.
// The caller should batch IDs (max 25 per call) to avoid URL length limits.
func (c *appleCMSCollector) FetchDetail(ctx context.Context, source *model.CollectSources, ids []uint32) (*Page, error) {
	body, err := c.fetchAppleDetail(ctx, source.CollectURL, source.APIKey, ids)
	if err != nil {
		return nil, err
	}
	return parseAppleCMSDetail(body)
}

// FetchCategories fetches Apple CMS class info via ac=list first page.
func (c *appleCMSCollector) FetchCategories(ctx context.Context, source *model.CollectSources) ([]RemoteCategory, error) {
	body, err := c.fetchAppleList(ctx, source.CollectURL, source.APIKey, 1, "all")
	if err != nil {
		return nil, err
	}
	lp, err := parseAppleCMSList(body)
	if err != nil {
		return nil, err
	}
	return lp.Categories, nil
}

// fetchAppleList GET the collect URL with ac=list mode to retrieve vod_id list and class info.
// dataRange filters by time via the h parameter (today/last1d/last3d/last1w/last1m/all).
// limit=50 is set to get more items per page (default is 20).
func (c *appleCMSCollector) fetchAppleList(ctx context.Context, baseURL, apiKey string, page int, dataRange string) ([]byte, error) {
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
	return c.fetcher.doGet(ctx, u.String(), apiKey)
}

// fetchAppleDetail GET the collect URL with ac=detail mode to retrieve full vod info for given IDs.
// The caller should batch IDs (max 25 per call) to avoid URL length limits.
func (c *appleCMSCollector) fetchAppleDetail(ctx context.Context, baseURL, apiKey string, ids []uint32) ([]byte, error) {
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
			parts[i] = strconv.FormatUint(uint64(id), 10)
		}
		q.Set("ids", strings.Join(parts, ","))
	}
	if apiKey != "" {
		if q.Get("key") == "" && q.Get("api_key") == "" {
			q.Set("key", apiKey)
		}
	}
	u.RawQuery = q.Encode()
	return c.fetcher.doGet(ctx, u.String(), apiKey)
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

// parseAppleCMSList parses an Apple CMS list API response (ac=list).
// It only extracts vod_id list, vod_time list, and class info.
// Detailed fields are obtained separately via the detail API.
func parseAppleCMSList(body []byte) (*ListPage, error) {
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
	page := utils.AnyToInt(raw.Page)
	pageCount := utils.AnyToInt(raw.PageCount)
	total := utils.AnyToInt(raw.Total)
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
	ids := make([]uint32, 0, len(rows))
	times := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, uint32(utils.AnyToInt(row.VodID)))
		times = append(times, strings.TrimSpace(row.VodTime))
	}

	categories := parseAppleCMSClasses(raw.Class)

	return &ListPage{Page: page, PageCount: pageCount, Total: total, IDs: ids, Times: times, Categories: categories}, nil
}

// parseAppleCMSDetail parses an Apple CMS detail API response (ac=detail).
// It extracts full vod fields for each item, including directors, actors, tags, and episodes.
func parseAppleCMSDetail(body []byte) (*Page, error) {
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
	page := utils.AnyToInt(raw.Page)
	pageCount := utils.AnyToInt(raw.PageCount)
	total := utils.AnyToInt(raw.Total)
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
			ExternalID:         utils.AnyToString(row.VodID),
			Title:              title,
			Subtitle:           strings.TrimSpace(row.VodSub),
			Description:        utils.StripHTMLTags(row.Content),
			Cover:              strings.TrimSpace(row.Pic),
			Poster:             strings.TrimSpace(row.Pic),
			Year:               int32(utils.AnyToInt(row.VodYear)),
			Region:             strings.TrimSpace(row.VodArea),
			Language:           strings.TrimSpace(row.VodLang),
			Duration:           int32(utils.AnyToInt(row.VodDuration)),
			ReleaseDate:        strings.TrimSpace(row.Pubdate),
			ExternalCategoryID: uint32(utils.AnyToInt(row.TypeID)),
			Directors:          utils.SplitNames(row.Director),
			Actors:             utils.SplitNames(row.Actor),
			Tags:               utils.SplitNames(row.VodTag, row.Class),
			Episodes:           parseApplePlayURLs(row.VodPlayURL),
			VodTime:            strings.TrimSpace(row.VodTime),
			Remarks:            strings.TrimSpace(row.VodRemarks),
		}
		items = append(items, it)
	}
	if total < len(items) {
		total = len(items)
	}

	categories := parseAppleCMSClasses(raw.Class)

	return &Page{Page: page, PageCount: pageCount, Total: total, Items: items, Classes: categories}, nil
}

// parseAppleCMSClasses parses the Apple CMS class field into RemoteCategory.
func parseAppleCMSClasses(raw json.RawMessage) []RemoteCategory {
	var classRows []appleClass
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &classRows); err != nil {
			return nil
		}
	}
	out := make([]RemoteCategory, 0, len(classRows))
	for _, c := range classRows {
		out = append(out, RemoteCategory{
			ID:       uint32(utils.AnyToInt(c.TypeID)),
			Name:     strings.TrimSpace(c.TypeName),
			ParentID: uint32(utils.AnyToInt(c.TypePID)),
		})
	}
	return out
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

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

// defaultCollector implements Collector for the system default (Open API) format.
// CollectURL is the Open API base path (e.g. https://host/api/open/v1);
// the collector appends /videos, /videos/detail, /categories.
type defaultCollector struct {
	fetcher *Fetcher
	log     *zap.Logger
}

// FetchListPage GET {base}/videos?page=N&page_size=50&data_range=xxx
func (c *defaultCollector) FetchListPage(ctx context.Context, source *model.CollectSources, page int, dataRange string) (*ListPage, error) {
	if page < 1 {
		page = 1
	}
	u, err := defaultEndpointURL(source.CollectURL, "videos")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", "50")
	if dr := strings.TrimSpace(dataRange); dr != "" && dr != "all" {
		q.Set("data_range", dr)
	}
	u.RawQuery = q.Encode()
	body, err := c.fetcher.doGet(ctx, u.String(), source.APIKey)
	if err != nil {
		return nil, err
	}
	return parseDefaultList(body)
}

// FetchDetail GET {base}/videos/detail?id=1&id=2... (max 50 ids per call)
func (c *defaultCollector) FetchDetail(ctx context.Context, source *model.CollectSources, ids []uint32) (*Page, error) {
	if len(ids) == 0 {
		return &Page{}, nil
	}
	if len(ids) > 50 {
		ids = ids[:50]
	}
	u, err := defaultEndpointURL(source.CollectURL, "videos/detail")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for _, id := range ids {
		q.Add("id", strconv.FormatUint(uint64(id), 10))
	}
	u.RawQuery = q.Encode()
	body, err := c.fetcher.doGet(ctx, u.String(), source.APIKey)
	if err != nil {
		return nil, err
	}
	return parseDefaultDetail(body)
}

// FetchCategories GET {base}/categories
func (c *defaultCollector) FetchCategories(ctx context.Context, source *model.CollectSources) ([]RemoteCategory, error) {
	u, err := defaultEndpointURL(source.CollectURL, "categories")
	if err != nil {
		return nil, err
	}
	body, err := c.fetcher.doGet(ctx, u.String(), source.APIKey)
	if err != nil {
		return nil, err
	}
	return parseDefaultCategories(body)
}

// defaultEndpointURL joins the base path with a sub-path (videos / videos/detail / categories).
// Existing query params on baseURL are preserved; the sub-path is appended to the path only.
func defaultEndpointURL(baseURL, sub string) (*url.URL, error) {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return nil, fmt.Errorf("采集地址为空")
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse collect url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("采集地址协议无效: %s", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("采集地址主机为空")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.TrimLeft(sub, "/")
	return u, nil
}

// parseDefaultList parses the Open API video list response.
// Expected shape (wrapped in standard response envelope):
//
//	{"code":0,"message":"success","data":{"list":[{"id":1,"title":"...","category_id":2,"created_at":"..."}],"total":100,"page":1,"page_size":50,"total_pages":2}}
func parseDefaultList(body []byte) (*ListPage, error) {
	var raw struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("default list unmarshal: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("default list error: code=%d message=%s", raw.Code, raw.Message)
	}
	var page struct {
		List       []defaultListItem `json:"list"`
		Total      int               `json:"total"`
		Page       int               `json:"page"`
		PageSize   int               `json:"page_size"`
		TotalPages int               `json:"total_pages"`
	}
	if len(raw.Data) == 0 {
		return &ListPage{Page: 1, PageCount: 1}, nil
	}
	if err := json.Unmarshal(raw.Data, &page); err != nil {
		return nil, fmt.Errorf("default list data unmarshal: %w", err)
	}
	if page.Page < 1 {
		page.Page = 1
	}
	if page.TotalPages < 1 {
		page.TotalPages = 1
	}
	ids := make([]uint32, 0, len(page.List))
	times := make([]string, 0, len(page.List))
	for _, it := range page.List {
		if it.ID <= 0 {
			continue
		}
		ids = append(ids, it.ID)
		times = append(times, strings.TrimSpace(it.CreatedAt))
	}
	return &ListPage{
		Page:      page.Page,
		PageCount: page.TotalPages,
		Total:     page.Total,
		IDs:       ids,
		Times:     times,
	}, nil
}

// parseDefaultDetail parses the Open API video detail response.
// Only the first source's episodes are taken (multi-line sources collapsed into one).
func parseDefaultDetail(body []byte) (*Page, error) {
	var raw struct {
		Code    int                `json:"code"`
		Message string             `json:"message"`
		Data    []defaultDetailRow `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("default detail unmarshal: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("default detail error: code=%d message=%s", raw.Code, raw.Message)
	}
	items := make([]Item, 0, len(raw.Data))
	for _, row := range raw.Data {
		it := mapDefaultDetailRow(row)
		if it.Title == "" {
			continue
		}
		items = append(items, it)
	}
	total := len(items)
	if total == 0 {
		return &Page{Page: 1, PageCount: 1, Items: items}, nil
	}
	return &Page{Page: 1, PageCount: 1, Total: total, Items: items}, nil
}

// parseDefaultCategories parses the Open API categories response.
func parseDefaultCategories(body []byte) ([]RemoteCategory, error) {
	var raw struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    []defaultCategoryRow `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("default categories unmarshal: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("default categories error: code=%d message=%s", raw.Code, raw.Message)
	}
	out := make([]RemoteCategory, 0, len(raw.Data))
	for _, c := range raw.Data {
		if c.ID <= 0 {
			continue
		}
		out = append(out, RemoteCategory{
			ID:       c.ID,
			Name:     strings.TrimSpace(c.Name),
			ParentID: c.ParentID,
		})
	}
	return out, nil
}

// mapDefaultDetailRow converts an Open API detail item to the normalized Item.
// Only the first play source's episodes are taken.
func mapDefaultDetailRow(row defaultDetailRow) Item {
	it := Item{
		ExternalID:         strconv.FormatUint(uint64(row.ID), 10),
		Title:              strings.TrimSpace(row.Title),
		Subtitle:           strings.TrimSpace(row.Subtitle),
		Description:        strings.TrimSpace(row.Description),
		Cover:              strings.TrimSpace(row.Cover),
		Year:               utils.Uint32ToInt32(row.Year),
		Region:             strings.TrimSpace(row.Region),
		Language:           strings.TrimSpace(row.Language),
		ReleaseDate:        strings.TrimSpace(row.ReleaseDate),
		ExternalCategoryID: row.CategoryID,
		Directors:          trimStrings(row.Directors),
		Actors:             trimStrings(row.Actors),
		VodTime:            strings.TrimSpace(row.CreatedAt),
	}
	if it.Cover != "" && it.Poster == "" {
		it.Poster = it.Cover
	}
	// only take the first source's episodes
	if len(row.Sources) > 0 {
		eps := make([]Episode, 0, len(row.Sources[0].Episodes))
		for _, ep := range row.Sources[0].Episodes {
			url := strings.TrimSpace(ep.URL)
			if url == "" {
				continue
			}
			num := utils.Uint32ToInt32(ep.Episode)
			if num <= 0 {
				num = 1
			}
			eps = append(eps, Episode{
				Number: num,
				Title:  strings.TrimSpace(ep.Title),
				URL:    url,
				Format: guessFormat(url),
			})
		}
		it.Episodes = eps
	}
	return it
}

func trimStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// defaultListItem is one compact item in the Open API video list.
type defaultListItem struct {
	ID         uint32 `json:"id"`
	Title      string `json:"title"`
	CategoryID uint32 `json:"category_id"`
	CreatedAt  string `json:"created_at"`
}

// defaultDetailRow is one full detail item in the Open API video detail.
type defaultDetailRow struct {
	ID          uint32             `json:"id"`
	Title       string             `json:"title"`
	Subtitle    string             `json:"subtitle"`
	Cover       string             `json:"cover"`
	CategoryID  uint32             `json:"category_id"`
	Year        uint32             `json:"year"`
	ReleaseDate string             `json:"release_date"`
	Region      string             `json:"region"`
	Language    string             `json:"language"`
	Description string             `json:"description"`
	Directors   []string           `json:"directors"`
	Actors      []string           `json:"actors"`
	Sources     []defaultSourceRow `json:"sources"`
	CreatedAt   string             `json:"created_at"`
}

type defaultSourceRow struct {
	ID       uint32              `json:"id"`
	Name     string              `json:"name"`
	Episodes []defaultEpisodeRow `json:"episodes"`
}

type defaultEpisodeRow struct {
	Episode uint32 `json:"episode"`
	Title   string `json:"title"`
	URL     string `json:"url"`
}

type defaultCategoryRow struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name"`
	ParentID uint32 `json:"parent_id"`
}

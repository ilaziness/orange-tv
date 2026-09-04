package seo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeSettings struct {
	groups map[string]map[string]model.SystemSettings
}

func (f *fakeSettings) LoadMapByGroup(ctx context.Context, group string) (map[string]model.SystemSettings, error) {
	if m, ok := f.groups[group]; ok {
		return m, nil
	}
	return map[string]model.SystemSettings{}, nil
}
func (f *fakeSettings) LoadGroupMaps(ctx context.Context, groups []string) (map[string]map[string]model.SystemSettings, error) {
	out := make(map[string]map[string]model.SystemSettings, len(groups))
	for _, g := range groups {
		m, err := f.LoadMapByGroup(ctx, g)
		if err != nil {
			return nil, err
		}
		out[g] = m
	}
	return out, nil
}
func (f *fakeSettings) MapGroupToResponse(group string, m map[string]model.SystemSettings) (any, error) {
	return service.NewSettingsService(nil, nil, zap.NewNop()).MapGroupToResponse(group, m)
}
func (f *fakeSettings) MapGroupsToResponse(groups []string, maps map[string]map[string]model.SystemSettings) (any, error) {
	return nil, nil
}
func (f *fakeSettings) UpsertMany(ctx context.Context, group string, upserts []repository.SettingUpsert) error {
	return nil
}

type fakeVideos struct {
	rows []repository.SitemapVideoRow
}

func (f *fakeVideos) CountOnlineForSitemap(ctx context.Context) (int, error) {
	return len(f.rows), nil
}
func (f *fakeVideos) ListOnlineForSitemap(ctx context.Context, afterID uint32, limit int) ([]repository.SitemapVideoRow, error) {
	out := make([]repository.SitemapVideoRow, 0, limit)
	for _, row := range f.rows {
		if row.ID <= afterID {
			continue
		}
		out = append(out, row)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
func (f *fakeVideos) OnlineIDAtOffset(ctx context.Context, offset int) (uint32, bool, error) {
	if offset < 0 || offset >= len(f.rows) {
		return 0, false, nil
	}
	return f.rows[offset].ID, true, nil
}

func setting(key, value string, typ uint8) model.SystemSettings {
	return model.SystemSettings{SettingKey: key, SettingValue: value, SettingType: typ}
}

func testSEOSettings() *fakeSettings {
	return &fakeSettings{groups: map[string]map[string]model.SystemSettings{
		constant.SettingGroupSEO: {
			constant.SettingSEOPublicBaseURL:   setting(constant.SettingSEOPublicBaseURL, "https://example.com", constant.SettingTypeString),
			constant.SettingSEOSitemapEnabled:  setting(constant.SettingSEOSitemapEnabled, "1", constant.SettingTypeBoolean),
			constant.SettingSEOLLMsEnabled:     setting(constant.SettingSEOLLMsEnabled, "1", constant.SettingTypeBoolean),
			constant.SettingSEOAllowAISearch:   setting(constant.SettingSEOAllowAISearch, "1", constant.SettingTypeBoolean),
			constant.SettingSEOAllowAITraining: setting(constant.SettingSEOAllowAITraining, "0", constant.SettingTypeBoolean),
			constant.SettingSEOLLMsIntro:       setting(constant.SettingSEOLLMsIntro, "hello\nworld", constant.SettingTypeString),
		},
		constant.SettingGroupSite: {
			constant.SettingSiteName: setting(constant.SettingSiteName, "Demo\nSite", constant.SettingTypeString),
		},
		constant.SettingGroupFeature: {
			constant.SettingFeatureLiveTVEnabled: setting(constant.SettingFeatureLiveTVEnabled, "1", constant.SettingTypeBoolean),
		},
	}}
}

func TestService_RobotsKeepsDisallowForAISearchBots(t *testing.T) {
	t.Parallel()
	svc := NewService(testSEOSettings(), &fakeVideos{}, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())

	robots, err := svc.Robots(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, robots.Status)
	body := string(robots.Body)
	require.Contains(t, body, "User-agent: GPTBot")
	gptIdx := strings.Index(body, "User-agent: GPTBot")
	nextIdx := strings.Index(body[gptIdx+1:], "User-agent:")
	block := body[gptIdx:]
	if nextIdx >= 0 {
		block = body[gptIdx : gptIdx+1+nextIdx]
	}
	require.Contains(t, block, "Disallow: /api/")
	require.NotContains(t, block, "Allow: /")
	require.Contains(t, body, "User-agent: Google-Extended")
	require.Contains(t, body, "Disallow: /\n")
	require.Contains(t, body, "Disallow: /swagger/")
	require.NotContains(t, body, "User-agent: *\nAllow: /")
}

func TestService_RobotsAndSingleSitemap(t *testing.T) {
	t.Parallel()
	videos := &fakeVideos{rows: []repository.SitemapVideoRow{
		{ID: 1, UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		{ID: 2, UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
	}}
	svc := NewService(testSEOSettings(), videos, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())

	sitemap, err := svc.Sitemap(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, sitemap.Status)
	xmlBody := string(sitemap.Body)
	require.Contains(t, xmlBody, "<urlset")
	require.Contains(t, xmlBody, "https://example.com/videos")
	require.Contains(t, xmlBody, "https://example.com/livetv")
	require.Contains(t, xmlBody, "https://example.com/video/1")

	llms, err := svc.LLMs(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, llms.Status)
	require.Contains(t, string(llms.Body), "# Demo Site")
	require.Contains(t, string(llms.Body), "> hello world")
	require.NotContains(t, string(llms.Body), "Demo\nSite")
}

func TestService_SitemapDisabledWithoutBaseURL(t *testing.T) {
	t.Parallel()
	settings := &fakeSettings{groups: map[string]map[string]model.SystemSettings{
		constant.SettingGroupSEO: {
			constant.SettingSEOSitemapEnabled: setting(constant.SettingSEOSitemapEnabled, "1", constant.SettingTypeBoolean),
			constant.SettingSEOLLMsEnabled:    setting(constant.SettingSEOLLMsEnabled, "1", constant.SettingTypeBoolean),
		},
	}}
	svc := NewService(settings, &fakeVideos{}, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())
	doc, err := svc.Sitemap(context.Background())
	require.NoError(t, err)
	require.Equal(t, 404, doc.Status)
	doc, err = svc.LLMs(context.Background())
	require.NoError(t, err)
	require.Equal(t, 404, doc.Status)
}

func TestService_SitemapVideosPageOutOfRange(t *testing.T) {
	t.Parallel()
	svc := NewService(testSEOSettings(), &fakeVideos{rows: []repository.SitemapVideoRow{{ID: 1}}}, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())
	doc, err := svc.SitemapVideos(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, 404, doc.Status)
	doc, err = svc.SitemapVideos(context.Background(), cache.MaxSEOSitemapVideoPages+1)
	require.NoError(t, err)
	require.Equal(t, 404, doc.Status)
}

func TestService_SitemapIndexAndVideoShards(t *testing.T) {
	prev := sitemapPageSize
	sitemapPageSize = 2
	t.Cleanup(func() { sitemapPageSize = prev })

	videos := &fakeVideos{rows: []repository.SitemapVideoRow{
		{ID: 1, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: 3, UpdatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: 4, UpdatedAt: time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)},
		{ID: 5, UpdatedAt: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)},
	}}
	svc := NewService(testSEOSettings(), videos, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())

	index, err := svc.Sitemap(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, index.Status)
	body := string(index.Body)
	require.Contains(t, body, "<sitemapindex")
	require.Contains(t, body, "https://example.com/sitemaps/pages.xml")
	require.Contains(t, body, "https://example.com/sitemaps/videos-1.xml")
	require.Contains(t, body, "https://example.com/sitemaps/videos-2.xml")
	require.Contains(t, body, "https://example.com/sitemaps/videos-3.xml")

	pages, err := svc.SitemapPages(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, pages.Status)
	require.Contains(t, string(pages.Body), "https://example.com/videos")
	require.NotContains(t, string(pages.Body), "/video/")

	page1, err := svc.SitemapVideos(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 200, page1.Status)
	require.Contains(t, string(page1.Body), "/video/1")
	require.Contains(t, string(page1.Body), "/video/2")
	require.NotContains(t, string(page1.Body), "/video/3")

	page2, err := svc.SitemapVideos(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, 200, page2.Status)
	require.Contains(t, string(page2.Body), "/video/3")
	require.Contains(t, string(page2.Body), "/video/4")

	page3, err := svc.SitemapVideos(context.Background(), 3)
	require.NoError(t, err)
	require.Equal(t, 200, page3.Status)
	require.Contains(t, string(page3.Body), "/video/5")

	doc, err := svc.SitemapVideos(context.Background(), 4)
	require.NoError(t, err)
	require.Equal(t, 404, doc.Status)
}

func TestService_SitemapIndexWhenStaticWouldOverflow(t *testing.T) {
	prev := sitemapPageSize
	sitemapPageSize = 3
	t.Cleanup(func() { sitemapPageSize = prev })

	// 2 static (/ , /videos) + livetv = 3 static; 1 video => 4 URLs > page size => index mode.
	videos := &fakeVideos{rows: []repository.SitemapVideoRow{{ID: 7}}}
	svc := NewService(testSEOSettings(), videos, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())

	index, err := svc.Sitemap(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, index.Status)
	require.Contains(t, string(index.Body), "<sitemapindex")
	require.Contains(t, string(index.Body), "/sitemaps/pages.xml")
	require.Contains(t, string(index.Body), "/sitemaps/videos-1.xml")

	pages, err := svc.SitemapPages(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, pages.Status)

	shard, err := svc.SitemapVideos(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 200, shard.Status)
	require.Contains(t, string(shard.Body), "/video/7")
}

func TestService_SitemapVideosCursorMissIs404(t *testing.T) {
	prev := sitemapPageSize
	sitemapPageSize = 1
	t.Cleanup(func() { sitemapPageSize = prev })

	svc := NewService(testSEOSettings(), &sparseVideos{count: 2}, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())

	doc, err := svc.SitemapVideos(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, 404, doc.Status)
}

// sparseVideos reports a multi-page count but cannot resolve page cursors.
type sparseVideos struct {
	count int
}

func (f *sparseVideos) CountOnlineForSitemap(ctx context.Context) (int, error) {
	return f.count, nil
}
func (f *sparseVideos) ListOnlineForSitemap(ctx context.Context, afterID uint32, limit int) ([]repository.SitemapVideoRow, error) {
	return []repository.SitemapVideoRow{{ID: afterID + 1}}, nil
}
func (f *sparseVideos) OnlineIDAtOffset(ctx context.Context, offset int) (uint32, bool, error) {
	return 0, false, nil
}

func TestService_LLMsFallbackSiteName(t *testing.T) {
	t.Parallel()
	settings := testSEOSettings()
	delete(settings.groups[constant.SettingGroupSite], constant.SettingSiteName)
	svc := NewService(settings, &fakeVideos{}, cache.NewManager(pkgcache.NewNopCache()), zap.NewNop())
	doc, err := svc.LLMs(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, doc.Status)
	require.Contains(t, string(doc.Body), "# 小橘TV")
}

func TestService_UsesCacheWhenEnabled(t *testing.T) {
	t.Parallel()
	mem, err := pkgcache.NewMemoryCache(pkgcache.MemoryCacheConfig{
		NumCounters: 1e4,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	require.NoError(t, err)
	c := cache.NewManager(mem)
	videos := &fakeVideos{rows: []repository.SitemapVideoRow{{ID: 9}}}
	svc := NewService(testSEOSettings(), videos, c, zap.NewNop())

	first, err := svc.Robots(context.Background())
	require.NoError(t, err)
	require.Equal(t, 200, first.Status)

	// Corrupt underlying settings; cached robots should still return previous body.
	settings := testSEOSettings()
	settings.groups[constant.SettingGroupSEO][constant.SettingSEOPublicBaseURL] = setting(constant.SettingSEOPublicBaseURL, "https://changed.example", constant.SettingTypeString)
	svc2 := NewService(settings, videos, c, zap.NewNop())
	second, err := svc2.Robots(context.Background())
	require.NoError(t, err)
	require.Contains(t, string(second.Body), "https://example.com/sitemap.xml")
	require.NotContains(t, string(second.Body), "changed.example")

	c.InvalidateSEO(context.Background())
	third, err := svc2.Robots(context.Background())
	require.NoError(t, err)
	require.Contains(t, string(third.Body), "https://changed.example/sitemap.xml")
}

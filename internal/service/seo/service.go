package seo

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/service"
	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// sitemapPageSize is the max URLs per sitemap shard (overridable in tests).
var sitemapPageSize = 50000

var (
	disallowPaths = []string{
		"/login", "/register", "/favorites", "/history", "/profile", "/play",
		"/api/", "/swagger/",
	}
	aiSearchBots = []string{
		"GPTBot",
		"OAI-SearchBot",
		"ChatGPT-User",
		"PerplexityBot",
		"ClaudeBot",
	}
	aiTrainingBots = []string{
		"Google-Extended",
		"Applebot-Extended",
	}
)

// Document is a generated SEO text/xml document.
type Document struct {
	Body        []byte
	ContentType string
	Status      int
}

// VideoStore provides the minimal video reads needed for sitemap generation.
type VideoStore interface {
	CountOnlineForSitemap(ctx context.Context) (int, error)
	ListOnlineForSitemap(ctx context.Context, afterID uint32, limit int) ([]repository.SitemapVideoRow, error)
	OnlineIDAtOffset(ctx context.Context, offset int) (id uint32, ok bool, err error)
}

// Service generates robots.txt, sitemap.xml and llms.txt on demand.
type Service interface {
	Robots(ctx context.Context) (Document, error)
	Sitemap(ctx context.Context) (Document, error)
	SitemapPages(ctx context.Context) (Document, error)
	SitemapVideos(ctx context.Context, page int) (Document, error)
	LLMs(ctx context.Context) (Document, error)
}

type seoService struct {
	settings service.SettingsService
	videos   VideoStore
	cache    *cache.Manager
	log      *zap.Logger
	sf       singleflight.Group
}

// NewService creates an SEO document service.
// cache may be a Nop-backed Manager when cache is disabled; Get then always misses and documents are generated live.
func NewService(settings service.SettingsService, videos VideoStore, c *cache.Manager, log *zap.Logger) Service {
	return &seoService{
		settings: settings,
		videos:   videos,
		cache:    c,
		log:      log,
	}
}

func (s *seoService) Robots(ctx context.Context) (Document, error) {
	return s.cached(ctx, cache.SEODocumentKey("robots"), func(ctx context.Context) (Document, error) {
		seoCfg, err := s.loadSEO(ctx)
		if err != nil {
			return Document{}, err
		}

		var b strings.Builder
		b.WriteString("User-agent: *\n")
		writeDisallows(&b, disallowPaths)
		b.WriteByte('\n')

		writeBotRules(&b, aiSearchBots, seoCfg.AllowAISearch)
		writeBotRules(&b, aiTrainingBots, seoCfg.AllowAITraining)

		if seoCfg.SitemapEnabled && seoCfg.PublicBaseURL != "" {
			b.WriteString("Sitemap: ")
			b.WriteString(seoCfg.PublicBaseURL)
			b.WriteString("/sitemap.xml\n")
		}

		return Document{
			Body:        []byte(b.String()),
			ContentType: "text/plain; charset=utf-8",
			Status:      http.StatusOK,
		}, nil
	})
}

func writeDisallows(b *strings.Builder, paths []string) {
	for _, path := range paths {
		b.WriteString("Disallow: ")
		b.WriteString(path)
		b.WriteByte('\n')
	}
}

func writeBotRules(b *strings.Builder, bots []string, allow bool) {
	for _, bot := range bots {
		b.WriteString("User-agent: ")
		b.WriteString(bot)
		b.WriteByte('\n')
		if allow {
			// 只写 Disallow：未禁止路径默认允许。避免再写 Allow:/ 被「先匹配先生效」的爬虫盖掉 Disallow。
			writeDisallows(b, disallowPaths)
		} else {
			b.WriteString("Disallow: /\n")
		}
		b.WriteByte('\n')
	}
}

func (s *seoService) Sitemap(ctx context.Context) (Document, error) {
	return s.cached(ctx, cache.SEODocumentKey("sitemap"), func(ctx context.Context) (Document, error) {
		seoCfg, err := s.requireSitemapBase(ctx)
		if err != nil {
			return Document{}, err
		}
		count, err := s.videos.CountOnlineForSitemap(ctx)
		if err != nil {
			s.log.Error("seo: count online videos failed", zap.Error(err))
			return Document{}, errcode.Wrap(errcode.DatabaseError, err)
		}

		liveTVEnabled, err := s.liveTVEnabled(ctx)
		if err != nil {
			return Document{}, err
		}
		statics := staticPaths(liveTVEnabled)

		// Reserve room for static page URLs so a single urlset never exceeds sitemapPageSize.
		if !needsSitemapIndex(count, len(statics)) {
			rows, listErr := s.videos.ListOnlineForSitemap(ctx, 0, sitemapPageSize)
			if listErr != nil {
				s.log.Error("seo: list online videos failed", zap.Error(listErr))
				return Document{}, errcode.Wrap(errcode.DatabaseError, listErr)
			}
			body, buildErr := buildURLSet(seoCfg.PublicBaseURL, statics, rows)
			if buildErr != nil {
				return Document{}, buildErr
			}
			return Document{Body: body, ContentType: "application/xml; charset=utf-8", Status: http.StatusOK}, nil
		}

		pages := (count + sitemapPageSize - 1) / sitemapPageSize
		if pages > cache.MaxSEOSitemapVideoPages {
			s.log.Warn("seo: sitemap video pages truncated",
				zap.Int("online_count", count),
				zap.Int("computed_pages", pages),
				zap.Int("max_pages", cache.MaxSEOSitemapVideoPages),
			)
			pages = cache.MaxSEOSitemapVideoPages
		}
		locs := make([]string, 0, pages+1)
		locs = append(locs, seoCfg.PublicBaseURL+"/sitemaps/pages.xml")
		for i := 1; i <= pages; i++ {
			locs = append(locs, fmt.Sprintf("%s/sitemaps/videos-%d.xml", seoCfg.PublicBaseURL, i))
		}
		body, buildErr := buildSitemapIndex(locs)
		if buildErr != nil {
			return Document{}, buildErr
		}
		return Document{Body: body, ContentType: "application/xml; charset=utf-8", Status: http.StatusOK}, nil
	})
}

func (s *seoService) SitemapPages(ctx context.Context) (Document, error) {
	return s.cached(ctx, cache.SEODocumentKey("sitemap-pages"), func(ctx context.Context) (Document, error) {
		seoCfg, err := s.requireSitemapBase(ctx)
		if err != nil {
			return Document{}, err
		}
		count, err := s.videos.CountOnlineForSitemap(ctx)
		if err != nil {
			s.log.Error("seo: count online videos failed", zap.Error(err))
			return Document{}, errcode.Wrap(errcode.DatabaseError, err)
		}
		liveTVEnabled, err := s.liveTVEnabled(ctx)
		if err != nil {
			return Document{}, err
		}
		statics := staticPaths(liveTVEnabled)
		if !needsSitemapIndex(count, len(statics)) {
			return Document{Status: http.StatusNotFound}, nil
		}
		body, buildErr := buildURLSet(seoCfg.PublicBaseURL, statics, nil)
		if buildErr != nil {
			return Document{}, buildErr
		}
		return Document{Body: body, ContentType: "application/xml; charset=utf-8", Status: http.StatusOK}, nil
	})
}

func (s *seoService) SitemapVideos(ctx context.Context, page int) (Document, error) {
	if page < 1 || page > cache.MaxSEOSitemapVideoPages {
		return Document{Status: http.StatusNotFound}, nil
	}
	return s.cached(ctx, cache.SEOSitemapVideosKey(page), func(ctx context.Context) (Document, error) {
		seoCfg, err := s.requireSitemapBase(ctx)
		if err != nil {
			return Document{}, err
		}
		count, err := s.videos.CountOnlineForSitemap(ctx)
		if err != nil {
			s.log.Error("seo: count online videos failed", zap.Error(err))
			return Document{}, errcode.Wrap(errcode.DatabaseError, err)
		}
		liveTVEnabled, err := s.liveTVEnabled(ctx)
		if err != nil {
			return Document{}, err
		}
		if !needsSitemapIndex(count, len(staticPaths(liveTVEnabled))) {
			return Document{Status: http.StatusNotFound}, nil
		}
		pages := (count + sitemapPageSize - 1) / sitemapPageSize
		if pages > cache.MaxSEOSitemapVideoPages {
			pages = cache.MaxSEOSitemapVideoPages
		}
		if page > pages {
			return Document{Status: http.StatusNotFound}, nil
		}

		afterID, err := s.pageAfterID(ctx, page)
		if err != nil {
			if errors.Is(err, errSEOPageCursorMiss) {
				return Document{Status: http.StatusNotFound}, nil
			}
			return Document{}, err
		}
		rows, listErr := s.videos.ListOnlineForSitemap(ctx, afterID, sitemapPageSize)
		if listErr != nil {
			s.log.Error("seo: list online videos failed", zap.Error(listErr))
			return Document{}, errcode.Wrap(errcode.DatabaseError, listErr)
		}
		if len(rows) == 0 {
			return Document{Status: http.StatusNotFound}, nil
		}
		body, buildErr := buildURLSet(seoCfg.PublicBaseURL, nil, rows)
		if buildErr != nil {
			return Document{}, buildErr
		}
		return Document{Body: body, ContentType: "application/xml; charset=utf-8", Status: http.StatusOK}, nil
	})
}

func (s *seoService) pageAfterID(ctx context.Context, page int) (uint32, error) {
	if page <= 1 {
		return 0, nil
	}
	if id, ok, err := s.cache.GetSEOPageCursor(ctx, page); err == nil && ok {
		return id, nil
	} else if err != nil && !errors.Is(err, pkgcache.ErrCacheMiss) {
		s.log.Warn("seo: page cursor cache get failed", zap.Int("page", page), zap.Error(err))
	}

	// Last id of previous page = item at 0-based offset (page-1)*pageSize - 1.
	offset := (page-1)*sitemapPageSize - 1
	id, ok, err := s.videos.OnlineIDAtOffset(ctx, offset)
	if err != nil {
		s.log.Error("seo: online id at offset failed", zap.Int("offset", offset), zap.Error(err))
		return 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	if !ok {
		// Never fall back to afterID=0 (would emit/cache page-1 content under page N).
		return 0, errSEOPageCursorMiss
	}
	if setErr := s.cache.SetSEOPageCursor(ctx, page, id); setErr != nil {
		s.log.Warn("seo: page cursor cache set failed", zap.Int("page", page), zap.Error(setErr))
	}
	return id, nil
}

// needsSitemapIndex reports whether static + video URLs would exceed one sitemap urlset.
func needsSitemapIndex(videoCount, staticCount int) bool {
	if videoCount < 0 {
		videoCount = 0
	}
	if staticCount < 0 {
		staticCount = 0
	}
	return videoCount+staticCount > sitemapPageSize
}

func (s *seoService) LLMs(ctx context.Context) (Document, error) {
	return s.cached(ctx, cache.SEODocumentKey("llms"), func(ctx context.Context) (Document, error) {
		seoCfg, err := s.loadSEO(ctx)
		if err != nil {
			return Document{}, err
		}
		if !seoCfg.LLMsEnabled || seoCfg.PublicBaseURL == "" {
			return Document{Status: http.StatusNotFound}, nil
		}
		site, err := s.loadSite(ctx)
		if err != nil {
			return Document{}, err
		}
		liveTVEnabled, err := s.liveTVEnabled(ctx)
		if err != nil {
			return Document{}, err
		}

		siteName := service.CollapseWhitespace(site.Name)
		if siteName == "" {
			siteName = "小橘TV"
		}
		intro := service.CollapseWhitespace(seoCfg.LLMsIntro)
		if intro == "" {
			intro = service.CollapseWhitespace(site.Description)
		}
		if intro == "" {
			intro = siteName + " 影视站点"
		}

		var b strings.Builder
		b.WriteString("# ")
		b.WriteString(siteName)
		b.WriteByte('\n')
		b.WriteByte('\n')
		b.WriteString("> ")
		b.WriteString(intro)
		b.WriteByte('\n')
		b.WriteByte('\n')
		b.WriteString("## 主要栏目\n")
		b.WriteString("- 首页: ")
		b.WriteString(seoCfg.PublicBaseURL)
		b.WriteString("/\n")
		b.WriteString("- 片库: ")
		b.WriteString(seoCfg.PublicBaseURL)
		b.WriteString("/videos\n")
		if liveTVEnabled {
			b.WriteString("- 电视直播: ")
			b.WriteString(seoCfg.PublicBaseURL)
			b.WriteString("/livetv\n")
		}
		b.WriteString("- 影片详情: ")
		b.WriteString(seoCfg.PublicBaseURL)
		b.WriteString("/video/{id}\n")
		b.WriteByte('\n')
		if seoCfg.SitemapEnabled {
			b.WriteString("## Sitemap\n")
			b.WriteString(seoCfg.PublicBaseURL)
			b.WriteString("/sitemap.xml\n")
		}

		return Document{
			Body:        []byte(b.String()),
			ContentType: "text/plain; charset=utf-8",
			Status:      http.StatusOK,
		}, nil
	})
}

func (s *seoService) requireSitemapBase(ctx context.Context) (dto.SEOSettings, error) {
	seoCfg, err := s.loadSEO(ctx)
	if err != nil {
		return dto.SEOSettings{}, err
	}
	if !seoCfg.SitemapEnabled || seoCfg.PublicBaseURL == "" {
		return dto.SEOSettings{}, errSEODisabled
	}
	return seoCfg, nil
}

var (
	errSEODisabled       = errors.New("seo sitemap disabled or base url missing")
	errSEOPageCursorMiss = errors.New("seo sitemap page cursor not found")
)

func (s *seoService) loadSEO(ctx context.Context) (dto.SEOSettings, error) {
	m, err := s.settings.LoadMapByGroup(ctx, constant.SettingGroupSEO)
	if err != nil {
		return dto.SEOSettings{}, err
	}
	return service.MapToSEOSettings(m), nil
}

func (s *seoService) loadSite(ctx context.Context) (dto.SiteSettings, error) {
	m, err := s.settings.LoadMapByGroup(ctx, constant.SettingGroupSite)
	if err != nil {
		return dto.SiteSettings{}, err
	}
	resp, err := s.settings.MapGroupToResponse(constant.SettingGroupSite, m)
	if err != nil {
		return dto.SiteSettings{}, err
	}
	site, ok := resp.(dto.SiteSettings)
	if !ok {
		return dto.SiteSettings{}, fmt.Errorf("unexpected site settings type")
	}
	return site, nil
}

func (s *seoService) liveTVEnabled(ctx context.Context) (bool, error) {
	m, err := s.settings.LoadMapByGroup(ctx, constant.SettingGroupFeature)
	if err != nil {
		return false, err
	}
	return service.BoolVal(m, constant.SettingFeatureLiveTVEnabled, false), nil
}

func staticPaths(liveTVEnabled bool) []string {
	paths := []string{"/", "/videos"}
	if liveTVEnabled {
		paths = append(paths, "/livetv")
	}
	return paths
}

func (s *seoService) cached(ctx context.Context, key string, gen func(context.Context) (Document, error)) (Document, error) {
	if entry, err := s.cache.GetSEODocument(ctx, key); err == nil && entry != nil {
		return Document{
			Body:        entry.Body,
			ContentType: entry.ContentType,
			Status:      entry.Status,
		}, nil
	} else if err != nil && !errors.Is(err, pkgcache.ErrCacheMiss) {
		s.log.Warn("seo: cache get failed", zap.String("key", key), zap.Error(err))
	}

	// Detach from caller cancel so a cancelled first request does not fail all singleflight waiters.
	genCtx := context.WithoutCancel(ctx)
	v, err, _ := s.sf.Do(key, func() (any, error) {
		// Re-check cache inside singleflight to avoid duplicate work after a peer fill.
		if entry, getErr := s.cache.GetSEODocument(genCtx, key); getErr == nil && entry != nil {
			return Document{
				Body:        entry.Body,
				ContentType: entry.ContentType,
				Status:      entry.Status,
			}, nil
		}

		doc, genErr := gen(genCtx)
		if genErr != nil {
			if errors.Is(genErr, errSEODisabled) {
				return Document{Status: http.StatusNotFound}, nil
			}
			return Document{}, genErr
		}
		if doc.Status == http.StatusNotFound {
			return doc, nil
		}

		if setErr := s.cache.SetSEODocument(genCtx, key, &cache.SEODocumentEntry{
			Body:        doc.Body,
			ContentType: doc.ContentType,
			Status:      doc.Status,
		}); setErr != nil {
			s.log.Warn("seo: cache set failed", zap.String("key", key), zap.Error(setErr))
		}
		return doc, nil
	})
	if err != nil {
		return Document{}, err
	}
	return v.(Document), nil
}

type urlSet struct {
	XMLName xml.Name   `xml:"urlset"`
	Xmlns   string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

type urlEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapIndex struct {
	XMLName  xml.Name       `xml:"sitemapindex"`
	Xmlns    string         `xml:"xmlns,attr"`
	Sitemaps []sitemapEntry `xml:"sitemap"`
}

type sitemapEntry struct {
	Loc string `xml:"loc"`
}

func buildURLSet(base string, paths []string, rows []repository.SitemapVideoRow) ([]byte, error) {
	set := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]urlEntry, 0, len(paths)+len(rows)),
	}
	for _, p := range paths {
		set.URLs = append(set.URLs, urlEntry{Loc: base + p})
	}
	for _, row := range rows {
		entry := urlEntry{Loc: fmt.Sprintf("%s/video/%d", base, row.ID)}
		if !row.UpdatedAt.IsZero() {
			entry.LastMod = row.UpdatedAt.UTC().Format(time.RFC3339)
		}
		set.URLs = append(set.URLs, entry)
	}
	return marshalXML(set)
}

func buildSitemapIndex(locs []string) ([]byte, error) {
	idx := sitemapIndex{
		Xmlns:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: make([]sitemapEntry, 0, len(locs)),
	}
	for _, loc := range locs {
		idx.Sitemaps = append(idx.Sitemaps, sitemapEntry{Loc: loc})
	}
	return marshalXML(idx)
}

func marshalXML(v any) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode sitemap xml: %w", err)
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

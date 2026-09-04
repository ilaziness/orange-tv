package seo

import "context"

// StubService is a no-op SEO Service for tests.
type StubService struct{}

func (StubService) Robots(ctx context.Context) (Document, error) {
	return Document{Body: []byte("User-agent: *\nDisallow: /api/\n"), ContentType: "text/plain; charset=utf-8", Status: 200}, nil
}
func (StubService) Sitemap(ctx context.Context) (Document, error) {
	return Document{Status: 404}, nil
}
func (StubService) SitemapPages(ctx context.Context) (Document, error) {
	return Document{Status: 404}, nil
}
func (StubService) SitemapVideos(ctx context.Context, page int) (Document, error) {
	return Document{Status: 404}, nil
}
func (StubService) LLMs(ctx context.Context) (Document, error) {
	return Document{Status: 404}, nil
}

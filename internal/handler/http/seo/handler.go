package seo

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	seosvc "github.com/ilaziness/orange-tv/internal/service/seo"
	"go.uber.org/zap"
)

// Handler serves robots.txt, sitemap and llms.txt.
type Handler struct {
	svc seosvc.Service
	log *zap.Logger
}

// NewHandler creates an SEO handler.
func NewHandler(svc seosvc.Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Robots
// @Summary robots.txt
// @Tags SEO
// @Produce plain
// @Success 200 {string} string
// @Router /robots.txt [get]
func (h *Handler) Robots(c *gin.Context) {
	h.serve(c, h.svc.Robots)
}

// Sitemap
// @Summary sitemap.xml
// @Tags SEO
// @Produce xml
// @Success 200 {string} string
// @Failure 404 {string} string
// @Router /sitemap.xml [get]
func (h *Handler) Sitemap(c *gin.Context) {
	h.serve(c, h.svc.Sitemap)
}

// SitemapPages
// @Summary sitemaps/pages.xml
// @Tags SEO
// @Produce xml
// @Success 200 {string} string
// @Failure 404 {string} string
// @Router /sitemaps/pages.xml [get]
func (h *Handler) SitemapPages(c *gin.Context) {
	h.serve(c, h.svc.SitemapPages)
}

// SitemapVideos
// @Summary sitemaps/videos-{n}.xml
// @Tags SEO
// @Produce xml
// @Param name path string true "形如 videos-1.xml；分页序号从 1 开始"
// @Success 200 {string} string
// @Failure 404 {string} string
// @Router /sitemaps/{name} [get]
func (h *Handler) SitemapVideos(c *gin.Context) {
	n, ok := parseSitemapVideoPage(c.Param("name"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	doc, svcErr := h.svc.SitemapVideos(c.Request.Context(), n)
	if svcErr != nil {
		h.log.Error("seo: sitemap videos failed", zap.Error(svcErr))
		c.Status(http.StatusInternalServerError)
		return
	}
	writeDocument(c, doc)
}

// parseSitemapVideoPage parses "videos-41.xml" into page number.
func parseSitemapVideoPage(name string) (int, bool) {
	const prefix = "videos-"
	const suffix = ".xml"
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix))
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

// LLMs
// @Summary llms.txt
// @Tags SEO
// @Produce plain
// @Success 200 {string} string
// @Failure 404 {string} string
// @Router /llms.txt [get]
func (h *Handler) LLMs(c *gin.Context) {
	h.serve(c, h.svc.LLMs)
}

func (h *Handler) serve(c *gin.Context, fn func(context.Context) (seosvc.Document, error)) {
	doc, err := fn(c.Request.Context())
	if err != nil {
		h.log.Error("seo: generate document failed", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	writeDocument(c, doc)
}

func writeDocument(c *gin.Context, doc seosvc.Document) {
	if doc.Status == http.StatusNotFound {
		c.Status(http.StatusNotFound)
		return
	}
	if doc.Status != 0 && doc.Status != http.StatusOK {
		c.Status(doc.Status)
		return
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, doc.ContentType, doc.Body)
}

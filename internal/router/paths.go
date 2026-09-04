package router

const (
	PathHealth    = "/health"
	PathReadiness = "/readiness"
	PathLiveness  = "/liveness"
	PathVersion   = "/version"

	PathClientV1   = "/api/client/v1"
	PathClientV2   = "/api/client/v2"
	PathAdminV1    = "/api/admin/v1"
	PathAdminV2    = "/api/admin/v2"
	PathInternalV1 = "/api/internal/v1"
	PathOpenV1     = "/api/open/v1"

	PathSwagger = "/swagger/*any"

	PathRobotsTxt       = "/robots.txt"
	PathSitemapXML      = "/sitemap.xml"
	PathSitemapPagesXML = "/sitemaps/pages.xml"
	// Gin cannot use videos-:n.xml (:n.xml becomes the param name). Match filename via :name.
	PathSitemapVideosXML = "/sitemaps/:name"
	PathLLMsTxt          = "/llms.txt"
)

// SystemPaths are public system endpoints that should skip auth, rate limits, and observability noise.
var SystemPaths = []string{PathHealth, PathReadiness, PathLiveness, PathVersion}

// RateLimitSkipPaths returns Gin FullPath patterns excluded from rate limiting.
// Includes health probes and SEO well-known documents (crawler-friendly).
func RateLimitSkipPaths() []string {
	paths := make([]string, 0, len(SystemPaths)+5)
	paths = append(paths, SystemPaths...)
	paths = append(paths,
		PathRobotsTxt,
		PathSitemapXML,
		PathSitemapPagesXML,
		PathSitemapVideosXML, // Gin FullPath for /sitemaps/:name
		PathLLMsTxt,
	)
	return paths
}

// DefaultJWTSkipPaths returns paths that skip JWT authentication by default.
// Includes public client API prefix, open resource API, and admin login; admin business routes still
// require JWT via RequireAuth (after a successful token parse).
func DefaultJWTSkipPaths() []string {
	paths := make([]string, 0, len(SystemPaths)+12)
	paths = append(paths, SystemPaths...)
	paths = append(paths,
		PathSwagger,
		PathClientV1+"/*", // 用户端公开只读 API（JWT 仍解析但不强制）
		PathClientV2+"/*",
		PathOpenV1+"/*", // 资源站开放 API（密钥自校验）
		PathAdminV1+"/auth/login",
		PathRobotsTxt,
		"/sitemap*", // /sitemap.xml、/sitemap-1.xml 等
		"/sitemaps/*",
		PathLLMsTxt,
	)
	return paths
}

// MergeJWTSkipPaths merges config jwt.skip_paths onto DefaultJWTSkipPaths.
// Defaults are always included; config may add extra paths. Duplicates are dropped.
func MergeJWTSkipPaths(configPaths []string) []string {
	base := DefaultJWTSkipPaths()
	if len(configPaths) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base)+len(configPaths))
	out := make([]string, 0, len(base)+len(configPaths))
	for _, p := range base {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, p := range configPaths {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

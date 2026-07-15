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

	// Client domain paths (v1)
	PathClientV1Categories   = PathClientV1 + "/categories"
	PathClientV1Videos       = PathClientV1 + "/videos"
	PathClientV1Search       = PathClientV1 + "/search"
	PathClientV1Live         = PathClientV1 + "/live"
	PathClientV1ThemeCurrent = PathClientV1 + "/theme/current"

	// Admin domain path prefixes (v1)
	PathAdminV1Auth           = PathAdminV1 + "/auth"
	PathAdminV1Categories     = PathAdminV1 + "/categories"
	PathAdminV1Videos         = PathAdminV1 + "/videos"
	PathAdminV1PlaySources    = PathAdminV1 + "/play-sources"
	PathAdminV1PlayEpisodes   = PathAdminV1 + "/play-episodes"
	PathAdminV1Directors      = PathAdminV1 + "/directors"
	PathAdminV1Actors         = PathAdminV1 + "/actors"
	PathAdminV1Tags           = PathAdminV1 + "/tags"
	PathAdminV1Live           = PathAdminV1 + "/live"
	PathAdminV1CollectSources = PathAdminV1 + "/collect-sources"
	PathAdminV1Collect        = PathAdminV1 + "/collect"
	PathAdminV1Admins         = PathAdminV1 + "/admins"
	PathAdminV1Groups         = PathAdminV1 + "/groups"
	PathAdminV1Users          = PathAdminV1 + "/users"
	PathAdminV1Settings       = PathAdminV1 + "/settings"
	PathAdminV1Themes         = PathAdminV1 + "/themes"

	PathSwagger = "/swagger/*any"
)

// SystemPaths are public system endpoints that should skip auth, rate limits, and observability noise.
var SystemPaths = []string{PathHealth, PathReadiness, PathLiveness, PathVersion}

// DefaultJWTSkipPaths returns paths that skip JWT authentication by default.
// Includes public client API prefix and admin login; admin business routes still
// require JWT via RequireAuth (after a successful token parse).
func DefaultJWTSkipPaths() []string {
	paths := make([]string, 0, len(SystemPaths)+4)
	paths = append(paths, SystemPaths...)
	paths = append(paths,
		PathSwagger,
		PathClientV1+"/*", // 用户端公开只读 API
		PathClientV2+"/*",
		PathAdminV1Auth+"/login",
	)
	return paths
}

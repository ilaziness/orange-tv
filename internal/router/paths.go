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
)

// SystemPaths are public system endpoints that should skip auth, rate limits, and observability noise.
var SystemPaths = []string{PathHealth, PathReadiness, PathLiveness, PathVersion}

// DefaultJWTSkipPaths returns paths that skip JWT authentication by default.
// Includes public client API prefix, open resource API, and admin login; admin business routes still
// require JWT via RequireAuth (after a successful token parse).
func DefaultJWTSkipPaths() []string {
	paths := make([]string, 0, len(SystemPaths)+5)
	paths = append(paths, SystemPaths...)
	paths = append(paths,
		PathSwagger,
		PathClientV1+"/*", // 用户端公开只读 API
		PathClientV2+"/*",
		PathOpenV1+"/*", // 资源站开放 API（密钥自校验）
		PathAdminV1+"/auth/login",
	)
	return paths
}

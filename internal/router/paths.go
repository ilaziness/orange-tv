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

	PathClientV1Users   = PathClientV1 + "/users"
	PathClientV2Users   = PathClientV2 + "/users"
	PathAdminV1Users    = PathAdminV1 + "/users"
	PathInternalV1Users = PathInternalV1 + "/users"

	PathSwagger = "/swagger/*any"
)

// SystemPaths are public system endpoints that should skip auth, rate limits, and observability noise.
var SystemPaths = []string{PathHealth, PathReadiness, PathLiveness, PathVersion}

// DefaultJWTSkipPaths returns paths that skip JWT authentication by default.
func DefaultJWTSkipPaths() []string {
	paths := make([]string, len(SystemPaths)+1)
	copy(paths, SystemPaths)
	paths[len(SystemPaths)] = PathSwagger
	return paths
}

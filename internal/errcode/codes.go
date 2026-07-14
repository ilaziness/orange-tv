// Package errcode provides error codes and error types following PRD specification.
// Error code format: {3-digit module code}{4-digit business code}
// Module codes: 100=General, 200=User, 300=Auth, 400=Order, 500=Payment, 900=System
// Business codes: 0001-0999=General, 1000-1999=Business Logic, 2000-2999=Permission, 5000-5999=System
package errcode

// 预定义错误码（遵循 {模块码}{业务码} 格式）
// 命名规范：模块+具体错误，如 ParamError（参数错误）、UserNotFound（用户不存在）
var (
	// 通用模块 (100xxxx)
	ParamError       = &Code{1000001, "参数错误", 400, nil}
	DataNotFound     = &Code{1000002, "数据不存在", 404, nil}
	ValidationError  = &Code{1000003, "验证失败", 422, nil}
	ResourceNotFound = &Code{1000004, "资源不存在", 404, nil}
	RequestTimeout   = &Code{1000005, "请求超时", 408, nil}
	TooManyRequests  = &Code{1000006, "请求过于频繁", 429, nil}

	// 用户模块 (200xxxx)
	UserNotFound      = &Code{2000001, "用户不存在", 404, nil}
	UserAlreadyExists = &Code{2000002, "用户已存在", 409, nil}
	UserDisabled      = &Code{2000003, "用户账号已禁用", 403, nil}
	InvalidUserStatus = &Code{2000004, "用户状态无效", 400, nil}

	// 认证模块 (300xxxx)
	AuthFailed             = &Code{3000001, "认证失败", 401, nil}
	TokenExpired           = &Code{3000002, "Token已过期", 401, nil}
	InsufficientPermission = &Code{3000003, "权限不足", 403, nil}
	InvalidToken           = &Code{3000004, "无效的Token", 401, nil}
	TokenRevoked           = &Code{3000005, "Token已吊销", 401, nil}

	// 系统模块 (900xxxx)
	InternalError        = &Code{9000001, "服务器内部错误", 500, nil}
	DatabaseError        = &Code{9000002, "数据库错误", 500, nil}
	CacheError           = &Code{9000003, "缓存错误", 500, nil}
	ExternalServiceError = &Code{9000004, "外部服务错误", 502, nil}
	ServiceUnavailable   = &Code{9000005, "服务不可用", 503, nil}
)

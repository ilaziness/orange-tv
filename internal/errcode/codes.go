// Package errcode provides error codes and error types following PRD specification.
// Error code format: {3-digit module code}{4-digit business code}
// Module codes: 100=General, 200=User, 300=Auth, 400=Content, 900=System
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
	UserNotFound       = &Code{2000001, "用户不存在", 404, nil}
	UserAlreadyExists  = &Code{2000002, "用户已存在", 409, nil}
	UserDisabled       = &Code{2000003, "用户账号已禁用", 403, nil}
	InvalidUserStatus  = &Code{2000004, "用户状态无效", 400, nil}
	AdminNotFound      = &Code{2000010, "管理员不存在", 404, nil}
	AdminAlreadyExists = &Code{2000011, "管理员已存在", 409, nil}
	UserGroupNotFound  = &Code{2000020, "用户组不存在", 404, nil}
	UserGroupNameDup   = &Code{2000021, "用户组名称已存在", 409, nil}
	FavoriteExists     = &Code{2000030, "已收藏该影视", 409, nil}
	FavoriteNotFound   = &Code{2000031, "未收藏该影视", 404, nil}
	CommentNotFound    = &Code{2000040, "评论不存在", 404, nil}
	CommentTooLong     = &Code{2000041, "评论内容过长", 400, nil}
	BannerNotFound     = &Code{2000050, "Banner不存在", 404, nil}

	// 认证模块 (300xxxx)
	AuthFailed             = &Code{3000001, "认证失败", 401, nil}
	TokenExpired           = &Code{3000002, "Token已过期", 401, nil}
	InsufficientPermission = &Code{3000003, "权限不足", 403, nil}
	InvalidToken           = &Code{3000004, "无效的Token", 401, nil}
	TokenRevoked           = &Code{3000005, "Token已吊销", 401, nil}
	InvalidCredentials     = &Code{3000006, "用户名或密码错误", 401, nil}
	AdminDisabled          = &Code{3000007, "管理员账号已禁用", 403, nil}

	// 内容模块 (400xxxx)
	CategoryNotFound     = &Code{4000001, "分类不存在", 404, nil}
	CategoryNameExists   = &Code{4000002, "分类名称已存在", 409, nil}
	CategoryHasChildren  = &Code{4000003, "分类下仍有子分类，无法删除", 409, nil}
	CategoryHasVideos    = &Code{4000004, "分类下仍有影视，无法删除", 409, nil}
	CategoryCycle        = &Code{4000005, "分类父级设置会导致循环", 400, nil}
	VideoNotFound        = &Code{4000010, "影视不存在", 404, nil}
	DirectorNotFound     = &Code{4000020, "导演不存在", 404, nil}
	DirectorNameExists   = &Code{4000021, "导演名称已存在", 409, nil}
	DirectorInUse        = &Code{4000022, "导演仍被影视引用，无法删除", 409, nil}
	ActorNotFound        = &Code{4000030, "演员不存在", 404, nil}
	ActorNameExists      = &Code{4000031, "演员名称已存在", 409, nil}
	ActorInUse           = &Code{4000032, "演员仍被影视引用，无法删除", 409, nil}
	TagNotFound          = &Code{4000040, "标签不存在", 404, nil}
	TagNameExists        = &Code{4000041, "标签名称已存在", 409, nil}
	TagInUse             = &Code{4000042, "标签仍被影视引用，无法删除", 409, nil}
	PlaySourceNotFound   = &Code{4000050, "播放源不存在", 404, nil}
	PlaySourceNameExists = &Code{4000051, "播放源名称已存在", 409, nil}
	PlaySourceInUse      = &Code{4000052, "播放源仍被剧集引用，无法删除", 409, nil}
	PlayEpisodeNotFound  = &Code{4000060, "剧集不存在", 404, nil}
	PlayEpisodeDuplicate = &Code{4000061, "同一影视与播放源下集数已存在", 409, nil}

	// 直播 (40007xx)
	LiveChannelNotFound = &Code{4000070, "直播频道不存在", 404, nil}

	// 采集 (40008xx)
	CollectSourceNotFound   = &Code{4000080, "采集源不存在", 404, nil}
	CollectSourceDisabled   = &Code{4000081, "采集源已禁用", 400, nil}
	CollectAlreadyRunning   = &Code{4000082, "该采集源正在执行中", 409, nil}
	CollectNotRunning       = &Code{4000083, "该采集源未在执行", 400, nil}
	CollectInvalidCron      = &Code{4000084, "定时采集 cron 表达式无效", 400, nil}
	CollectFetchFailed      = &Code{4000085, "采集拉取失败", 502, nil}
	CollectParseFailed      = &Code{4000086, "采集数据解析失败", 422, nil}
	CollectCategoryMapEmpty = &Code{4000087, "请先配置分类映射", 400, nil}

	// 主题 (40009xx)
	ThemeNotFound      = &Code{4000090, "主题不存在", 404, nil}
	ThemeIdentifierDup = &Code{4000091, "主题标识已存在", 409, nil}
	ThemeCannotDelete  = &Code{4000092, "当前激活或默认主题不可删除", 409, nil}
	ThemeNoActive      = &Code{4000093, "未配置激活主题", 404, nil}

	// 系统模块 (900xxxx)
	InternalError         = &Code{9000001, "服务器内部错误", 500, nil}
	DatabaseError         = &Code{9000002, "数据库错误", 500, nil}
	CacheError            = &Code{9000003, "缓存错误", 500, nil}
	ExternalServiceError  = &Code{9000004, "外部服务错误", 502, nil}
	ServiceUnavailable    = &Code{9000005, "服务不可用", 503, nil}
	SettingInvalid        = &Code{9000010, "系统设置无效", 400, nil}
	ResourceAPIDisabled   = &Code{9000011, "资源站 API 已关闭", 403, nil}
	ResourceAPIKeyInvalid = &Code{9000012, "资源站 API 密钥无效", 401, nil}
)

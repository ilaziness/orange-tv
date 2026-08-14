package constant

// 客户端端点类型。
const (
	// ClientTypeWeb 网页端。
	ClientTypeWeb = "web"
	// ClientTypeApp 移动 App 端。
	ClientTypeApp = "app"
	// ClientTypeTV 电视端。
	ClientTypeTV = "tv"
	// ClientTypeDesktop 桌面端。
	ClientTypeDesktop = "desktop"

	// ClientTypeHeader 客户端声明端点类型的请求头。
	ClientTypeHeader = "X-Client-Type"
)

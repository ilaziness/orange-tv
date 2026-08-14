package http

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/clienttype"
	"github.com/ilaziness/orange-tv/internal/constant"
)

// tvUAMarkers 识别 TV 平台的 UA 关键词。
// 必须用明确平台词，避免裸 "tv" 子串误匹配 iptv/atv/mtv 等任意含 tv 的 UA。
var tvUAMarkers = []string{
	"androidtv", "android tv", "googletv", "tizen", "webos", "roku",
	"firetv", "smarttv", "sm-tv", "vidaa", "viera", "nettv", "bravia",
}

// ClientTypeMiddleware 识别客户端端类型并注入 context。
// systemPaths 中的系统路径（如 /health）跳过识别，避免给系统端点附加端类型响应头噪音。
// 识别优先级：X-Client-Type 头（应用层契约，Flutter 端必须注入）> UA 兜底。
// UA 兜底仅对浏览器/原生 UA 有效；Flutter 的 "Dart/" UA 无法区分端，一律回退 web（安全默认）。
// 同时响应头回显 X-Client-Type，便于前端调试与排查。
func ClientTypeMiddleware(systemPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if slices.Contains(systemPaths, c.Request.URL.Path) {
			c.Next()
			return
		}

		t := DetectClientType(c.GetHeader(constant.ClientTypeHeader), c.GetHeader("User-Agent"))
		c.Header(constant.ClientTypeHeader, t)
		c.Request = c.Request.WithContext(clienttype.WithContext(c.Request.Context(), t))
		c.Next()
	}
}

// DetectClientType 纯函数，便于单测。
func DetectClientType(header, userAgent string) string {
	if header != "" {
		switch strings.ToLower(header) {
		case constant.ClientTypeApp, constant.ClientTypeTV,
			constant.ClientTypeDesktop, constant.ClientTypeWeb:
			return strings.ToLower(header)
		}
	}
	ua := strings.ToLower(userAgent)
	switch {
	// TV（明确的 TV 平台关键词；不匹配裸 "tv" 子串，避免 iptv/atv 等误判）
	case matchTVUA(ua):
		return constant.ClientTypeTV
	// 移动端（原生 App 或浏览器）
	case strings.Contains(ua, "android") || strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") ||
		strings.Contains(ua, "mobile"):
		return constant.ClientTypeApp
	// 桌面浏览器 → web
	case strings.Contains(ua, "mozilla") && !strings.Contains(ua, "mobile"):
		return constant.ClientTypeWeb
	default:
		// 含 "Dart/" 在内的未知 UA → 安全默认 web
		return constant.ClientTypeWeb
	}
}

// matchTVUA 判断 UA 是否包含明确的 TV 平台关键词。
func matchTVUA(ua string) bool {
	for _, marker := range tvUAMarkers {
		if strings.Contains(ua, marker) {
			return true
		}
	}
	return false
}

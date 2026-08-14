// Package clienttype 提供客户端端点类型（web/app/tv/desktop）的 context 存取。
// 由端识别中间件写入，service 等业务层只读 context，避免跨层依赖 HTTP 层。
package clienttype

import (
	"context"

	"github.com/ilaziness/orange-tv/internal/constant"
)

type ctxKey struct{}

// WithContext 返回携带端类型的新 context。
func WithContext(ctx context.Context, t string) context.Context {
	return context.WithValue(ctx, ctxKey{}, t)
}

// FromContext 读取端类型；缺失或未知时默认 web（最安全，不泄露流地址等敏感字段）。
func FromContext(ctx context.Context) string {
	t, _ := ctx.Value(ctxKey{}).(string)
	switch t {
	case constant.ClientTypeApp, constant.ClientTypeTV, constant.ClientTypeDesktop:
		return t
	default:
		// web 与未知都归 web
		return constant.ClientTypeWeb
	}
}

// IsStreamEnabled 端是否需要返回直播流链接（web 不需要）。
func IsStreamEnabled(ctx context.Context) bool {
	return FromContext(ctx) != constant.ClientTypeWeb
}

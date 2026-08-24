// Package captcha 提供验证码能力，定义获取与校验的统一接口约定。
//
// 当前提供基于 github.com/mojocn/base64Captcha 的图像验证码实现（png.go），
// 未来可扩展其他验证码类型（如 Cloudflare Turnstile、reCAPTCHA 等基于
// token 的非图像验证码），只需实现 Captcha 接口即可接入。
//
// 设计目标：
//   - 业务侧只依赖 Captcha 接口，屏蔽底层渲染与存储实现，便于后期扩展
//     其他验证码类型或更换底层库。
//   - 存储可从外部注入；未注入时使用 base64Captcha 内置内存 store，
//     保证应用未配置 Cache 时验证码仍可用。
package captcha

import "context"

// Image 一次图像验证码的展示载体。
//
// 对非图像验证码（如基于 token 的 Cloudflare Turnstile），
// 实现可把 Challenge URL / sitekey 等放入 Content，由前端按约定渲染。
type Image struct {
	// ID 验证码唯一标识，前端提交校验时需原样带回。
	ID string
	// Content 展示内容，data URI 形式（如 data:image/png;base64,...），
	// 前端可直接作为 <img src> 使用。
	Content string
	// ExpiresIn 有效期（秒），供前端倒计时刷新。
	ExpiresIn int
}

// Captcha 验证码服务约定：获取验证码 + 校验验证码。
//
// 实现需保证校验后验证码作废（一次性），无论校验成功与否，
// 同一 ID 不可二次校验，防止重放与暴力穷举。
type Captcha interface {
	// Generate 生成并存储一个验证码，返回展示载体与标识。
	// scene 为业务场景（如 login/register），用于隔离不同场景的验证码。
	Generate(ctx context.Context, scene string) (*Image, error)

	// Verify 校验验证码答案。
	// 校验后无论成败均作废该验证码（一次性）。
	// 返回的 error 应能被 errors.Is 识别：ErrNotFound / ErrNotMatched。
	Verify(ctx context.Context, id, answer string) error
}

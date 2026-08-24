package captcha

import "errors"

// 验证码相关错误。调用方通过 errors.Is 判断具体失败类型，
// 映射为对应的业务错误码。
var (
	// ErrNotFound 验证码不存在或已作废（已校验过 / 已过期 / 从未生成）。
	ErrNotFound = errors.New("captcha: not found or expired")

	// ErrNotMatched 验证码答案不匹配。
	ErrNotMatched = errors.New("captcha: answer not matched")
)

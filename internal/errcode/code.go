// Package errcode provides error codes and error types following PRD specification.
// Error code format: {3-digit module code}{4-digit business code}
// Module codes: 100=General, 200=User, 300=Auth, 400=Order, 500=Payment, 900=System
// Business codes: 0001-0999=General, 1000-1999=Business Logic, 2000-2999=Permission, 5000-5999=System
package errcode

import (
	"errors"
	"fmt"
)

// Code 错误码类型，实现 error 接口
type Code struct {
	Code       int    // 错误码，遵循 {3位模块码}{4位业务码} 格式
	Message    string // 错误消息
	HTTPStatus int    // HTTP 状态码
	Cause      error  // 底层错误（可选）
}

// Error 实现 error 接口
func (c *Code) Error() string {
	if c.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", c.Code, c.Message, c.Cause)
	}
	return fmt.Sprintf("[%d] %s", c.Code, c.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (c *Code) Unwrap() error {
	return c.Cause
}

// GetHTTPStatus 获取 HTTP 状态码
func (c *Code) GetHTTPStatus() int {
	return c.HTTPStatus
}

// Wrap 包装底层错误
// 用于在已知错误码的基础上包装底层错误
func Wrap(code *Code, cause error) *Code {
	return &Code{
		Code:       code.Code,
		Message:    code.Message,
		HTTPStatus: code.HTTPStatus,
		Cause:      cause,
	}
}

// WithMessage 自定义消息
// 用于在已知错误码的基础上自定义错误消息
func WithMessage(code *Code, message string) *Code {
	if message == "" {
		return code
	}
	return &Code{
		Code:       code.Code,
		Message:    message,
		HTTPStatus: code.HTTPStatus,
		Cause:      code.Cause,
	}
}

// As 类型断言（类似 errors.As）
// 用于检查错误是否为 Code 类型
func As(err error) (*Code, bool) {
	var code *Code
	if errors.As(err, &code) {
		return code, true
	}
	return nil, false
}

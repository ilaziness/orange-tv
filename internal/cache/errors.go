package cache

import (
	"errors"
	"fmt"
)

var (
	// ErrCacheMiss 缓存未命中
	ErrCacheMiss = errors.New("cache: key not found")
	// ErrCacheUnavailable 缓存服务不可用
	ErrCacheUnavailable = errors.New("cache: service unavailable")
	// ErrInvalidKey 无效的缓存键
	ErrInvalidKey = errors.New("cache: invalid key")
)

// CacheError 缓存错误类型
type CacheError struct {
	Code    string
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *CacheError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("cache error [%s]: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("cache error [%s]: %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *CacheError) Unwrap() error {
	return e.Err
}

// IsCacheMiss 检查是否为缓存未命中错误
func IsCacheMiss(err error) bool {
	return errors.Is(err, ErrCacheMiss)
}

// IsInvalidKey 检查是否为无效键错误
func IsInvalidKey(err error) bool {
	return errors.Is(err, ErrInvalidKey)
}

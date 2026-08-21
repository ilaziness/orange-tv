// Package lock 提供通用分布式锁能力，支持 Redis 实现与本地内存实现，
// 调用方通过 Locker 接口屏蔽底层差异。
package lock

import (
	"context"
	"errors"
	"time"
)

// ErrLockNotHeld 表示锁未获取到（被他人持有）或已丢失。
// 调用方据此判断是「并发冲突」还是「真实错误」。
var ErrLockNotHeld = errors.New("lock: not held")

// Lock 表示已获取到的锁实例，调用方在业务完成后应调用 Release 释放。
// 实现需保证 Release 幂等：锁已过期或被他人抢占时返回 nil。
type Lock interface {
	// Release 释放锁。context 用于底层网络调用（Redis 场景）。
	Release(ctx context.Context) error
}

// Locker 锁容器，提供获取锁的统一入口。
// 抢锁失败（已被持有）返回 ErrLockNotHeld；其他底层错误原样返回。
type Locker interface {
	// Lock 尝试获取指定 key 的锁。opts 可指定 TTL 等。
	// 同一 key 同时只有一个调用方能成功获取；其余返回 ErrLockNotHeld。
	Lock(ctx context.Context, key string, opts ...Option) (Lock, error)
}

// Options 锁选项。
type Options struct {
	// TTL 锁的存活时间，到期自动释放（防进程崩溃后死锁）。
	TTL time.Duration
}

// Option 配置函数。
type Option func(*Options)

// WithTTL 设置锁的 TTL。
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) { o.TTL = ttl }
}

// defaultTTL 默认 TTL，防调用方未指定时锁永久持有。
const defaultTTL = 30 * time.Second

// applyOptions 合并选项并填充默认值。
func applyOptions(opts []Option) *Options {
	o := &Options{TTL: defaultTTL}
	for _, opt := range opts {
		opt(o)
	}
	if o.TTL <= 0 {
		o.TTL = defaultTTL
	}
	return o
}

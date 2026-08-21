package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// memoryLocker 进程内互斥锁实现。用于未启用 Redis 时的降级方案。
// 仅在单进程内保证互斥，多实例部署时不能跨进程互斥，需配合 DB 唯一索引兜底。
type memoryLocker struct {
	mu      sync.Mutex
	entries map[string]*memoryEntry
}

// memoryEntry 单个锁条目，含 token（持有者标识）与过期时间。
type memoryEntry struct {
	token    string
	expireAt time.Time
}

// NewMemoryLocker 创建本地内存 Locker。
func NewMemoryLocker() Locker {
	return &memoryLocker{entries: make(map[string]*memoryEntry)}
}

// Lock 尝试获取 key 的锁。先惰性清理该 key 的过期条目，再判断是否可获取。
// 未占用则写入 token 与 expireAt 并返回；已占用返回 ErrLockNotHeld。
func (l *memoryLocker) Lock(ctx context.Context, key string, opts ...Option) (Lock, error) {
	// 仅占位，避免 ctx unused 警告（内存锁不涉及 IO）。
	_ = ctx

	o := applyOptions(opts)
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	// 惰性清理：若该 key 存在且已过期，视为可重新获取。
	if e, ok := l.entries[key]; ok && now.Before(e.expireAt) {
		return nil, ErrLockNotHeld
	}

	l.entries[key] = &memoryEntry{
		token:    token,
		expireAt: now.Add(o.TTL),
	}
	return &memoryLock{locker: l, key: key, token: token}, nil
}

// memoryLock 实现内存锁的释放。释放时校验 token，避免误释放他人持有的锁。
type memoryLock struct {
	locker *memoryLocker
	key    string
	token  string
}

// Release 释放锁。token 不匹配视为锁已过期被他人抢占，幂等返回 nil。
func (l *memoryLock) Release(ctx context.Context) error {
	_ = ctx

	l.locker.mu.Lock()
	defer l.locker.mu.Unlock()

	e, ok := l.locker.entries[l.key]
	if !ok {
		return nil
	}
	if e.token != l.token {
		// 锁已被他人重新获取，幂等返回 nil。
		return nil
	}
	delete(l.locker.entries, l.key)
	return nil
}

// generateToken 生成 16 字节随机 hex 串作为锁持有者标识。
// 用 crypto/rand 避免引入额外依赖。
func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

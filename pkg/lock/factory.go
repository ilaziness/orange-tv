package lock

import "github.com/redis/go-redis/v9"

// LockerFactoryOptions 锁工厂参数。
// RedisClient 为 nil（含 typed-nil，如 (*redis.Client)(nil)）时降级为内存锁，
// 避免在未启用 Redis 的部署中崩溃。
type LockerFactoryOptions struct {
	RedisClient redis.UniversalClient
}

// LockerFactory 锁工厂。
type LockerFactory struct {
	opts LockerFactoryOptions
}

// NewLockerFactory 创建锁工厂。
func NewLockerFactory(opts LockerFactoryOptions) *LockerFactory {
	return &LockerFactory{opts: opts}
}

// Create 根据是否传入 Redis client 创建对应实现的 Locker。
// RedisClient 为 nil（含 typed-nil）时降级为 memoryLocker。
//
// 注意：直接把 (*redis.Client)(nil) 赋给 redis.UniversalClient 接口会得到
// 一个非 nil 的接口值（type=*redis.Client, value=nil），简单的 == nil 判断
// 无法识别，需要用类型断言检测底层指针是否为 nil，否则会创建出持有 nil
// client 的 redisLocker，运行时 nil 解引用 panic。
func (f *LockerFactory) Create() Locker {
	if isNilRedisClient(f.opts.RedisClient) {
		return NewMemoryLocker()
	}
	return NewRedisLocker(f.opts.RedisClient)
}

// isNilRedisClient 判断 UniversalClient 是否为 nil。
// 同时识别 untyped nil 和 typed-nil（如 (*redis.Client)(nil)）。
func isNilRedisClient(c redis.UniversalClient) bool {
	if c == nil {
		return true
	}
	switch v := c.(type) {
	case *redis.Client:
		return v == nil
	case *redis.ClusterClient:
		return v == nil
	case *redis.Ring:
		return v == nil
	}
	return false
}

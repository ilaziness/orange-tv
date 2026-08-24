package captcha

import (
	"context"
	"fmt"
	"strings"
	"time"

	pkgcache "github.com/ilaziness/orange-tv/pkg/cache"
	"github.com/mojocn/base64Captcha"
)

// cacheStoreKeyPrefix 外部 store 的 key 前缀，避免与其他业务缓存冲突。
const cacheStoreKeyPrefix = "captcha:"

// cacheStore 将 pkg/cache.Cache 适配为 base64Captcha.Store，
// 用于多实例部署下通过 Redis 共享验证码。
//
// 注意：
//   - base64Captcha.Store 的 Get/Set/Verify 接口不接收 context，
//     这里统一使用 context.Background()，受限于第三方库接口设计。
//   - base64Captcha.Store.Get(id, clear) 不返回 error，底层 cache 故障时
//     按「不存在」处理（返回空串），由上层 Verify 映射为 ErrNotFound，
//     保证校验安全（宁可拒绝也不放行）。
//   - pngCaptcha.Verify 采用「先 Get(id,false) 探测，再 Get(id,true) 删除」
//     两步操作，在 Redis 等远程存储上非原子。验证码场景为单用户单次
//     校验，并发探测概率极低；若未来需要严格原子，可在此实现内用
//     Redis Lua 脚本做 get-and-delete，接口不变。
type cacheStore struct {
	cache pkgcache.Cache
	ttl   time.Duration
}

// NewCacheStore 基于 pkg/cache.Cache 创建 base64Captcha.Store 适配器。
// ttl 为验证码有效期，与验证码本身的有效期保持一致。
func NewCacheStore(c pkgcache.Cache, ttl time.Duration) base64Captcha.Store {
	if ttl <= 0 {
		ttl = defaultExpireTTL
	}
	return &cacheStore{cache: c, ttl: ttl}
}

// WithCacheStore 配置基于 pkg/cache.Cache 的共享存储（Redis/内存），
// 用于多实例部署下通过外部缓存共享验证码。
//
// ttl <= 0 时使用默认有效期，与验证码本身的有效期一致。
// 调用方无需感知 base64Captcha.Store 类型，避免第三方库类型泄露到组装层。
func WithCacheStore(c pkgcache.Cache, ttl time.Duration) Option {
	return func(o *Options) {
		o.Store = NewCacheStore(c, ttl)
	}
}

// Set 存储 id 对应的答案。
func (s *cacheStore) Set(id string, value string) error {
	if id == "" {
		return nil
	}
	ctx := context.Background()
	if err := s.cache.Set(ctx, s.key(id), value, s.ttl); err != nil {
		return fmt.Errorf("captcha: cache store set: %w", err)
	}
	return nil
}

// Get 读取 id 对应的答案。clear=true 时删除（一次性）。
func (s *cacheStore) Get(id string, clear bool) string {
	if id == "" {
		return ""
	}
	ctx := context.Background()
	v, err := s.cache.Get(ctx, s.key(id))
	if err != nil || v == nil {
		// 未命中或底层故障，统一按不存在处理。
		return ""
	}
	ans, ok := v.(string)
	if !ok {
		return ""
	}
	if clear {
		_ = s.cache.Delete(ctx, s.key(id))
	}
	return ans
}

// Verify 校验答案，clear=true 时校验后删除。
// 与 base64Captcha 内置 store 保持一致：TrimSpace + 大小写不敏感。
func (s *cacheStore) Verify(id, answer string, clear bool) bool {
	vv := s.Get(id, clear)
	return strings.EqualFold(strings.TrimSpace(vv), strings.TrimSpace(answer))
}

// key 生成带前缀的存储键。
func (s *cacheStore) key(id string) string {
	return cacheStoreKeyPrefix + id
}

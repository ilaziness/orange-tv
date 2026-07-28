package cache

import (
	"strings"
)

// DefaultKeyGenerator 默认缓存键生成器
type DefaultKeyGenerator struct{}

// NewDefaultKeyGenerator 创建默认键生成器
func NewDefaultKeyGenerator() *DefaultKeyGenerator {
	return &DefaultKeyGenerator{}
}

// Generate 生成缓存键
// 例如: Generate("user", "123", "profile") -> "user:123:profile"
func (g *DefaultKeyGenerator) Generate(prefix string, parts ...string) string {
	if len(parts) == 0 {
		return prefix
	}

	allParts := append([]string{prefix}, parts...)
	return strings.Join(allParts, ":")
}

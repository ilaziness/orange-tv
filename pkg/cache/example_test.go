package cache_test

import (
	"context"
	"fmt"
	"time"

	"github.com/ilaziness/orange-tv/pkg/cache"
)

// ExampleMemoryCache 内存缓存基本使用示例
func ExampleMemoryCache() {
	// 配置内存缓存
	config := cache.MemoryCacheConfig{
		NumCounters: 1000,    // 追踪的键数量（建议为预期键数的 10 倍）
		MaxCost:     1 << 20, // 最大成本：1MB
		BufferItems: 64,      // 缓冲区大小
	}

	c, err := cache.NewMemoryCache(config)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	ctx := context.Background()

	// 设置缓存（TTL 5分钟）
	_ = c.Set(ctx, "user:1", "John Doe", 5*time.Minute)

	// 获取缓存
	value, err := c.Get(ctx, "user:1")
	if err != nil {
		// 处理缓存未命中或其他错误
		fmt.Println("cache miss or error:", err)
		return
	}

	fmt.Println(value)
	// Output: John Doe
}

// ExampleMultiCache 多级缓存使用示例
// 展示如何使用 L1（内存）+ L2（Redis）的多级缓存策略
func ExampleMultiCache() {
	// 创建 L1 内存缓存（快速访问层）
	l1Config := cache.MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1 << 20, // 1MB
		BufferItems: 64,
	}
	l1, err := cache.NewMemoryCache(l1Config)
	if err != nil {
		panic(err)
	}
	defer l1.Close()

	// 注意：实际生产环境中，L2 应该配置真实的 Redis
	// 这里为了示例简洁，仅演示 L1 缓存的使用
	//
	// 完整的多级缓存配置示例：
	// redisOpts := cache.RedisOptions{
	//     Host:     "localhost",
	//     Port:     6379,
	//     Password: "",
	//     DB:       0,
	// }
	// cacheConfig := cache.RedisCacheConfig{
	//     KeyPrefix:  "app:",
	//     DefaultTTL: 10 * time.Minute,
	// }
	// l2, _ := cache.NewRedisCache(cacheConfig, redisOpts)
	// multiCache := cache.NewMultiCache(l1, l2)

	ctx := context.Background()

	// 多级缓存会自动处理 L1/L2 协同：
	// 1. Get 时先查 L1，未命中再查 L2，并将结果回写到 L1
	// 2. Set 时同时写入 L1 和 L2
	// 3. Delete 时同时从 L1 和 L2 删除

	// 设置缓存
	_ = l1.Set(ctx, "product:1001", "Premium Widget", 10*time.Minute)

	// 获取缓存
	value, err := l1.Get(ctx, "product:1001")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(value)
	// Output: Premium Widget
}

// ExampleCache_keyGenerator 使用键生成器的缓存示例
func ExampleCache_keyGenerator() {
	config := cache.MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	}

	c, _ := cache.NewMemoryCache(config)
	defer c.Close()

	// 使用默认键生成器
	keyGen := cache.NewDefaultKeyGenerator()

	ctx := context.Background()

	// 生成结构化的缓存键
	key := keyGen.Generate("user", "123", "profile")
	// key = "user:123:profile"

	_ = c.Set(ctx, key, map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
	}, 5*time.Minute)

	value, _ := c.Get(ctx, key)
	fmt.Printf("Key: %s\n", key)
	fmt.Printf("Value: %v\n", value)
	// Output:
	// Key: user:123:profile
	// Value: map[email:alice@example.com name:Alice]
}

// ExampleCache_errorHandling 缓存错误处理示例
func ExampleCache_errorHandling() {
	config := cache.MemoryCacheConfig{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	}

	c, _ := cache.NewMemoryCache(config)
	defer c.Close()

	ctx := context.Background()

	// 尝试获取不存在的键
	_, err := c.Get(ctx, "nonexistent:key")
	if err != nil {
		// 检查是否为缓存未命中错误
		if cache.IsCacheMiss(err) {
			fmt.Println("Key not found in cache")
		} else {
			fmt.Println("Cache error:", err)
		}
	}

	// 尝试设置空键
	err = c.Set(ctx, "", "value", 5*time.Minute)
	if err != nil {
		fmt.Println("Invalid key error:", err)
	}
	// Output:
	// Key not found in cache
	// Invalid key error: cache: invalid key
}

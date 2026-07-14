package utils

import (
	"sync"

	"go.uber.org/zap"
)

var (
	// maxGoroutines 限制最大并发 Goroutine 数量
	maxGoroutines = 1000
	// semaphore 用于控制并发数
	semaphore = make(chan struct{}, maxGoroutines)
	// wg 用于等待所有 Goroutine 完成
	wg sync.WaitGroup
	// once 确保 logger 只初始化一次
	once sync.Once
	// log 全局 logger 实例
	log *zap.Logger
	// mu 保护 maxGoroutines 和 semaphore 的并发访问
	mu sync.RWMutex
)

// SetLogger 设置全局 logger
func SetLogger(logger *zap.Logger) {
	once.Do(func() {
		log = logger
	})
}

// SetMaxGoroutines 设置最大并发 Goroutine 数量
func SetMaxGoroutines(n int) {
	if n <= 0 {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	maxGoroutines = n
	newSemaphore := make(chan struct{}, maxGoroutines)
	// 迁移旧的 semaphore 中的信号
	for len(semaphore) > 0 {
		<-semaphore
		newSemaphore <- struct{}{}
	}
	semaphore = newSemaphore
}

// GetMaxGoroutines 获取当前最大并发 Goroutine 数量
func GetMaxGoroutines() int {
	mu.RLock()
	defer mu.RUnlock()
	return maxGoroutines
}

// GetActiveGoroutines 获取当前活跃的 Goroutine 数量
func GetActiveGoroutines() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(semaphore)
}

// Go 启动一个受控的 Goroutine
// 统一 Goroutine 启动入口，限制最大数量，添加 panic recovery
func Go(fn func()) {
	wg.Add(1)
	mu.RLock()
	semaphore <- struct{}{}
	mu.RUnlock()

	go func() {
		defer func() {
			mu.RLock()
			<-semaphore
			mu.RUnlock()
			wg.Done()
			if r := recover(); r != nil {
				if log != nil {
					log.Error("goroutine panic recovered",
						zap.Any("panic", r),
						zap.Stack("stack"))
				}
			}
		}()
		fn()
	}()
}

// Wait 等待所有通过 Go 启动的 Goroutine 完成
func Wait() {
	wg.Wait()
}

// GoWithRetry 启动一个带重试机制的 Goroutine
// maxRetries: 最大重试次数
// fn: 要执行的函数
func GoWithRetry(maxRetries int, fn func() error) {
	Go(func() {
		var err error
		for i := 0; i <= maxRetries; i++ {
			if err = fn(); err == nil {
				return
			}
			if i < maxRetries {
				if log != nil {
					log.Warn("goroutine retry",
						zap.Int("attempt", i+1),
						zap.Int("max_retries", maxRetries),
						zap.Error(err))
				}
			}
		}
		if log != nil {
			log.Error("goroutine failed after retries",
				zap.Int("max_retries", maxRetries),
				zap.Error(err))
		}
	})
}

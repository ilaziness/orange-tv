package utils

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetMaxGoroutines(t *testing.T) {
	original := GetMaxGoroutines()
	defer SetMaxGoroutines(original)

	SetMaxGoroutines(500)
	assert.Equal(t, 500, GetMaxGoroutines())

	SetMaxGoroutines(100)
	assert.Equal(t, 100, GetMaxGoroutines())
}

func TestSetMaxGoroutines_Invalid(t *testing.T) {
	original := GetMaxGoroutines()
	defer SetMaxGoroutines(original)

	SetMaxGoroutines(0)
	assert.Equal(t, original, GetMaxGoroutines())

	SetMaxGoroutines(-1)
	assert.Equal(t, original, GetMaxGoroutines())
}

func TestGetActiveGoroutines(t *testing.T) {
	SetMaxGoroutines(10)
	defer SetMaxGoroutines(1000)

	active := GetActiveGoroutines()
	assert.Equal(t, 0, active)

	// 启动一些 goroutine
	var counter int32
	for i := 0; i < 5; i++ {
		Go(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
	}

	time.Sleep(10 * time.Millisecond) // 等待 goroutine 启动
	active = GetActiveGoroutines()
	assert.Greater(t, active, 0)
	assert.LessOrEqual(t, active, 5)

	Wait()
	assert.Equal(t, int32(5), atomic.LoadInt32(&counter))
}

func TestGo_PanicRecovery(t *testing.T) {
	SetLogger(zap.NewNop()) // 使用 no-op logger 避免日志输出

	var counter int32
	for i := 0; i < 5; i++ {
		Go(func() {
			if i%2 == 0 {
				panic("test panic")
			}
			atomic.AddInt32(&counter, 1)
		})
	}

	Wait()
	// 即使有 panic，Wait 也应该正常返回
	assert.Equal(t, int32(2), atomic.LoadInt32(&counter))
}

func TestGo_ConcurrencyLimit(t *testing.T) {
	SetMaxGoroutines(5)
	defer SetMaxGoroutines(1000)

	var activeCount int32
	var maxActive int32
	var mu sync.Mutex

	// 启动超过限制的 goroutine
	for i := 0; i < 20; i++ {
		Go(func() {
			mu.Lock()
			activeCount++
			if activeCount > maxActive {
				maxActive = activeCount
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			activeCount--
			mu.Unlock()
		})
	}

	Wait()

	// 最大并发数不应该超过限制
	assert.LessOrEqual(t, maxActive, int32(5))
}

func TestWait(t *testing.T) {
	var counter int32
	for i := 0; i < 10; i++ {
		Go(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
	}

	Wait()
	assert.Equal(t, int32(10), atomic.LoadInt32(&counter))
}

func TestGoWithRetry_Success(t *testing.T) {
	SetLogger(zap.NewNop())

	var attempts int32
	GoWithRetry(3, func() error {
		atomic.AddInt32(&attempts, 1)
		if atomic.LoadInt32(&attempts) < 2 {
			return assert.AnError
		}
		return nil
	})

	Wait()
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestGoWithRetry_Failure(t *testing.T) {
	SetLogger(zap.NewNop())

	var attempts int32
	GoWithRetry(3, func() error {
		atomic.AddInt32(&attempts, 1)
		return assert.AnError
	})

	Wait()
	assert.Equal(t, int32(4), atomic.LoadInt32(&attempts)) // 初始尝试 + 3次重试
}

func TestSetLogger(t *testing.T) {
	logger := zap.NewNop()
	SetLogger(logger)
	// 不应该 panic
	assert.NotNil(t, logger)
}

package scheduler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/logger"
)

// mockLockProvider is a test mock for LockProvider.
type mockLockProvider struct {
	mu           sync.Mutex
	locked       bool
	acquireErr   error
	renewErr     error
	releaseErr   error
	renewCount   int64
	releaseCount int64
}

func (m *mockLockProvider) AcquireSchedulerLock(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.acquireErr != nil {
		return false, m.acquireErr
	}
	if m.locked {
		return false, nil
	}
	m.locked = true
	return true, nil
}

func (m *mockLockProvider) RenewSchedulerLock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.renewCount++
	if m.renewErr != nil {
		return m.renewErr
	}
	return nil
}

func (m *mockLockProvider) ReleaseSchedulerLock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releaseCount++
	m.locked = false
	if m.releaseErr != nil {
		return m.releaseErr
	}
	return nil
}

// isLocked checks if the mock lock is held (thread-safe).
func (m *mockLockProvider) isLocked() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locked
}

// getReleaseCount returns the ReleaseSchedulerLock call count (thread-safe).
func (m *mockLockProvider) getReleaseCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.releaseCount
}

// mockJob is a test implementation of the Job interface.
type mockJob struct {
	initErr   error
	initCount int64
}

func (j *mockJob) Init(ctx context.Context, sched *Scheduler) error {
	atomic.AddInt64(&j.initCount, 1)
	return j.initErr
}

// initLoggerForTest initializes the global logger to avoid nil panic in tests.
func initLoggerForTest(t *testing.T) {
	t.Helper()
	if logger.Log != nil {
		return
	}
	logInst, err := logger.New(logger.Config{Level: "debug", Output: "stdout"})
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	logger.Log = logInst
}

func TestScheduler_StartWithoutLock(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	// No LockProvider (single-instance mode)
	sched := newSchedulerWithLock(nil)
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if atomic.LoadInt64(&job.initCount) != 1 {
		t.Error("expected Job.Init to be called once")
	}
	if !sched.cronStarted.Load() {
		t.Error("expected cronStarted to be true")
	}

	// Stop must stop cron (this was the bug when lock=nil)
	if err := sched.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if sched.cronStarted.Load() {
		t.Error("expected cronStarted to be false after Stop")
	}
}

// TestScheduler_WithNoopLockProvider verifies that NoopLockProvider works as a single-instance
// lock: Acquire always succeeds, no heartbeat goroutine is started, Stop is clean.
func TestScheduler_WithNoopLockProvider(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	sched := newSchedulerWithLock(NoopLockProvider{})
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start with NoopLockProvider failed: %v", err)
	}
	if !sched.cronStarted.Load() {
		t.Error("expected cronStarted to be true with NoopLockProvider")
	}
	if !sched.lockOwned {
		t.Error("expected lockOwned to be true with NoopLockProvider")
	}
	if atomic.LoadInt64(&job.initCount) != 1 {
		t.Error("expected Job.Init to be called once")
	}
	// NoopLockProvider should not start heartbeat goroutine
	if sched.heartbeatCancel != nil {
		t.Error("expected no heartbeat goroutine with NoopLockProvider")
	}

	if err := sched.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if sched.cronStarted.Load() {
		t.Error("expected cronStarted to be false after Stop")
	}
}

func TestScheduler_StartWithLock_Acquired(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{}
	sched := newSchedulerWithLock(lock)
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !lock.isLocked() {
		t.Error("expected lock to be held after Start")
	}
	if !sched.lockOwned {
		t.Error("expected lockOwned to be true")
	}
	if atomic.LoadInt64(&job.initCount) != 1 {
		t.Error("expected Job.Init to be called once")
	}

	if err := sched.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if lock.isLocked() {
		t.Error("expected lock to be released after Stop")
	}
	if lock.getReleaseCount() != 1 {
		t.Error("expected ReleaseSchedulerLock to be called once")
	}
}

func TestScheduler_StartWithLock_AlreadyHeld(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{}
	// Pre-occupy the lock
	lock.locked = true

	sched := newSchedulerWithLock(lock)
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start should not return error when lock is held by another instance: %v", err)
	}
	// When lock is held by another instance, cron should not start, Job.Init should not be called
	if sched.cronStarted.Load() {
		t.Error("expected cronStarted to be false when lock is held by another instance")
	}
	if atomic.LoadInt64(&job.initCount) != 0 {
		t.Error("expected Job.Init NOT to be called when lock is held by another instance")
	}

	// Stop should not panic or error (cron not started)
	if err := sched.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	// Should not release a lock it does not own
	if lock.getReleaseCount() != 0 {
		t.Error("expected ReleaseSchedulerLock NOT to be called when lock was not owned")
	}
}

func TestScheduler_StartWithLock_AcquireError(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{acquireErr: errors.New("redis down")}
	sched := newSchedulerWithLock(lock)
	job := &mockJob{}
	sched.Register(job)

	err := sched.Start(ctx)
	if err == nil {
		t.Fatal("expected error when AcquireSchedulerLock fails")
	}
	if sched.cronStarted.Load() {
		t.Error("expected cronStarted to be false on acquire error")
	}
	if atomic.LoadInt64(&job.initCount) != 0 {
		t.Error("expected Job.Init NOT to be called on acquire error")
	}
}

func TestScheduler_StartJobInitFailure_ReleasesLock(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{}
	sched := newSchedulerWithLock(lock)
	job := &mockJob{initErr: errors.New("init failed")}
	sched.Register(job)

	if err := sched.Start(ctx); err == nil {
		t.Fatal("expected Start to return error when Job.Init fails")
	}
	// Lock should be released after Job.Init failure
	if lock.isLocked() {
		t.Error("expected lock to be released after Job.Init failure")
	}
	if sched.cronStarted.Load() {
		t.Error("expected cronStarted to be false after Job.Init failure")
	}
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	sched := newSchedulerWithLock(nil)
	// Stop without Start should be safe
	if err := sched.Stop(ctx); err != nil {
		t.Fatalf("Stop without Start should not error: %v", err)
	}
}

func TestScheduler_StopReleasesLock_AllowsOtherInstance(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{}

	// Instance 1 starts and acquires lock
	sched1 := newSchedulerWithLock(lock)
	if err := sched1.Start(ctx); err != nil {
		t.Fatalf("sched1 Start failed: %v", err)
	}
	if !lock.isLocked() {
		t.Fatal("expected lock to be held by sched1")
	}

	// Instance 2 tries to start, should skip due to lock held
	sched2 := newSchedulerWithLock(lock)
	if err := sched2.Start(ctx); err != nil {
		t.Fatalf("sched2 Start failed: %v", err)
	}
	if sched2.cronStarted.Load() {
		t.Error("sched2 should not start cron when lock is held by sched1")
	}

	// Instance 1 stops, releases lock
	if err := sched1.Stop(ctx); err != nil {
		t.Fatalf("sched1 Stop failed: %v", err)
	}
	if lock.isLocked() {
		t.Fatal("expected lock to be released after sched1 Stop")
	}

	// Instance 3 should now acquire lock and start
	sched3 := newSchedulerWithLock(lock)
	if err := sched3.Start(ctx); err != nil {
		t.Fatalf("sched3 Start failed: %v", err)
	}
	if !lock.isLocked() {
		t.Error("expected lock to be held by sched3")
	}
	if !sched3.cronStarted.Load() {
		t.Error("expected sched3 to start cron")
	}
	_ = sched3.Stop(ctx)
}

// TestScheduler_HeartbeatStopsOnRelease verifies heartbeat goroutine stops correctly on Stop.
// releaseLock's <-heartbeatDone blocks until heartbeat goroutine exits;
// if heartbeat doesn't stop, this test exposes the issue via Stop timeout.
func TestScheduler_HeartbeatStopsOnRelease(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{}
	sched := newSchedulerWithLock(lock)
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop internally cancels heartbeatCtx and waits on heartbeatDone
	// if heartbeat goroutine doesn't exit, Stop will hang
	done := make(chan error, 1)
	go func() {
		done <- sched.Stop(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop failed: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Stop timed out, heartbeat goroutine likely not stopped")
	}

	if lock.getReleaseCount() != 1 {
		t.Error("expected ReleaseSchedulerLock to be called once")
	}
}

// TestScheduler_HeartbeatStopsCronOnLockLoss verifies heartbeat stops cron on lock loss.
// Scenario: Renew returns ErrLockNotHeld (lock expired or preempted),
// heartbeat should stop cron and exit to avoid split-brain.
func TestScheduler_HeartbeatStopsCronOnLockLoss(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	// renewErr = ErrLockNotHeld simulates lock loss
	lock := &mockLockProvider{renewErr: ErrLockNotHeld}
	sched := newSchedulerWithLock(lock)
	sched.lockRenewInterval = 50 * time.Millisecond // short interval for fast test
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !sched.cronStarted.Load() {
		t.Fatal("expected cronStarted to be true after Start")
	}

	// Wait for heartbeat to detect lock loss and stop cron (50ms interval, max 2s)
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !sched.cronStarted.Load() {
				// cron stopped by heartbeat
				// verify Stop is still safe (idempotent, doesn't re-stop cron, only releases lock)
				if err := sched.Stop(ctx); err != nil {
					t.Fatalf("Stop after lock loss failed: %v", err)
				}
				return
			}
		case <-deadline:
			t.Fatal("cron was not stopped within 2s, heartbeat did not detect lock loss")
		}
	}
}

// TestScheduler_StopAfterHeartbeatStoppedCron verifies Stop idempotency after heartbeat stopped cron.
// After heartbeat stops cron due to lock loss, Stop should not re-stop cron but should release lock.
func TestScheduler_StopAfterHeartbeatStoppedCron(t *testing.T) {
	initLoggerForTest(t)
	ctx := context.Background()

	lock := &mockLockProvider{renewErr: ErrLockNotHeld}
	sched := newSchedulerWithLock(lock)
	sched.lockRenewInterval = 50 * time.Millisecond // short interval for fast test
	job := &mockJob{}
	sched.Register(job)

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for heartbeat to stop cron
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	waitForCronStop := func() {
		for {
			select {
			case <-ticker.C:
				if !sched.cronStarted.Load() {
					return
				}
			case <-deadline:
				t.Fatal("cron was not stopped within 2s")
			}
		}
	}
	waitForCronStop()

	// Stop should be safe, no panic, no deadlock
	done := make(chan error, 1)
	go func() {
		done <- sched.Stop(ctx)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Stop timed out after heartbeat stopped cron")
	}
}

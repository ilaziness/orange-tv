package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStdoutLogger(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"invalid level defaults to info", "nonsense"},
		{"empty level defaults to info", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewStdoutLogger(tt.level)
			require.NotNil(t, l)
			// The logger should be able to log without panicking.
			l.Info("test message")
			_ = l.Sync()
		})
	}
}

// TestNewStdoutLogger_LevelFiltering verifies that the stdout logger respects
// the configured log level. It calls the real internal code path
// (newStdoutLoggerWithSyncer) with an in-memory WriteSyncer so output can be
// inspected without capturing process-global stdout.
func TestNewStdoutLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		shouldDebug bool
		shouldInfo  bool
		shouldWarn  bool
	}{
		{"debug passes all", "debug", true, true, true},
		{"info filters debug", "info", false, true, true},
		{"warn filters debug+info", "warn", false, false, true},
		{"error filters below error", "error", false, false, false},
		{"invalid level defaults to info", "nonsense", false, true, true},
		{"empty level defaults to info", "", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf syncBuffer
			l := newStdoutLoggerWithSyncer(tt.level, &buf)

			l.Debug("debug-msg")
			l.Info("info-msg")
			l.Warn("warn-msg")
			_ = l.Sync()

			output := buf.String()
			require.Equal(t, tt.shouldDebug, strings.Contains(output, "debug-msg"),
				"debug-msg visibility mismatch")
			require.Equal(t, tt.shouldInfo, strings.Contains(output, "info-msg"),
				"info-msg visibility mismatch")
			require.Equal(t, tt.shouldWarn, strings.Contains(output, "warn-msg"),
				"warn-msg visibility mismatch")
		})
	}
}

// TestNewStdoutLogger_WritesToSyncer verifies the logger actually produces
// output through the provided WriteSyncer.
func TestNewStdoutLogger_WritesToSyncer(t *testing.T) {
	var buf syncBuffer
	l := newStdoutLoggerWithSyncer("info", &buf)
	l.Info("hello-from-stdout-logger")
	_ = l.Sync()

	require.Contains(t, buf.String(), "hello-from-stdout-logger")
}

// TestNew_FileOutputCreatesFile verifies that the main New logger does write
// to a file when configured with "file" output, contrasting with
// NewStdoutLogger which must not.
func TestNew_FileOutputCreatesFile(t *testing.T) {
	// Use a manual temp path instead of t.TempDir() because lumberjack keeps
	// the file handle open on Windows, which blocks TempDir's auto-cleanup.
	filename := filepath.Join(os.TempDir(), "orange-tv-test-app.log")
	t.Cleanup(func() { _ = os.Remove(filename) })

	logInst, err := New(Config{
		Level:    "info",
		Output:   "file",
		Filename: filename,
		MaxSize:  1,
	})
	require.NoError(t, err)
	logInst.Info("file-test-marker")
	_ = logInst.Sync()

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Contains(t, string(data), "file-test-marker")
}

// TestNew_StdoutOutputDoesNotCreateFile verifies that New with "stdout" output
// does not create a log file.
func TestNew_StdoutOutputDoesNotCreateFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "should-not-exist.log")

	logInst, err := New(Config{
		Level:    "info",
		Output:   "stdout",
		Filename: filename,
	})
	require.NoError(t, err)
	logInst.Info("stdout-only-marker")
	_ = logInst.Sync()

	_, statErr := os.Stat(filename)
	require.True(t, os.IsNotExist(statErr),
		"stdout output mode must not create log files")
}

// TestBuildEncoders_ReturnsDistinctEncoders ensures both encoders are non-nil.
func TestBuildEncoders_ReturnsDistinctEncoders(t *testing.T) {
	stdoutEnc, fileEnc := buildEncoders()
	require.NotNil(t, stdoutEnc)
	require.NotNil(t, fileEnc)
}

// syncBuffer is a thread-safe in-memory WriteSyncer for testing log output
// without capturing process-global stdout.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) Sync() error { return nil }

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

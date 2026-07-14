// Package logger provides structured logging with zap and log rotation support.
package logger

import (
	"os"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config represents logger configuration.
type Config struct {
	Level      string
	Format     string
	Output     string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// Logger wraps zap.Logger with additional functionality.
type Logger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

// Log is the package-level logger instance initialized by Init.
var Log *Logger

// Init initializes the package-level logger instance.
func Init(cfg Config) error {
	logInst, err := New(cfg)
	if err != nil {
		return err
	}
	Log = logInst
	return nil
}

// New creates a new logger instance.
func New(cfg Config) (*Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder based on format
	var encoder zapcore.Encoder
	if strings.ToLower(cfg.Format) == constant.LogFormatConsole {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create writers based on output
	var cores []zapcore.Core

	switch strings.ToLower(cfg.Output) {
	case constant.LogOutputFile:
		cores = append(cores, createFileCore(cfg, encoder, level))
	case constant.LogOutputBoth:
		cores = append(cores, createFileCore(cfg, encoder, level))
		cores = append(cores, createStdoutCore(encoder, level))
	default: // constant.LogOutputStdout
		cores = append(cores, createStdoutCore(encoder, level))
	}

	// Create core
	core := zapcore.NewTee(cores...)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		Logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// createFileCore creates a file-based core with log rotation.
func createFileCore(cfg Config, encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	writer := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level)
}

// createStdoutCore creates a stdout-based core.
func createStdoutCore(encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

// Sugar returns the sugared logger.
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		sugar:  l.Logger.With(fields...).Sugar(),
	}
}

// Named adds a sub-logger name.
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
		sugar:  l.sugar.Named(name),
	}
}

// WithTraceID adds a trace_id field to the logger.
func (l *Logger) WithTraceID(traceID string) *Logger {
	return l.With(zap.String("trace_id", traceID))
}

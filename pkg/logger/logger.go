package logger

import (
	"log/slog"
	"os"
)

// Logger wraps slog for structured logging with convenient methods
// This demonstrates Go's structured logging capabilities introduced in Go 1.21
type Logger struct {
	*slog.Logger
}

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// New creates a new logger with the specified level
// The logger uses JSON output for machine-readable logs
func New(level LogLevel) *Logger {
	var slogLevel slog.Level

	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create a handler with options
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	// Use TextHandler for human-readable output
	// For production, you might want JSONHandler for machine parsing
	handler := slog.NewTextHandler(os.Stdout, opts)

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Info logs an informational message with optional structured fields
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Debug logs a debug message with optional structured fields
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Warn logs a warning message with optional structured fields
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message with optional structured fields
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Fatal logs a fatal error and exits the program
func (l *Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// With returns a new logger with additional context fields
// This is useful for adding request IDs, client IDs, etc.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

package logger

import (
	"log"
	"log/slog"
	"os"
)

// Logger is a structured logger interface
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// New creates a new logger instance
func New() Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return &slogLogger{
		logger: slog.New(handler),
	}
}

// NewDevelopment creates a logger with debug level enabled
func NewDevelopment() Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return &slogLogger{
		logger: slog.New(handler),
	}
}

// slogLogger wraps slog.Logger to implement our Logger interface
type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{
		logger: l.logger.With(args...),
	}
}

// NoOpLogger is a logger that discards all output (useful for tests)
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, args ...any) {}
func (n *NoOpLogger) Info(msg string, args ...any)  {}
func (n *NoOpLogger) Warn(msg string, args ...any)  {}
func (n *NoOpLogger) Error(msg string, args ...any) {}
func (n *NoOpLogger) With(args ...any) Logger       { return n }

// FallbackLogger uses standard library log (for early initialization)
type FallbackLogger struct {
	log *log.Logger
}

func NewFallbackLogger() Logger {
	return &FallbackLogger{
		log: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (f *FallbackLogger) Debug(msg string, args ...any) {
	f.log.Printf("[DEBUG] %s %v", msg, args)
}

func (f *FallbackLogger) Info(msg string, args ...any) {
	f.log.Printf("[INFO] %s %v", msg, args)
}

func (f *FallbackLogger) Warn(msg string, args ...any) {
	f.log.Printf("[WARN] %s %v", msg, args)
}

func (f *FallbackLogger) Error(msg string, args ...any) {
	f.log.Printf("[ERROR] %s %v", msg, args)
}

func (f *FallbackLogger) With(args ...any) Logger {
	return f
}


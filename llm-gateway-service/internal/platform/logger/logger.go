package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Logger provides simple logging interface
type Logger struct {
	logger *log.Logger
}

// New creates a new logger
func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] %s", formatMessage(msg, args...))
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] %s", formatMessage(msg, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("[WARN] %s", formatMessage(msg, args...))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Printf("[DEBUG] %s", formatMessage(msg, args...))
}

func formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	var builder strings.Builder
	if len(args)%2 == 0 {
		builder.WriteString("====\n")
		builder.WriteString(msg)
		builder.WriteString("\n====\n")
		for i := 0; i < len(args); i += 2 {
			builder.WriteString(fmt.Sprintf("%v: %v\n", args[i], args[i+1]))
			builder.WriteString("---\n")
		}
		builder.WriteString("====")
		return builder.String()
	}
	builder.WriteString(fmt.Sprintf("%s %v", msg, args))
	return builder.String()
}

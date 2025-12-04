package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional utilities
type Logger struct {
	*logrus.Logger
}

var defaultLogger *Logger

func init() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		PrettyPrint:     false,
	})
	l.SetLevel(logrus.InfoLevel)
	defaultLogger = &Logger{l}
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() *Logger {
	return defaultLogger
}

// SetLogger sets the global logger instance
func SetLogger(l *Logger) {
	defaultLogger = l
}

// New creates a new logger instance
func New() *Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		PrettyPrint:     false,
	})
	l.SetLevel(logrus.InfoLevel)
	return &Logger{l}
}

// WithContext returns a logger entry with context values
func WithContext(ctx context.Context) *logrus.Entry {
	return defaultLogger.WithContext(ctx)
}

// WithFields returns a logger entry with fields
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return defaultLogger.WithFields(fields)
}

// WithField returns a logger entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return defaultLogger.WithField(key, value)
}

// Debug logs at debug level
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf logs at debug level with formatting
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info logs at info level
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof logs at info level with formatting
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn logs at warn level
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf logs at warn level with formatting
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error logs at error level
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf logs at error level with formatting
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal logs at fatal level and exits
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf logs at fatal level with formatting and exits
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Entry represents a log entry
type Entry = logrus.Entry

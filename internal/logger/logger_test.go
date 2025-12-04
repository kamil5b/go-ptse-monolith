package logger

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := New()
	require.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
	assert.Equal(t, logrus.InfoLevel, logger.Level)
}

func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	require.NotNil(t, logger)
	assert.Equal(t, logrus.InfoLevel, logger.Level)
}

func TestSetLogger(t *testing.T) {
	newLogger := New()
	newLogger.SetLevel(logrus.DebugLevel)

	SetLogger(newLogger)
	retrieved := GetDefaultLogger()

	assert.Equal(t, newLogger, retrieved)
	assert.Equal(t, logrus.DebugLevel, retrieved.Level)
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	entry := WithContext(ctx)

	require.NotNil(t, entry)
	assert.Equal(t, ctx, entry.Context)
}

func TestWithFields(t *testing.T) {
	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "login",
	}

	entry := WithFields(fields)
	require.NotNil(t, entry)

	// Check if fields are set
	for key, value := range fields {
		assert.Equal(t, value, entry.Data[key])
	}
}

func TestWithField(t *testing.T) {
	entry := WithField("request_id", "req-123")
	require.NotNil(t, entry)
	assert.Equal(t, "req-123", entry.Data["request_id"])
}

func TestDebug(t *testing.T) {
	logger := New()
	SetLogger(logger)

	// Should not panic
	Debug("debug message")
	Debugf("debug message: %s", "formatted")
}

func TestInfo(t *testing.T) {
	logger := New()
	SetLogger(logger)

	// Should not panic
	Info("info message")
	Infof("info message: %s", "formatted")
}

func TestWarn(t *testing.T) {
	logger := New()
	SetLogger(logger)

	// Should not panic
	Warn("warn message")
	Warnf("warn message: %s", "formatted")
}

func TestError(t *testing.T) {
	logger := New()
	SetLogger(logger)

	// Should not panic
	Error("error message")
	Errorf("error message: %s", "formatted")
}

func TestLoggerLevel(t *testing.T) {
	logger := New()
	logger.SetLevel(logrus.DebugLevel)

	assert.Equal(t, logrus.DebugLevel, logger.Level)

	logger.SetLevel(logrus.ErrorLevel)
	assert.Equal(t, logrus.ErrorLevel, logger.Level)
}

func TestJSONFormatter(t *testing.T) {
	logger := New()
	formatter, ok := logger.Formatter.(*logrus.JSONFormatter)

	assert.True(t, ok)
	assert.Equal(t, "2006-01-02T15:04:05Z07:00", formatter.TimestampFormat)
	assert.False(t, formatter.PrettyPrint)
}

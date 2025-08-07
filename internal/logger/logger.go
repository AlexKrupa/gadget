package logger

import (
	"fmt"
	"sync"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelSuccess
	LogLevelError
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelSuccess:
		return "SUCCESS"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
}

type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Success(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// Renderer interface defines how log entries are displayed
type Renderer interface {
	Render(entry LogEntry)
}

type GlobalLogger struct {
	renderer Renderer
	mu       sync.RWMutex
}

var (
	globalLogger = &GlobalLogger{}
)

func GetLogger() *GlobalLogger {
	return globalLogger
}

func SetRenderer(r Renderer) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.renderer = r
}

func (l *GlobalLogger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

func (l *GlobalLogger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

func (l *GlobalLogger) Success(format string, args ...interface{}) {
	l.log(LogLevelSuccess, format, args...)
}

func (l *GlobalLogger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

func (l *GlobalLogger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.RLock()
	renderer := l.renderer
	l.mu.RUnlock()

	if renderer == nil {
		return // No renderer set, ignore log
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
	}

	renderer.Render(entry)
}

// Convenience functions that use the global logger
func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

func Success(format string, args ...interface{}) {
	globalLogger.Success(format, args...)
}

func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

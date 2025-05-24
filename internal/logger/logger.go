package logger

import "log"

// Logger interface for debug logging
type Logger interface {
	Debug(msg string, args ...interface{})
}

// DebugLogger implements Logger with configurable debug mode
type DebugLogger struct {
	enabled bool
}

// NewDebugLogger creates a new DebugLogger
func NewDebugLogger(enabled bool) *DebugLogger {
	return &DebugLogger{enabled: enabled}
}

// Debug logs a message if debug mode is enabled
func (d *DebugLogger) Debug(msg string, args ...interface{}) {
	if d.enabled {
		log.Printf(msg, args...)
	}
}

// NoOpLogger implements Logger but does nothing
type NoOpLogger struct{}

// Debug does nothing
func (n *NoOpLogger) Debug(msg string, args ...interface{}) {
	// Do nothing
}

// NewNoOpLogger creates a new NoOpLogger
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

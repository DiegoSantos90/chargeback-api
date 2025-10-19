package service

import (
	"context"
	"errors"
	"fmt"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	// LogLevelDebug represents debug level logs for detailed troubleshooting
	LogLevelDebug LogLevel = iota

	// LogLevelInfo represents informational logs for general application flow
	LogLevelInfo

	// LogLevelWarn represents warning logs for potentially harmful situations
	LogLevelWarn

	// LogLevelError represents error logs for error events that might still allow the application to continue
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the log level is valid
func (l LogLevel) IsValid() bool {
	return l >= LogLevelDebug && l <= LogLevelError
}

// LogEntry represents a single log entry with structured data
type LogEntry struct {
	// Level is the severity level of the log entry
	Level LogLevel

	// Message is the main log message
	Message string

	// Fields contains structured key-value pairs for additional context
	Fields map[string]interface{}
}

// Validate validates the log entry for required fields and valid values
func (e LogEntry) Validate() error {
	if e.Message == "" {
		return errors.New("log message cannot be empty")
	}

	if !e.Level.IsValid() {
		return fmt.Errorf("invalid log level: %v", e.Level)
	}

	return nil
}

// WithField creates a new LogEntry with an additional field
// This method creates a copy to maintain immutability
func (e LogEntry) WithField(key string, value interface{}) LogEntry {
	newFields := make(map[string]interface{})

	// Copy existing fields
	for k, v := range e.Fields {
		newFields[k] = v
	}

	// Add new field
	newFields[key] = value

	return LogEntry{
		Level:   e.Level,
		Message: e.Message,
		Fields:  newFields,
	}
}

// WithFields creates a new LogEntry with additional fields
// This method creates a copy to maintain immutability
func (e LogEntry) WithFields(fields map[string]interface{}) LogEntry {
	newFields := make(map[string]interface{})

	// Copy existing fields
	for k, v := range e.Fields {
		newFields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newFields[k] = v
	}

	return LogEntry{
		Level:   e.Level,
		Message: e.Message,
		Fields:  newFields,
	}
}

// Logger defines the contract for logging operations in the domain layer
// This interface abstracts the logging implementation details from the business logic
type Logger interface {
	// Log writes a structured log entry
	Log(ctx context.Context, entry LogEntry) error

	// Debug logs a debug message with optional structured fields
	Debug(ctx context.Context, message string, fields ...map[string]interface{}) error

	// Info logs an informational message with optional structured fields
	Info(ctx context.Context, message string, fields ...map[string]interface{}) error

	// Warn logs a warning message with optional structured fields
	Warn(ctx context.Context, message string, fields ...map[string]interface{}) error

	// Error logs an error message with optional structured fields
	Error(ctx context.Context, message string, fields ...map[string]interface{}) error

	// WithContext returns a logger instance with additional context
	// This can be used to add request-scoped fields like correlation IDs
	WithContext(ctx context.Context) Logger
}

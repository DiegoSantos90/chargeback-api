package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/service"
)

// LogFormat represents the output format for logs
type LogFormat int

const (
	// FormatJSON outputs logs in JSON format for structured logging
	FormatJSON LogFormat = iota

	// FormatText outputs logs in human-readable text format
	FormatText
)

// String returns the string representation of the log format
func (f LogFormat) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatText:
		return "text"
	default:
		return "unknown"
	}
}

// IsValid checks if the log format is valid
func (f LogFormat) IsValid() bool {
	return f >= FormatJSON && f <= FormatText
}

// LoggerConfig holds the configuration for the structured logger
type LoggerConfig struct {
	// Level is the minimum log level to output
	Level service.LogLevel

	// Format determines the output format (JSON or text)
	Format LogFormat

	// ServiceName is the name of the service for structured logging
	ServiceName string

	// Version is the version of the service
	Version string
}

// Validate validates the logger configuration
func (c LoggerConfig) Validate() error {
	if !c.Level.IsValid() {
		return fmt.Errorf("invalid log level: %v", c.Level)
	}

	if !c.Format.IsValid() {
		return fmt.Errorf("invalid log format: %v", c.Format)
	}

	return nil
}

// StructuredLogger implements the domain Logger interface using Go's slog package
type StructuredLogger struct {
	logger      *slog.Logger
	level       service.LogLevel
	config      LoggerConfig
	contextKeys []string
}

// NewStructuredLogger creates a new structured logger with the given configuration
func NewStructuredLogger(config LoggerConfig, writer io.Writer) (*StructuredLogger, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	// If no writer provided, use stdout
	if writer == nil {
		writer = os.Stdout
	}

	// Convert domain log level to slog level
	slogLevel := convertLogLevel(config.Level)

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	// Create appropriate handler based on format
	var handler slog.Handler
	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, opts)
	case FormatText:
		handler = slog.NewTextHandler(writer, opts)
	default:
		return nil, fmt.Errorf("unsupported log format: %v", config.Format)
	}

	// Create base logger
	logger := slog.New(handler)

	// Add service metadata if provided
	if config.ServiceName != "" {
		logger = logger.With("service", config.ServiceName)
	}

	if config.Version != "" {
		logger = logger.With("version", config.Version)
	}

	return &StructuredLogger{
		logger: logger,
		level:  config.Level,
		config: config,
	}, nil
}

// convertLogLevel converts domain log level to slog level
func convertLogLevel(level service.LogLevel) slog.Level {
	switch level {
	case service.LogLevelDebug:
		return slog.LevelDebug
	case service.LogLevelInfo:
		return slog.LevelInfo
	case service.LogLevelWarn:
		return slog.LevelWarn
	case service.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// convertDomainLevelToSlog converts domain log level to slog level for individual calls
func convertDomainLevelToSlog(level service.LogLevel) slog.Level {
	switch level {
	case service.LogLevelDebug:
		return slog.LevelDebug
	case service.LogLevelInfo:
		return slog.LevelInfo
	case service.LogLevelWarn:
		return slog.LevelWarn
	case service.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Log writes a structured log entry
func (s *StructuredLogger) Log(ctx context.Context, entry service.LogEntry) error {
	// Validate the log entry
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid log entry: %w", err)
	}

	// Check if the level is enabled
	slogLevel := convertDomainLevelToSlog(entry.Level)
	if !s.logger.Enabled(ctx, slogLevel) {
		return nil
	}

	// Convert fields to slog attributes
	var attrs []slog.Attr
	for key, value := range entry.Fields {
		attrs = append(attrs, slog.Any(key, value))
	}

	// Log with the appropriate level
	s.logger.LogAttrs(ctx, slogLevel, entry.Message, attrs...)

	return nil
}

// Debug logs a debug message with optional structured fields
func (s *StructuredLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := service.LogEntry{
		Level:   service.LogLevelDebug,
		Message: message,
		Fields:  make(map[string]interface{}),
	}

	// Merge all field maps
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	return s.Log(ctx, entry)
}

// Info logs an informational message with optional structured fields
func (s *StructuredLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := service.LogEntry{
		Level:   service.LogLevelInfo,
		Message: message,
		Fields:  make(map[string]interface{}),
	}

	// Merge all field maps
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	return s.Log(ctx, entry)
}

// Warn logs a warning message with optional structured fields
func (s *StructuredLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := service.LogEntry{
		Level:   service.LogLevelWarn,
		Message: message,
		Fields:  make(map[string]interface{}),
	}

	// Merge all field maps
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	return s.Log(ctx, entry)
}

// Error logs an error message with optional structured fields
func (s *StructuredLogger) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := service.LogEntry{
		Level:   service.LogLevelError,
		Message: message,
		Fields:  make(map[string]interface{}),
	}

	// Merge all field maps
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	return s.Log(ctx, entry)
}

// WithContext returns a logger instance with additional context
// This can be used to add request-scoped fields like correlation IDs
func (s *StructuredLogger) WithContext(ctx context.Context) service.Logger {
	// Extract correlation ID from context if available
	contextLogger := s.logger

	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			contextLogger = contextLogger.With("correlation_id", id)
		}
	}

	// Extract request ID from context if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			contextLogger = contextLogger.With("request_id", id)
		}
	}

	// Extract user ID from context if available
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			contextLogger = contextLogger.With("user_id", id)
		}
	}

	return &StructuredLogger{
		logger: contextLogger,
		level:  s.level,
		config: s.config,
	}
}

// NewDefaultLogger creates a logger with sensible defaults for development
func NewDefaultLogger() (*StructuredLogger, error) {
	config := LoggerConfig{
		Level:       service.LogLevelInfo,
		Format:      FormatJSON,
		ServiceName: "chargeback-api",
	}

	return NewStructuredLogger(config, os.Stdout)
}

// NewDevelopmentLogger creates a logger optimized for development
func NewDevelopmentLogger() (*StructuredLogger, error) {
	config := LoggerConfig{
		Level:       service.LogLevelDebug,
		Format:      FormatText,
		ServiceName: "chargeback-api",
	}

	return NewStructuredLogger(config, os.Stdout)
}

// NewProductionLogger creates a logger optimized for production
func NewProductionLogger() (*StructuredLogger, error) {
	config := LoggerConfig{
		Level:       service.LogLevelInfo,
		Format:      FormatJSON,
		ServiceName: "chargeback-api",
	}

	return NewStructuredLogger(config, os.Stdout)
}

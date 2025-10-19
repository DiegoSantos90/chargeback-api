package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/service"
)

// TestStructuredLogger_New tests the creation of a new structured logger
func TestStructuredLogger_New(t *testing.T) {
	tests := []struct {
		name   string
		config LoggerConfig
		want   bool // whether creation should succeed
	}{
		{
			name: "Valid config with JSON format should succeed",
			config: LoggerConfig{
				Level:  service.LogLevelInfo,
				Format: FormatJSON,
			},
			want: true,
		},
		{
			name: "Valid config with text format should succeed",
			config: LoggerConfig{
				Level:  service.LogLevelDebug,
				Format: FormatText,
			},
			want: true,
		},
		{
			name: "Invalid log level should fail",
			config: LoggerConfig{
				Level:  service.LogLevel(999),
				Format: FormatJSON,
			},
			want: false,
		},
		{
			name: "Invalid format should fail",
			config: LoggerConfig{
				Level:  service.LogLevelInfo,
				Format: LogFormat(999),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := NewStructuredLogger(tt.config, &buf)

			if tt.want && err != nil {
				t.Errorf("NewStructuredLogger() error = %v, want success", err)
			}

			if !tt.want && err == nil {
				t.Error("NewStructuredLogger() should have failed but succeeded")
			}

			if tt.want && logger == nil {
				t.Error("NewStructuredLogger() returned nil logger")
			}
		})
	}
}

// TestStructuredLogger_Log tests the basic Log method
func TestStructuredLogger_Log(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelDebug,
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()
	entry := service.LogEntry{
		Level:   service.LogLevelInfo,
		Message: "Test message",
		Fields: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	err = logger.Log(ctx, entry)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected log output, got empty string")
	}

	// Verify JSON structure
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	// Check required fields
	if logData["level"] != "INFO" {
		t.Errorf("Expected level 'INFO', got %v", logData["level"])
	}

	if logData["msg"] != "Test message" {
		t.Errorf("Expected message 'Test message', got %v", logData["msg"])
	}

	if logData["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got %v", logData["key1"])
	}

	if logData["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected key2 42, got %v", logData["key2"])
	}
}

// TestStructuredLogger_LogLevelFiltering tests that logs below the configured level are filtered
func TestStructuredLogger_LogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelWarn, // Only WARN and ERROR should be logged
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()

	// This should be filtered out (DEBUG < WARN)
	debugEntry := service.LogEntry{
		Level:   service.LogLevelDebug,
		Message: "Debug message",
		Fields:  map[string]interface{}{},
	}

	err = logger.Log(ctx, debugEntry)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	// Buffer should be empty since DEBUG is below WARN
	if buf.String() != "" {
		t.Error("Expected empty output for filtered debug log")
	}

	// This should be logged (WARN >= WARN)
	warnEntry := service.LogEntry{
		Level:   service.LogLevelWarn,
		Message: "Warning message",
		Fields:  map[string]interface{}{},
	}

	err = logger.Log(ctx, warnEntry)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected log output for warn level")
	}

	if !strings.Contains(output, "Warning message") {
		t.Error("Expected warning message in output")
	}
}

// TestStructuredLogger_Debug tests the Debug convenience method
func TestStructuredLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelDebug,
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()
	fields := map[string]interface{}{
		"operation": "test_debug",
		"duration":  123,
	}

	err = logger.Debug(ctx, "Debug test message", fields)
	if err != nil {
		t.Errorf("Debug() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected debug output")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	if logData["level"] != "DEBUG" {
		t.Errorf("Expected level 'DEBUG', got %v", logData["level"])
	}

	if logData["msg"] != "Debug test message" {
		t.Errorf("Expected message 'Debug test message', got %v", logData["msg"])
	}

	if logData["operation"] != "test_debug" {
		t.Errorf("Expected operation 'test_debug', got %v", logData["operation"])
	}
}

// TestStructuredLogger_Info tests the Info convenience method
func TestStructuredLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelInfo,
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()

	err = logger.Info(ctx, "Info test message")
	if err != nil {
		t.Errorf("Info() error = %v", err)
	}

	output := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	if logData["level"] != "INFO" {
		t.Errorf("Expected level 'INFO', got %v", logData["level"])
	}
}

// TestStructuredLogger_Warn tests the Warn convenience method
func TestStructuredLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelWarn,
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()

	err = logger.Warn(ctx, "Warning test message")
	if err != nil {
		t.Errorf("Warn() error = %v", err)
	}

	output := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	if logData["level"] != "WARN" {
		t.Errorf("Expected level 'WARN', got %v", logData["level"])
	}
}

// TestStructuredLogger_Error tests the Error convenience method
func TestStructuredLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelError,
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()

	err = logger.Error(ctx, "Error test message")
	if err != nil {
		t.Errorf("Error() error = %v", err)
	}

	output := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	if logData["level"] != "ERROR" {
		t.Errorf("Expected level 'ERROR', got %v", logData["level"])
	}
}

// TestStructuredLogger_TextFormat tests text format output
func TestStructuredLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelInfo,
		Format: FormatText,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	ctx := context.Background()
	entry := service.LogEntry{
		Level:   service.LogLevelInfo,
		Message: "Text format test",
		Fields: map[string]interface{}{
			"key1": "value1",
		},
	}

	err = logger.Log(ctx, entry)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected text output")
	}

	// Text format should contain the message and be human readable
	if !strings.Contains(output, "Text format test") {
		t.Error("Expected message in text output")
	}

	if !strings.Contains(output, "INFO") {
		t.Error("Expected level in text output")
	}
}

// TestStructuredLogger_WithContext tests context handling
func TestStructuredLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:  service.LogLevelInfo,
		Format: FormatJSON,
	}

	logger, err := NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error = %v", err)
	}

	// Create context with correlation ID
	ctx := context.WithValue(context.Background(), "correlation_id", "test-123")

	contextLogger := logger.WithContext(ctx)
	if contextLogger == nil {
		t.Error("WithContext() should not return nil")
	}

	err = contextLogger.Info(ctx, "Context test message")
	if err != nil {
		t.Errorf("Info() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output with context")
	}
}

// TestLoggerConfig_Validate tests configuration validation
func TestLoggerConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    LoggerConfig
		expectErr bool
	}{
		{
			name: "Valid JSON config",
			config: LoggerConfig{
				Level:  service.LogLevelInfo,
				Format: FormatJSON,
			},
			expectErr: false,
		},
		{
			name: "Valid text config",
			config: LoggerConfig{
				Level:  service.LogLevelDebug,
				Format: FormatText,
			},
			expectErr: false,
		},
		{
			name: "Invalid log level",
			config: LoggerConfig{
				Level:  service.LogLevel(999),
				Format: FormatJSON,
			},
			expectErr: true,
		},
		{
			name: "Invalid format",
			config: LoggerConfig{
				Level:  service.LogLevelInfo,
				Format: LogFormat(999),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("LoggerConfig.Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// TestLogFormat_String tests log format string representation
func TestLogFormat_String(t *testing.T) {
	tests := []struct {
		name     string
		format   LogFormat
		expected string
	}{
		{
			name:     "JSON format",
			format:   FormatJSON,
			expected: "json",
		},
		{
			name:     "Text format",
			format:   FormatText,
			expected: "text",
		},
		{
			name:     "Invalid format",
			format:   LogFormat(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.format.String()
			if result != tt.expected {
				t.Errorf("LogFormat.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

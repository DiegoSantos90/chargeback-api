package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/service"
	"github.com/DiegoSantos90/chargeback-api/internal/infra/logging"
)

// TestLoggingIntegration_EndToEnd tests the complete logging flow
func TestLoggingIntegration_EndToEnd(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	config := logging.LoggerConfig{
		Level:       service.LogLevelInfo,
		Format:      logging.FormatJSON,
		ServiceName: "chargeback-api-test",
		Version:     "1.0.0",
	}

	logger, err := logging.NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()

	// Act - Log different levels
	logger.Info(ctx, "Application starting", map[string]interface{}{
		"port":    "8080",
		"version": "1.0.0",
	})

	logger.Warn(ctx, "Configuration warning", map[string]interface{}{
		"missing_config": "optional_setting",
	})

	logger.Error(ctx, "Database connection failed", map[string]interface{}{
		"error":    "connection timeout",
		"attempts": 3,
	})

	// Assert - Parse and verify JSON output
	output := buf.String()
	if output == "" {
		t.Fatal("Expected logging output, got empty string")
	}

	// Split by lines to check each log entry
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	validEntries := 0

	for _, line := range lines {
		if len(line) == 0 {
			continue // Skip empty lines
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal(line, &logEntry); err != nil {
			t.Errorf("Failed to parse JSON log entry: %v, line: %s", err, string(line))
			continue
		}

		validEntries++

		// Verify required fields
		if logEntry["level"] == nil {
			t.Error("Log entry missing 'level' field")
		}

		if logEntry["msg"] == nil {
			t.Error("Log entry missing 'msg' field")
		}

		if logEntry["time"] == nil {
			t.Error("Log entry missing 'time' field")
		}

		// Verify service metadata
		if logEntry["service"] != "chargeback-api-test" {
			t.Errorf("Expected service 'chargeback-api-test', got %v", logEntry["service"])
		}

		if logEntry["version"] == nil {
			t.Error("Expected version field to be present")
		} else if logEntry["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %v (type: %T)", logEntry["version"], logEntry["version"])
		}
	}

	// Should have 3 valid log entries (info, warn, error)
	if validEntries != 3 {
		t.Errorf("Expected 3 valid log entries, got %d", validEntries)
	}
}

// TestLoggingIntegration_ContextualLogging tests context-aware logging
func TestLoggingIntegration_ContextualLogging(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	config := logging.LoggerConfig{
		Level:  service.LogLevelInfo,
		Format: logging.FormatJSON,
	}

	logger, err := logging.NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create context with request metadata
	ctx := context.WithValue(context.Background(), "correlation_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	contextLogger := logger.WithContext(ctx)

	// Act
	contextLogger.Info(ctx, "User action performed", map[string]interface{}{
		"action": "create_chargeback",
	})

	// Assert
	output := buf.String()
	if output == "" {
		t.Fatal("Expected logging output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log entry: %v", err)
	}

	// Verify context fields are included
	if logEntry["correlation_id"] != "req-123" {
		t.Errorf("Expected correlation_id 'req-123', got %v", logEntry["correlation_id"])
	}

	if logEntry["user_id"] != "user-456" {
		t.Errorf("Expected user_id 'user-456', got %v", logEntry["user_id"])
	}

	if logEntry["action"] != "create_chargeback" {
		t.Errorf("Expected action 'create_chargeback', got %v", logEntry["action"])
	}
}

// TestLoggingIntegration_LogLevelFiltering tests that logs are filtered by level
func TestLoggingIntegration_LogLevelFiltering(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	config := logging.LoggerConfig{
		Level:  service.LogLevelWarn, // Only WARN and ERROR should pass
		Format: logging.FormatJSON,
	}

	logger, err := logging.NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()

	// Act - Log at different levels
	logger.Debug(ctx, "Debug message")  // Should be filtered out
	logger.Info(ctx, "Info message")    // Should be filtered out
	logger.Warn(ctx, "Warning message") // Should pass
	logger.Error(ctx, "Error message")  // Should pass

	// Assert
	output := buf.String()
	if output == "" {
		t.Fatal("Expected some logging output")
	}

	// Count log entries
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	validEntries := 0

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		validEntries++

		var logEntry map[string]interface{}
		if err := json.Unmarshal(line, &logEntry); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
			continue
		}

		level := logEntry["level"]
		if level != "WARN" && level != "ERROR" {
			t.Errorf("Unexpected log level %v should have been filtered", level)
		}
	}

	// Should only have 2 entries (WARN and ERROR)
	if validEntries != 2 {
		t.Errorf("Expected 2 log entries (WARN, ERROR), got %d", validEntries)
	}
}

// TestLoggingIntegration_TextFormat tests text format output
func TestLoggingIntegration_TextFormat(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	config := logging.LoggerConfig{
		Level:  service.LogLevelInfo,
		Format: logging.FormatText,
	}

	logger, err := logging.NewStructuredLogger(config, &buf)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()

	// Act
	logger.Info(ctx, "Test message in text format", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	// Assert
	output := buf.String()
	if output == "" {
		t.Fatal("Expected logging output")
	}

	// Text format should be human-readable, not JSON
	if bytes.Contains(buf.Bytes(), []byte("{")) {
		t.Error("Text format output should not contain JSON braces")
	}

	// Should contain the message and level
	if !bytes.Contains(buf.Bytes(), []byte("Test message in text format")) {
		t.Error("Output should contain the log message")
	}

	if !bytes.Contains(buf.Bytes(), []byte("INFO")) {
		t.Error("Output should contain the log level")
	}
}

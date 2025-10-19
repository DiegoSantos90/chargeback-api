package service

import (
	"context"
	"testing"
)

// TestLogLevel_String tests the string representation of log levels
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{
			name:     "Debug level should return debug",
			level:    LogLevelDebug,
			expected: "DEBUG",
		},
		{
			name:     "Info level should return info",
			level:    LogLevelInfo,
			expected: "INFO",
		},
		{
			name:     "Warn level should return warn",
			level:    LogLevelWarn,
			expected: "WARN",
		},
		{
			name:     "Error level should return error",
			level:    LogLevelError,
			expected: "ERROR",
		},
		{
			name:     "Invalid level should return unknown",
			level:    LogLevel(999),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLogLevel_IsValid tests log level validation
func TestLogLevel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected bool
	}{
		{
			name:     "Debug level is valid",
			level:    LogLevelDebug,
			expected: true,
		},
		{
			name:     "Info level is valid",
			level:    LogLevelInfo,
			expected: true,
		},
		{
			name:     "Warn level is valid",
			level:    LogLevelWarn,
			expected: true,
		},
		{
			name:     "Error level is valid",
			level:    LogLevelError,
			expected: true,
		},
		{
			name:     "Invalid level is not valid",
			level:    LogLevel(999),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.IsValid()
			if result != tt.expected {
				t.Errorf("LogLevel.IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLogEntry_Validate tests log entry validation
func TestLogEntry_Validate(t *testing.T) {
	tests := []struct {
		name      string
		entry     LogEntry
		expectErr bool
	}{
		{
			name: "Valid log entry should not return error",
			entry: LogEntry{
				Level:   LogLevelInfo,
				Message: "Test message",
				Fields:  map[string]interface{}{"key": "value"},
			},
			expectErr: false,
		},
		{
			name: "Empty message should return error",
			entry: LogEntry{
				Level:   LogLevelInfo,
				Message: "",
				Fields:  map[string]interface{}{"key": "value"},
			},
			expectErr: true,
		},
		{
			name: "Invalid level should return error",
			entry: LogEntry{
				Level:   LogLevel(999),
				Message: "Test message",
				Fields:  map[string]interface{}{"key": "value"},
			},
			expectErr: true,
		},
		{
			name: "Nil fields map is allowed",
			entry: LogEntry{
				Level:   LogLevelInfo,
				Message: "Test message",
				Fields:  nil,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("LogEntry.Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// TestLogEntry_WithField tests adding fields to log entry
func TestLogEntry_WithField(t *testing.T) {
	entry := LogEntry{
		Level:   LogLevelInfo,
		Message: "Test message",
		Fields:  make(map[string]interface{}),
	}

	newEntry := entry.WithField("key1", "value1")
	
	// Original entry should be unchanged
	if len(entry.Fields) != 0 {
		t.Error("Original entry should not be modified")
	}
	
	// New entry should have the field
	if len(newEntry.Fields) != 1 {
		t.Errorf("New entry should have 1 field, got %d", len(newEntry.Fields))
	}
	
	if newEntry.Fields["key1"] != "value1" {
		t.Errorf("Field value should be 'value1', got %v", newEntry.Fields["key1"])
	}
}

// TestLogEntry_WithFields tests adding multiple fields to log entry
func TestLogEntry_WithFields(t *testing.T) {
	entry := LogEntry{
		Level:   LogLevelInfo,
		Message: "Test message",
		Fields:  make(map[string]interface{}),
	}

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	newEntry := entry.WithFields(fields)
	
	// Original entry should be unchanged
	if len(entry.Fields) != 0 {
		t.Error("Original entry should not be modified")
	}
	
	// New entry should have all fields
	if len(newEntry.Fields) != 3 {
		t.Errorf("New entry should have 3 fields, got %d", len(newEntry.Fields))
	}
	
	for key, expectedValue := range fields {
		if newEntry.Fields[key] != expectedValue {
			t.Errorf("Field %s should be %v, got %v", key, expectedValue, newEntry.Fields[key])
		}
	}
}

// MockLogger for testing purposes
type MockLogger struct {
	entries []LogEntry
}

func (m *MockLogger) Log(ctx context.Context, entry LogEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := LogEntry{
		Level:   LogLevelDebug,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
	
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}
	
	return m.Log(ctx, entry)
}

func (m *MockLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := LogEntry{
		Level:   LogLevelInfo,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
	
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}
	
	return m.Log(ctx, entry)
}

func (m *MockLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := LogEntry{
		Level:   LogLevelWarn,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
	
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}
	
	return m.Log(ctx, entry)
}

func (m *MockLogger) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	entry := LogEntry{
		Level:   LogLevelError,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
	
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}
	
	return m.Log(ctx, entry)
}

func (m *MockLogger) WithContext(ctx context.Context) Logger {
	return m
}

// TestLogger_Debug tests the Debug method
func TestLogger_Debug(t *testing.T) {
	mockLogger := &MockLogger{}
	ctx := context.Background()
	
	err := mockLogger.Debug(ctx, "Debug message")
	
	if err != nil {
		t.Errorf("Debug() should not return error, got %v", err)
	}
	
	if len(mockLogger.entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockLogger.entries))
	}
	
	entry := mockLogger.entries[0]
	if entry.Level != LogLevelDebug {
		t.Errorf("Expected DEBUG level, got %v", entry.Level)
	}
	
	if entry.Message != "Debug message" {
		t.Errorf("Expected 'Debug message', got %v", entry.Message)
	}
}

// TestLogger_Info tests the Info method
func TestLogger_Info(t *testing.T) {
	mockLogger := &MockLogger{}
	ctx := context.Background()
	
	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "login",
	}
	
	err := mockLogger.Info(ctx, "User logged in", fields)
	
	if err != nil {
		t.Errorf("Info() should not return error, got %v", err)
	}
	
	if len(mockLogger.entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockLogger.entries))
	}
	
	entry := mockLogger.entries[0]
	if entry.Level != LogLevelInfo {
		t.Errorf("Expected INFO level, got %v", entry.Level)
	}
	
	if entry.Message != "User logged in" {
		t.Errorf("Expected 'User logged in', got %v", entry.Message)
	}
	
	if entry.Fields["user_id"] != "123" {
		t.Errorf("Expected user_id to be '123', got %v", entry.Fields["user_id"])
	}
	
	if entry.Fields["action"] != "login" {
		t.Errorf("Expected action to be 'login', got %v", entry.Fields["action"])
	}
}

// TestLogger_Warn tests the Warn method
func TestLogger_Warn(t *testing.T) {
	mockLogger := &MockLogger{}
	ctx := context.Background()
	
	err := mockLogger.Warn(ctx, "Warning message")
	
	if err != nil {
		t.Errorf("Warn() should not return error, got %v", err)
	}
	
	if len(mockLogger.entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockLogger.entries))
	}
	
	entry := mockLogger.entries[0]
	if entry.Level != LogLevelWarn {
		t.Errorf("Expected WARN level, got %v", entry.Level)
	}
}

// TestLogger_Error tests the Error method
func TestLogger_Error(t *testing.T) {
	mockLogger := &MockLogger{}
	ctx := context.Background()
	
	err := mockLogger.Error(ctx, "Error message")
	
	if err != nil {
		t.Errorf("Error() should not return error, got %v", err)
	}
	
	if len(mockLogger.entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockLogger.entries))
	}
	
	entry := mockLogger.entries[0]
	if entry.Level != LogLevelError {
		t.Errorf("Expected ERROR level, got %v", entry.Level)
	}
}
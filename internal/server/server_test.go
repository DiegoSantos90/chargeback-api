package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/domain/service"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// MockCreateChargebackUseCase for testing
type MockCreateChargebackUseCase struct {
	ExecuteFunc func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error)
}

func (m *MockCreateChargebackUseCase) Execute(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// testLogger is a simple logger for testing that ignores all output
type testLogger struct{}

func (t *testLogger) Log(ctx context.Context, entry service.LogEntry) error { return nil }
func (t *testLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return nil
}
func (t *testLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return nil
}
func (t *testLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return nil
}
func (t *testLogger) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return nil
}
func (t *testLogger) WithContext(ctx context.Context) service.Logger { return t }

func createTestLogger() service.Logger {
	return &testLogger{}
}

func TestServer_Routes_POST_Chargebacks(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{
		ExecuteFunc: func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
			return &usecase.CreateChargebackResponse{
				ID:              "chargeback-123",
				TransactionID:   req.TransactionID,
				MerchantID:      req.MerchantID,
				Amount:          req.Amount,
				Currency:        req.Currency,
				CardNumber:      "****-****-****-1234",
				Status:          entity.StatusPending,
				Reason:          req.Reason,
				Description:     req.Description,
				CreatedAt:       time.Now(),
				TransactionDate: req.TransactionDate,
			}, nil
		},
	}

	server := NewServer(ServerConfig{
		Port: "8080",
	}, mockUseCase, createTestLogger())

	// Valid request payload
	payload := map[string]interface{}{
		"transaction_id":   "txn-456",
		"merchant_id":      "merchant-123",
		"amount":           99.99,
		"currency":         "USD",
		"card_number":      "1234567890123456",
		"reason":           "fraud",
		"description":      "Suspicious transaction",
		"transaction_date": "2023-01-15T10:30:00Z",
	}

	jsonPayload, _ := json.Marshal(payload)

	// Act
	req := httptest.NewRequest(http.MethodPost, "/chargebacks", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response usecase.CreateChargebackResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.TransactionID != "txn-456" {
		t.Errorf("Expected transaction_id 'txn-456', got '%s'", response.TransactionID)
	}
}

func TestServer_Routes_GET_Health(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	server := NewServer(ServerConfig{
		Port: "8080",
	}, mockUseCase, createTestLogger())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", status)
	}
}

func TestServer_Routes_NotFound(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	server := NewServer(ServerConfig{
		Port: "8080",
	}, mockUseCase, createTestLogger())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestServer_Middleware_CORS(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	server := NewServer(ServerConfig{
		Port: "8080",
	}, mockUseCase, createTestLogger())

	// Act
	req := httptest.NewRequest(http.MethodOptions, "/chargebacks", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expectedValue := range expectedHeaders {
		if actualValue := recorder.Header().Get(header); actualValue != expectedValue {
			t.Errorf("Expected header %s: '%s', got '%s'", header, expectedValue, actualValue)
		}
	}
}

func TestServer_Middleware_Logging(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	server := NewServer(ServerConfig{
		Port: "8080",
	}, mockUseCase, createTestLogger())

	// Act
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	// Assert - This test ensures logging middleware doesn't break the request
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d with logging middleware, got %d", http.StatusOK, recorder.Code)
	}
}

func TestServerConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config ServerConfig
		valid  bool
	}{
		{
			name:   "valid config",
			config: ServerConfig{Port: "8080"},
			valid:  true,
		},
		{
			name:   "empty port",
			config: ServerConfig{Port: ""},
			valid:  false,
		},
		{
			name:   "invalid port format",
			config: ServerConfig{Port: "abc"},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected valid config, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid config, got no error")
			}
		})
	}
}

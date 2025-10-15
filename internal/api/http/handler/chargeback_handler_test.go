package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/api/http/handler"
	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// MockCreateChargebackUseCase is a mock implementation of CreateChargebackUseCase
type MockCreateChargebackUseCase struct {
	ExecuteFunc func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error)
}

func (m *MockCreateChargebackUseCase) Execute(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

func TestChargebackHandler_CreateChargeback_Success(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{
		ExecuteFunc: func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
			return &usecase.CreateChargebackResponse{
				ID:              "cb_12345",
				TransactionID:   req.TransactionID,
				MerchantID:      req.MerchantID,
				Amount:          req.Amount,
				Currency:        req.Currency,
				CardNumber:      "************1111",
				Reason:          req.Reason,
				Status:          entity.StatusPending,
				Description:     req.Description,
				TransactionDate: req.TransactionDate,
				ChargebackDate:  time.Now(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil
		},
	}

	h := handler.NewChargebackHandler(mockUseCase)

	requestBody := map[string]interface{}{
		"transaction_id":   "tx-12345",
		"merchant_id":      "merchant-789",
		"amount":           150.75,
		"currency":         "USD",
		"card_number":      "4111111111111111",
		"reason":           "fraud",
		"description":      "Suspicious transaction",
		"transaction_date": "2023-10-10T10:00:00Z",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/chargebacks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] == nil {
		t.Error("Expected response to contain 'id' field")
	}

	if response["transaction_id"] != "tx-12345" {
		t.Errorf("Expected transaction_id 'tx-12345', got '%v'", response["transaction_id"])
	}

	if response["status"] != "pending" {
		t.Errorf("Expected status 'pending', got '%v'", response["status"])
	}

	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain 'application/json', got '%s'", contentType)
	}
}

func TestChargebackHandler_CreateChargeback_InvalidJSON(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	h := handler.NewChargebackHandler(mockUseCase)

	req := httptest.NewRequest(http.MethodPost, "/chargebacks", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected response to contain 'error' field")
	}
}

func TestChargebackHandler_CreateChargeback_ValidationError(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{
		ExecuteFunc: func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
			return nil, errors.New("validation errors: transaction ID is required")
		},
	}

	h := handler.NewChargebackHandler(mockUseCase)

	requestBody := map[string]interface{}{
		"transaction_id": "", // Invalid - empty
		"merchant_id":    "merchant-789",
		"amount":         150.75,
		"currency":       "USD",
		"card_number":    "4111111111111111",
		"reason":         "fraud",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/chargebacks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected response to contain 'error' field")
	}
}

func TestChargebackHandler_CreateChargeback_DuplicateTransaction(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{
		ExecuteFunc: func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
			return nil, errors.New("chargeback already exists for transaction tx-12345")
		},
	}

	h := handler.NewChargebackHandler(mockUseCase)

	requestBody := map[string]interface{}{
		"transaction_id":   "tx-12345",
		"merchant_id":      "merchant-789",
		"amount":           150.75,
		"currency":         "USD",
		"card_number":      "4111111111111111",
		"reason":           "fraud",
		"transaction_date": "2023-10-10T10:00:00Z",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/chargebacks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected response to contain 'error' field")
	}
}

func TestChargebackHandler_CreateChargeback_InternalServerError(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{
		ExecuteFunc: func(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error) {
			return nil, errors.New("failed to save chargeback: database connection failed")
		},
	}

	h := handler.NewChargebackHandler(mockUseCase)

	requestBody := map[string]interface{}{
		"transaction_id":   "tx-12345",
		"merchant_id":      "merchant-789",
		"amount":           150.75,
		"currency":         "USD",
		"card_number":      "4111111111111111",
		"reason":           "fraud",
		"transaction_date": "2023-10-10T10:00:00Z",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/chargebacks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected response to contain 'error' field")
	}
}

func TestChargebackHandler_CreateChargeback_WrongHTTPMethod(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	h := handler.NewChargebackHandler(mockUseCase)

	req := httptest.NewRequest(http.MethodGet, "/chargebacks", nil)
	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}

func TestChargebackHandler_CreateChargeback_MissingContentType(t *testing.T) {
	// Arrange
	mockUseCase := &MockCreateChargebackUseCase{}
	h := handler.NewChargebackHandler(mockUseCase)

	requestBody := map[string]interface{}{
		"transaction_id": "tx-12345",
		"merchant_id":    "merchant-789",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/chargebacks", bytes.NewReader(jsonBody))
	// Missing Content-Type header

	recorder := httptest.NewRecorder()

	// Act
	h.CreateChargeback(recorder, req)

	// Assert
	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status %d, got %d", http.StatusUnsupportedMediaType, recorder.Code)
	}
}

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// CreateChargebackUseCase interface defines the contract for creating chargebacks
type CreateChargebackUseCase interface {
	Execute(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error)
}

// ChargebackHandler handles HTTP requests for chargeback operations
type ChargebackHandler struct {
	createChargebackUC CreateChargebackUseCase
}

// NewChargebackHandler creates a new chargeback handler
func NewChargebackHandler(createChargebackUC CreateChargebackUseCase) *ChargebackHandler {
	return &ChargebackHandler{
		createChargebackUC: createChargebackUC,
	}
}

// CreateChargebackRequest represents the HTTP request body for creating a chargeback
type CreateChargebackRequest struct {
	TransactionID   string  `json:"transaction_id"`
	MerchantID      string  `json:"merchant_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	CardNumber      string  `json:"card_number"`
	Reason          string  `json:"reason"`
	Description     string  `json:"description,omitempty"`
	TransactionDate string  `json:"transaction_date"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateChargeback handles POST /chargebacks
func (h *ChargebackHandler) CreateChargeback(w http.ResponseWriter, r *http.Request) {
	// Check HTTP method
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check Content-Type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Content-Type must be application/json"})
		return
	}

	// Parse JSON request body
	var req CreateChargebackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Parse transaction date
	transactionDate, err := time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid transaction_date format. Use RFC3339 format"})
		return
	}

	// Convert reason string to enum
	reason, err := parseChargebackReason(req.Reason)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	// Create use case request
	useCaseReq := usecase.CreateChargebackRequest{
		TransactionID:   req.TransactionID,
		MerchantID:      req.MerchantID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		CardNumber:      req.CardNumber,
		Reason:          reason,
		Description:     req.Description,
		TransactionDate: transactionDate,
	}

	// Execute use case
	response, err := h.createChargebackUC.Execute(r.Context(), useCaseReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleUseCaseError handles different types of use case errors and returns appropriate HTTP status codes
func (h *ChargebackHandler) handleUseCaseError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	errorMessage := err.Error()

	// Determine status code based on error type
	switch {
	case strings.Contains(errorMessage, "validation errors"):
		w.WriteHeader(http.StatusBadRequest)
	case strings.Contains(errorMessage, "already exists"):
		w.WriteHeader(http.StatusConflict)
	case strings.Contains(errorMessage, "failed to create chargeback entity"):
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(ErrorResponse{Error: errorMessage})
}

// parseChargebackReason converts string reason to ChargebackReason enum
func parseChargebackReason(reason string) (entity.ChargebackReason, error) {
	switch strings.ToLower(reason) {
	case "fraud":
		return entity.ReasonFraud, nil
	case "authorization_error":
		return entity.ReasonAuthorizationError, nil
	case "processing_error":
		return entity.ReasonProcessingError, nil
	case "consumer_dispute":
		return entity.ReasonConsumerDispute, nil
	default:
		return "", fmt.Errorf("invalid reason '%s'. Valid options: fraud, authorization_error, processing_error, consumer_dispute", reason)
	}
}

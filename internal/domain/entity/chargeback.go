package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ChargebackStatus represents the possible statuses of a chargeback
type ChargebackStatus string

const (
	StatusPending  ChargebackStatus = "pending"
	StatusApproved ChargebackStatus = "approved"
	StatusRejected ChargebackStatus = "rejected"
)

// ChargebackReason represents the reason for the chargeback
type ChargebackReason string

const (
	ReasonFraud              ChargebackReason = "fraud"
	ReasonAuthorizationError ChargebackReason = "authorization_error"
	ReasonProcessingError    ChargebackReason = "processing_error"
	ReasonConsumerDispute    ChargebackReason = "consumer_dispute"
)

// Chargeback represents a chargeback entity in the domain
type Chargeback struct {
	ID              string           `json:"id"`
	TransactionID   string           `json:"transaction_id"`
	MerchantID      string           `json:"merchant_id"`
	Amount          float64          `json:"amount"`
	Currency        string           `json:"currency"`
	CardNumber      string           `json:"card_number"` // Masked card number
	Reason          ChargebackReason `json:"reason"`
	Status          ChargebackStatus `json:"status"`
	Description     string           `json:"description"`
	TransactionDate time.Time        `json:"transaction_date"`
	ChargebackDate  time.Time        `json:"chargeback_date"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// CreateChargebackRequest represents the data needed to create a new chargeback
type CreateChargebackRequest struct {
	TransactionID   string           `json:"transaction_id"`
	MerchantID      string           `json:"merchant_id"`
	Amount          float64          `json:"amount"`
	Currency        string           `json:"currency"`
	CardNumber      string           `json:"card_number"`
	Reason          ChargebackReason `json:"reason"`
	Description     string           `json:"description,omitempty"`
	TransactionDate time.Time        `json:"transaction_date"`
}

// Validate validates the create chargeback request
func (req *CreateChargebackRequest) Validate() error {
	var errors []string

	if strings.TrimSpace(req.TransactionID) == "" {
		errors = append(errors, "transaction ID is required")
	}

	if strings.TrimSpace(req.MerchantID) == "" {
		errors = append(errors, "merchant ID is required")
	}

	if req.Amount <= 0 {
		errors = append(errors, "amount must be greater than zero")
	}

	if strings.TrimSpace(req.Currency) == "" {
		errors = append(errors, "currency is required")
	}

	if strings.TrimSpace(req.CardNumber) == "" {
		errors = append(errors, "card number is required")
	}

	if !isValidReason(req.Reason) {
		errors = append(errors, "invalid chargeback reason")
	}

	if req.TransactionDate.IsZero() {
		errors = append(errors, "transaction date is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// NewChargeback creates a new chargeback from a request
func NewChargeback(req CreateChargebackRequest) (*Chargeback, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	return &Chargeback{
		TransactionID:   req.TransactionID,
		MerchantID:      req.MerchantID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		CardNumber:      maskCardNumber(req.CardNumber),
		Reason:          req.Reason,
		Status:          StatusPending, // Always starts as pending
		Description:     req.Description,
		TransactionDate: req.TransactionDate,
		ChargebackDate:  now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Approve changes the chargeback status to approved
func (c *Chargeback) Approve() error {
	if c.Status != StatusPending {
		return errors.New("only pending chargebacks can be approved")
	}

	c.Status = StatusApproved
	c.UpdatedAt = time.Now()
	return nil
}

// Reject changes the chargeback status to rejected
func (c *Chargeback) Reject() error {
	if c.Status != StatusPending {
		return errors.New("only pending chargebacks can be rejected")
	}

	c.Status = StatusRejected
	c.UpdatedAt = time.Now()
	return nil
}

// IsValid checks if the chargeback has all required fields
func (c *Chargeback) IsValid() bool {
	return c.TransactionID != "" &&
		c.MerchantID != "" &&
		c.Amount > 0 &&
		c.Currency != "" &&
		c.CardNumber != "" &&
		c.Reason != "" &&
		!c.TransactionDate.IsZero() &&
		!c.ChargebackDate.IsZero()
}

// isValidReason checks if the provided reason is valid
func isValidReason(reason ChargebackReason) bool {
	validReasons := []ChargebackReason{
		ReasonFraud,
		ReasonAuthorizationError,
		ReasonProcessingError,
		ReasonConsumerDispute,
	}

	for _, validReason := range validReasons {
		if reason == validReason {
			return true
		}
	}

	return false
}

// maskCardNumber masks the card number showing only the last 4 digits
func maskCardNumber(cardNumber string) string {
	// Remove any spaces or special characters
	cleaned := strings.ReplaceAll(cardNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	if len(cleaned) < 4 {
		return "****"
	}

	// Show only last 4 digits
	lastFour := cleaned[len(cleaned)-4:]
	masked := strings.Repeat("*", len(cleaned)-4) + lastFour

	return masked
}

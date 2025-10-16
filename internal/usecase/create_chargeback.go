package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/domain/repository"
)

// CreateChargebackRequest represents the input for creating a chargeback
type CreateChargebackRequest struct {
	TransactionID   string                  `json:"transaction_id"`
	MerchantID      string                  `json:"merchant_id"`
	Amount          float64                 `json:"amount"`
	Currency        string                  `json:"currency"`
	CardNumber      string                  `json:"card_number"`
	Reason          entity.ChargebackReason `json:"reason"`
	Description     string                  `json:"description,omitempty"`
	TransactionDate time.Time               `json:"transaction_date"`
}

// CreateChargebackResponse represents the output of creating a chargeback
type CreateChargebackResponse struct {
	ID              string                  `json:"id"`
	TransactionID   string                  `json:"transaction_id"`
	MerchantID      string                  `json:"merchant_id"`
	Amount          float64                 `json:"amount"`
	Currency        string                  `json:"currency"`
	CardNumber      string                  `json:"card_number"`
	Reason          entity.ChargebackReason `json:"reason"`
	Status          entity.ChargebackStatus `json:"status"`
	Description     string                  `json:"description"`
	TransactionDate time.Time               `json:"transaction_date"`
	ChargebackDate  time.Time               `json:"chargeback_date"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// CreateChargebackUseCase handles the creation of chargebacks
type CreateChargebackUseCase struct {
	chargebackRepo repository.ChargebackRepository
}

// NewCreateChargebackUseCase creates a new instance of CreateChargebackUseCase
func NewCreateChargebackUseCase(chargebackRepo repository.ChargebackRepository) *CreateChargebackUseCase {
	return &CreateChargebackUseCase{
		chargebackRepo: chargebackRepo,
	}
}

// Execute creates a new chargeback following business rules
func (uc *CreateChargebackUseCase) Execute(ctx context.Context, req CreateChargebackRequest) (*CreateChargebackResponse, error) {
	// 1. Check if chargeback already exists for this transaction
	existingChargeback, err := uc.chargebackRepo.FindByTransactionID(ctx, req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing chargeback: %w", err)
	}

	if existingChargeback != nil {
		return nil, fmt.Errorf("chargeback already exists for transaction %s", req.TransactionID)
	}

	// 2. Create chargeback entity from request
	chargebackReq := entity.CreateChargebackRequest{
		TransactionID:   req.TransactionID,
		MerchantID:      req.MerchantID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		CardNumber:      req.CardNumber,
		Reason:          req.Reason,
		Description:     req.Description,
		TransactionDate: req.TransactionDate,
	}

	chargeback, err := entity.NewChargeback(chargebackReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create chargeback entity: %w", err)
	}

	// 3. Save chargeback to repository
	if err := uc.chargebackRepo.Save(ctx, chargeback); err != nil {
		return nil, fmt.Errorf("failed to save chargeback: %w", err)
	}

	// 4. Return response
	return &CreateChargebackResponse{
		ID:              chargeback.ID,
		TransactionID:   chargeback.TransactionID,
		MerchantID:      chargeback.MerchantID,
		Amount:          chargeback.Amount,
		Currency:        chargeback.Currency,
		CardNumber:      chargeback.CardNumber,
		Reason:          chargeback.Reason,
		Status:          chargeback.Status,
		Description:     chargeback.Description,
		TransactionDate: chargeback.TransactionDate,
		ChargebackDate:  chargeback.ChargebackDate,
		CreatedAt:       chargeback.CreatedAt,
		UpdatedAt:       chargeback.UpdatedAt,
	}, nil
}

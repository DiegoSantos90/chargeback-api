package repository

import (
	"context"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
)

// ChargebackRepository defines the contract for chargeback persistence operations
type ChargebackRepository interface {
	// Save persists a new chargeback to the data store
	Save(ctx context.Context, chargeback *entity.Chargeback) error

	// FindByID retrieves a chargeback by its unique identifier
	FindByID(ctx context.Context, id string) (*entity.Chargeback, error)

	// FindByTransactionID retrieves a chargeback by transaction ID
	FindByTransactionID(ctx context.Context, transactionID string) (*entity.Chargeback, error)

	// FindByMerchantID retrieves all chargebacks for a specific merchant
	FindByMerchantID(ctx context.Context, merchantID string) ([]*entity.Chargeback, error)

	// Update updates an existing chargeback in the data store
	Update(ctx context.Context, chargeback *entity.Chargeback) error

	// Delete removes a chargeback from the data store
	Delete(ctx context.Context, id string) error

	// FindByStatus retrieves chargebacks by their status
	FindByStatus(ctx context.Context, status entity.ChargebackStatus) ([]*entity.Chargeback, error)

	// List retrieves chargebacks with pagination support
	List(ctx context.Context, offset, limit int) ([]*entity.Chargeback, error)
}

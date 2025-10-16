package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// MockChargebackRepository is a mock implementation of ChargebackRepository
type MockChargebackRepository struct {
	SaveFunc                func(ctx context.Context, chargeback *entity.Chargeback) error
	FindByIDFunc            func(ctx context.Context, id string) (*entity.Chargeback, error)
	FindByTransactionIDFunc func(ctx context.Context, transactionID string) (*entity.Chargeback, error)
	FindByMerchantIDFunc    func(ctx context.Context, merchantID string) ([]*entity.Chargeback, error)
	UpdateFunc              func(ctx context.Context, chargeback *entity.Chargeback) error
	DeleteFunc              func(ctx context.Context, id string) error
	FindByStatusFunc        func(ctx context.Context, status entity.ChargebackStatus) ([]*entity.Chargeback, error)
	ListFunc                func(ctx context.Context, offset, limit int) ([]*entity.Chargeback, error)
}

func (m *MockChargebackRepository) Save(ctx context.Context, chargeback *entity.Chargeback) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, chargeback)
	}
	return nil
}

func (m *MockChargebackRepository) FindByID(ctx context.Context, id string) (*entity.Chargeback, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockChargebackRepository) FindByTransactionID(ctx context.Context, transactionID string) (*entity.Chargeback, error) {
	if m.FindByTransactionIDFunc != nil {
		return m.FindByTransactionIDFunc(ctx, transactionID)
	}
	return nil, nil
}

func (m *MockChargebackRepository) FindByMerchantID(ctx context.Context, merchantID string) ([]*entity.Chargeback, error) {
	if m.FindByMerchantIDFunc != nil {
		return m.FindByMerchantIDFunc(ctx, merchantID)
	}
	return nil, nil
}

func (m *MockChargebackRepository) Update(ctx context.Context, chargeback *entity.Chargeback) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, chargeback)
	}
	return nil
}

func (m *MockChargebackRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockChargebackRepository) FindByStatus(ctx context.Context, status entity.ChargebackStatus) ([]*entity.Chargeback, error) {
	if m.FindByStatusFunc != nil {
		return m.FindByStatusFunc(ctx, status)
	}
	return nil, nil
}

func (m *MockChargebackRepository) List(ctx context.Context, offset, limit int) ([]*entity.Chargeback, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return nil, nil
}

func TestCreateChargebackUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockChargebackRepository{
		FindByTransactionIDFunc: func(ctx context.Context, transactionID string) (*entity.Chargeback, error) {
			return nil, nil // No existing chargeback found
		},
		SaveFunc: func(ctx context.Context, chargeback *entity.Chargeback) error {
			// Simulate successful save
			chargeback.ID = "cb_12345"
			return nil
		},
	}

	useCase := usecase.NewCreateChargebackUseCase(mockRepo)
	ctx := context.Background()

	request := usecase.CreateChargebackRequest{
		TransactionID:   "tx-12345",
		MerchantID:      "merchant-789",
		Amount:          150.75,
		Currency:        "USD",
		CardNumber:      "4111111111111111",
		Reason:          entity.ReasonFraud,
		Description:     "Suspicious transaction",
		TransactionDate: time.Now().AddDate(0, 0, -5),
	}

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.ID == "" {
		t.Error("Expected chargeback ID to be set")
	}

	if response.TransactionID != request.TransactionID {
		t.Errorf("Expected TransactionID %s, got %s", request.TransactionID, response.TransactionID)
	}

	if response.Status != entity.StatusPending {
		t.Errorf("Expected status %s, got %s", entity.StatusPending, response.Status)
	}

	if response.CardNumber == request.CardNumber {
		t.Error("Expected card number to be masked")
	}
}

func TestCreateChargebackUseCase_Execute_DuplicateTransaction(t *testing.T) {
	// Arrange
	existingChargeback := &entity.Chargeback{
		ID:            "cb_existing",
		TransactionID: "tx-12345",
		Status:        entity.StatusPending,
	}

	mockRepo := &MockChargebackRepository{
		FindByTransactionIDFunc: func(ctx context.Context, transactionID string) (*entity.Chargeback, error) {
			return existingChargeback, nil // Existing chargeback found
		},
	}

	useCase := usecase.NewCreateChargebackUseCase(mockRepo)
	ctx := context.Background()

	request := usecase.CreateChargebackRequest{
		TransactionID:   "tx-12345",
		MerchantID:      "merchant-789",
		Amount:          150.75,
		Currency:        "USD",
		CardNumber:      "4111111111111111",
		Reason:          entity.ReasonFraud,
		TransactionDate: time.Now().AddDate(0, 0, -5),
	}

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	if err == nil {
		t.Error("Expected error for duplicate transaction, got nil")
	}

	if response != nil {
		t.Error("Expected nil response when error occurs")
	}

	expectedError := "chargeback already exists for transaction tx-12345"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateChargebackUseCase_Execute_InvalidRequest(t *testing.T) {
	// Arrange
	mockRepo := &MockChargebackRepository{}
	useCase := usecase.NewCreateChargebackUseCase(mockRepo)
	ctx := context.Background()

	// Test cases for invalid requests
	testCases := []struct {
		name    string
		request usecase.CreateChargebackRequest
	}{
		{
			name: "empty transaction ID",
			request: usecase.CreateChargebackRequest{
				TransactionID: "", // Invalid
				MerchantID:    "merchant-789",
				Amount:        150.75,
				Currency:      "USD",
				CardNumber:    "4111111111111111",
				Reason:        entity.ReasonFraud,
			},
		},
		{
			name: "zero amount",
			request: usecase.CreateChargebackRequest{
				TransactionID: "tx-12345",
				MerchantID:    "merchant-789",
				Amount:        0, // Invalid
				Currency:      "USD",
				CardNumber:    "4111111111111111",
				Reason:        entity.ReasonFraud,
			},
		},
		{
			name: "empty currency",
			request: usecase.CreateChargebackRequest{
				TransactionID: "tx-12345",
				MerchantID:    "merchant-789",
				Amount:        150.75,
				Currency:      "", // Invalid
				CardNumber:    "4111111111111111",
				Reason:        entity.ReasonFraud,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			response, err := useCase.Execute(ctx, tc.request)

			// Assert
			if err == nil {
				t.Error("Expected validation error, got nil")
			}

			if response != nil {
				t.Error("Expected nil response when validation fails")
			}
		})
	}
}

func TestCreateChargebackUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockChargebackRepository{
		FindByTransactionIDFunc: func(ctx context.Context, transactionID string) (*entity.Chargeback, error) {
			return nil, errors.New("database connection failed")
		},
	}

	useCase := usecase.NewCreateChargebackUseCase(mockRepo)
	ctx := context.Background()

	request := usecase.CreateChargebackRequest{
		TransactionID:   "tx-12345",
		MerchantID:      "merchant-789",
		Amount:          150.75,
		Currency:        "USD",
		CardNumber:      "4111111111111111",
		Reason:          entity.ReasonFraud,
		TransactionDate: time.Now().AddDate(0, 0, -5),
	}

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	if err == nil {
		t.Error("Expected repository error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response when repository error occurs")
	}
}

func TestCreateChargebackUseCase_Execute_SaveError(t *testing.T) {
	// Arrange
	mockRepo := &MockChargebackRepository{
		FindByTransactionIDFunc: func(ctx context.Context, transactionID string) (*entity.Chargeback, error) {
			return nil, nil // No existing chargeback
		},
		SaveFunc: func(ctx context.Context, chargeback *entity.Chargeback) error {
			return errors.New("failed to save to database")
		},
	}

	useCase := usecase.NewCreateChargebackUseCase(mockRepo)
	ctx := context.Background()

	request := usecase.CreateChargebackRequest{
		TransactionID:   "tx-12345",
		MerchantID:      "merchant-789",
		Amount:          150.75,
		Currency:        "USD",
		CardNumber:      "4111111111111111",
		Reason:          entity.ReasonFraud,
		TransactionDate: time.Now().AddDate(0, 0, -5),
	}

	// Act
	response, err := useCase.Execute(ctx, request)

	// Assert
	if err == nil {
		t.Error("Expected save error, got nil")
	}

	if response != nil {
		t.Error("Expected nil response when save error occurs")
	}
}

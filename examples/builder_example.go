package main

import (
	"fmt"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
)

func main() {
	// Example 1: Creating a valid chargeback with Go idioms
	req := entity.CreateChargebackRequest{
		TransactionID:   "tx-12345",
		MerchantID:      "merchant-789",
		Amount:          150.75,
		Currency:        "USD",
		CardNumber:      "4111111111111111",
		Reason:          entity.ReasonFraud,
		Description:     "Suspicious transaction reported by cardholder",
		TransactionDate: time.Now().AddDate(0, 0, -5),
	}

	chargeback, err := entity.NewChargeback(req)
	if err != nil {
		fmt.Printf("Error creating chargeback: %v\n", err)
		return
	}

	fmt.Printf("Chargeback created successfully:\n")
	fmt.Printf("Transaction ID: %s\n", chargeback.TransactionID)
	fmt.Printf("Amount: %.2f %s\n", chargeback.Amount, chargeback.Currency)
	fmt.Printf("Masked Card: %s\n", chargeback.CardNumber)
	fmt.Printf("Status: %s\n", chargeback.Status)
	fmt.Printf("Reason: %s\n", chargeback.Reason)

	// Example 2: Request with validation errors
	invalidReq := entity.CreateChargebackRequest{
		TransactionID: "",   // Empty - will cause error
		Amount:        -100, // Negative - will cause error
		Currency:      "USD",
	}

	_, err = entity.NewChargeback(invalidReq)
	if err != nil {
		fmt.Printf("\nValidation errors (as expected): %v\n", err)
	}

	// Example 3: Inline creation for simple cases
	chargeback2, err := entity.NewChargeback(entity.CreateChargebackRequest{
		TransactionID:   "tx-67890",
		MerchantID:      "merchant-456",
		Amount:          99.99,
		Currency:        "BRL",
		CardNumber:      "5555555555554444",
		Reason:          entity.ReasonConsumerDispute,
		Description:     "Customer dispute",
		TransactionDate: time.Now().AddDate(0, 0, -10),
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nSecond chargeback created: %s\n", chargeback2.TransactionID)

	// Example 4: Pre-validation before creation
	req3 := entity.CreateChargebackRequest{
		TransactionID:   "tx-99999",
		MerchantID:      "merchant-999",
		Amount:          200.00,
		Currency:        "EUR",
		CardNumber:      "4000000000000002",
		Reason:          entity.ReasonProcessingError,
		TransactionDate: time.Now().AddDate(0, 0, -1),
	}

	// Can validate separately if needed
	if err := req3.Validate(); err != nil {
		fmt.Printf("Request validation failed: %v\n", err)
		return
	}

	chargeback3, err := entity.NewChargeback(req3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Third chargeback created: %s\n", chargeback3.TransactionID)
}

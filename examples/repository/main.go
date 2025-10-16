package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/infra/db"
	"github.com/DiegoSantos90/chargeback-api/internal/infra/repository"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	cfg := db.LoadDynamoDBConfigFromEnv()
	fmt.Printf("DynamoDB Config: %+v\n", cfg)

	// Create DynamoDB client
	client, err := db.NewDynamoDBClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	// Create repository
	repo := repository.NewDynamoDBChargebackRepository(client, cfg.TableName)

	// Example 1: Create and save a chargeback
	fmt.Println("\n=== Creating a new chargeback ===")

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
		log.Fatalf("Failed to create chargeback: %v", err)
	}

	// Save to DynamoDB
	if err := repo.Save(ctx, chargeback); err != nil {
		log.Printf("Failed to save chargeback: %v", err)
		// In a real application, you might want to handle this error differently
		// For now, we'll continue with the example
	} else {
		fmt.Printf("Chargeback saved successfully with ID: %s\n", chargeback.ID)
	}

	// Example 2: Find by ID
	fmt.Println("\n=== Finding chargeback by ID ===")

	if chargeback.ID != "" {
		found, err := repo.FindByID(ctx, chargeback.ID)
		if err != nil {
			log.Printf("Failed to find chargeback: %v", err)
		} else if found != nil {
			fmt.Printf("Found chargeback: %s (Amount: %.2f %s)\n",
				found.TransactionID, found.Amount, found.Currency)
		} else {
			fmt.Println("Chargeback not found")
		}
	}

	// Example 3: Find by Transaction ID
	fmt.Println("\n=== Finding chargeback by Transaction ID ===")

	foundByTx, err := repo.FindByTransactionID(ctx, req.TransactionID)
	if err != nil {
		log.Printf("Failed to find chargeback by transaction ID: %v", err)
	} else if foundByTx != nil {
		fmt.Printf("Found chargeback by transaction ID: %s (Status: %s)\n",
			foundByTx.ID, foundByTx.Status)
	} else {
		fmt.Println("Chargeback not found by transaction ID")
	}

	// Example 4: Find by Merchant ID
	fmt.Println("\n=== Finding chargebacks by Merchant ID ===")

	merchantChargebacks, err := repo.FindByMerchantID(ctx, req.MerchantID)
	if err != nil {
		log.Printf("Failed to find chargebacks by merchant ID: %v", err)
	} else {
		fmt.Printf("Found %d chargebacks for merchant %s\n",
			len(merchantChargebacks), req.MerchantID)
		for i, cb := range merchantChargebacks {
			fmt.Printf("  %d. %s - %.2f %s (%s)\n",
				i+1, cb.TransactionID, cb.Amount, cb.Currency, cb.Status)
		}
	}

	// Example 5: Find by Status
	fmt.Println("\n=== Finding chargebacks by Status ===")

	pendingChargebacks, err := repo.FindByStatus(ctx, entity.StatusPending)
	if err != nil {
		log.Printf("Failed to find chargebacks by status: %v", err)
	} else {
		fmt.Printf("Found %d pending chargebacks\n", len(pendingChargebacks))
	}

	// Example 6: Update chargeback status
	fmt.Println("\n=== Updating chargeback status ===")

	if chargeback.ID != "" {
		if err := chargeback.Approve(); err != nil {
			log.Printf("Failed to approve chargeback: %v", err)
		} else {
			if err := repo.Update(ctx, chargeback); err != nil {
				log.Printf("Failed to update chargeback: %v", err)
			} else {
				fmt.Printf("Chargeback %s approved successfully\n", chargeback.ID)
			}
		}
	}

	// Example 7: List with pagination
	fmt.Println("\n=== Listing chargebacks with pagination ===")

	chargebacks, err := repo.List(ctx, 0, 10) // First 10 items
	if err != nil {
		log.Printf("Failed to list chargebacks: %v", err)
	} else {
		fmt.Printf("Listed %d chargebacks\n", len(chargebacks))
		for i, cb := range chargebacks {
			fmt.Printf("  %d. %s - %s (%s)\n",
				i+1, cb.ID, cb.TransactionID, cb.Status)
		}
	}

	fmt.Println("\n=== Repository example completed ===")
	fmt.Println("Note: Some operations might fail if DynamoDB is not running or not configured properly.")
	fmt.Println("To run DynamoDB Local: docker run -p 8000:8000 amazon/dynamodb-local")
}

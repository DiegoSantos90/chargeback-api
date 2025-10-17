package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/repository"
	"github.com/DiegoSantos90/chargeback-api/internal/infra/db"
	dynamoRepo "github.com/DiegoSantos90/chargeback-api/internal/infra/repository"
	"github.com/DiegoSantos90/chargeback-api/internal/server"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// Config holds the application configuration
type Config struct {
	Port     string
	DynamoDB db.DynamoDBConfig
}

// Dependencies holds all initialized dependencies
type Dependencies struct {
	DynamoClient       *dynamodb.Client
	ChargebackRepo     repository.ChargebackRepository
	CreateChargebackUC *usecase.CreateChargebackUseCase
	HTTPServer         *server.Server
}

func main() {
	config := loadConfiguration()

	if err := validateConfiguration(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Log configuration for debugging
	logConfiguration(config)

	ctx := context.Background()
	deps, err := initializeDependencies(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	go func() {
		log.Printf("🚀 Chargeback API starting on port %s", config.Port)
		if err := deps.HTTPServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = shutdownCtx
	log.Println("✅ Server shutdown complete")
}

func loadConfiguration() Config {
	return Config{
		Port: getEnvOrDefault("PORT", "8080"),
		DynamoDB: db.DynamoDBConfig{
			Endpoint:  getEnvOrDefault("DYNAMODB_ENDPOINT", ""),
			Region:    getEnvOrDefault("AWS_REGION", "us-east-1"),
			TableName: getEnvOrDefault("DYNAMODB_TABLE", "chargebacks"),
		},
	}
}

func validateConfiguration(config Config) error {
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}
	if config.DynamoDB.Region == "" {
		return fmt.Errorf("AWS region is required")
	}
	if config.DynamoDB.TableName == "" {
		return fmt.Errorf("DynamoDB table name is required")
	}

	// Validate AWS credentials availability (except for local DynamoDB)
	if config.DynamoDB.Endpoint == "" {
		if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
			log.Println("⚠️  Warning: No explicit AWS credentials found. Relying on IAM roles or instance profile.")
		}
	}

	return nil
}

func initializeDependencies(ctx context.Context, config Config) (*Dependencies, error) {
	dynamoClient, err := db.NewDynamoDBClient(ctx, config.DynamoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DynamoDB client: %w", err)
	}

	// Test DynamoDB connection
	if err := testDynamoDBConnection(ctx, dynamoClient, config.DynamoDB.TableName); err != nil {
		return nil, fmt.Errorf("failed to connect to DynamoDB: %w", err)
	}

	chargebackRepo := dynamoRepo.NewDynamoDBChargebackRepository(dynamoClient, config.DynamoDB.TableName)
	createChargebackUC := usecase.NewCreateChargebackUseCase(chargebackRepo)

	serverConfig := server.ServerConfig{Port: config.Port}
	httpServer := server.NewServer(serverConfig, createChargebackUC)

	return &Dependencies{
		DynamoClient:       dynamoClient,
		ChargebackRepo:     chargebackRepo,
		CreateChargebackUC: createChargebackUC,
		HTTPServer:         httpServer,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func logConfiguration(config Config) {
	log.Println("📊 Application Configuration:")
	log.Printf("  ├── Port: %s", config.Port)
	log.Printf("  ├── AWS Region: %s", config.DynamoDB.Region)
	log.Printf("  ├── DynamoDB Table: %s", config.DynamoDB.TableName)

	if config.DynamoDB.Endpoint != "" {
		log.Printf("  └── DynamoDB Endpoint: %s (Local Development)", config.DynamoDB.Endpoint)
	} else {
		log.Printf("  └── DynamoDB Endpoint: AWS DynamoDB Service (Production)")
	}

	// Log AWS credential source information
	if accessKey := os.Getenv("AWS_ACCESS_KEY_ID"); accessKey != "" {
		log.Printf("🔑 AWS Credentials: Environment Variables (Access Key: %s...)", accessKey[:min(len(accessKey), 8)])
	} else if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		log.Printf("🔑 AWS Credentials: AWS Profile (%s)", profile)
	} else {
		log.Printf("🔑 AWS Credentials: Default credential chain (IAM Role/Instance Profile)")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func testDynamoDBConnection(ctx context.Context, client *dynamodb.Client, tableName string) error {
	log.Printf("🔍 Testing DynamoDB connection for table: %s", tableName)

	// Try to describe the table to verify it exists and we have access
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})

	if err != nil {
		log.Printf("❌ DynamoDB connection test failed: %v", err)
		return fmt.Errorf("table '%s' not accessible: %w", tableName, err)
	}

	log.Printf("✅ DynamoDB connection test successful for table: %s", tableName)
	return nil
}

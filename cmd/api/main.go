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
	return nil
}

func initializeDependencies(ctx context.Context, config Config) (*Dependencies, error) {
	dynamoClient, err := db.NewDynamoDBClient(ctx, config.DynamoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DynamoDB client: %w", err)
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

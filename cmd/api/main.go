package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/repository"
	"github.com/DiegoSantos90/chargeback-api/internal/domain/service"
	"github.com/DiegoSantos90/chargeback-api/internal/infra/db"
	"github.com/DiegoSantos90/chargeback-api/internal/infra/logging"
	dynamoRepo "github.com/DiegoSantos90/chargeback-api/internal/infra/repository"
	"github.com/DiegoSantos90/chargeback-api/internal/server"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// Config holds the application configuration
type Config struct {
	Port     string
	DynamoDB db.DynamoDBConfig
	Logging  LoggingConfig
}

// LoggingConfig holds the logging configuration
type LoggingConfig struct {
	Level   service.LogLevel
	Format  logging.LogFormat
	Service string
	Version string
}

// Dependencies holds all initialized dependencies
type Dependencies struct {
	Logger             service.Logger
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
		deps.Logger.Info(ctx, "Chargeback API starting", map[string]interface{}{
			"port": config.Port,
		})
		if err := deps.HTTPServer.Start(); err != nil {
			deps.Logger.Error(ctx, "Failed to start server", map[string]interface{}{
				"error": err.Error(),
			})
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	deps.Logger.Info(ctx, "Shutting down server", nil)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = shutdownCtx
	deps.Logger.Info(ctx, "Server shutdown complete", nil)
}

func loadConfiguration() Config {
	return Config{
		Port: getEnvOrDefault("PORT", "8080"),
		DynamoDB: db.DynamoDBConfig{
			Endpoint:  getEnvOrDefault("DYNAMODB_ENDPOINT", ""),
			Region:    getEnvOrDefault("AWS_REGION", "us-east-1"),
			TableName: getEnvOrDefault("DYNAMODB_TABLE", "chargebacks"),
		},
		Logging: LoggingConfig{
			Level:   parseLogLevel(getEnvOrDefault("LOG_LEVEL", "info")),
			Format:  parseLogFormat(getEnvOrDefault("LOG_FORMAT", "json")),
			Service: "chargeback-api",
			Version: getEnvOrDefault("APP_VERSION", "dev"),
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
	// Initialize logger first
	loggerConfig := logging.LoggerConfig{
		Level:       config.Logging.Level,
		Format:      config.Logging.Format,
		ServiceName: config.Logging.Service,
		Version:     config.Logging.Version,
	}

	logger, err := logging.NewStructuredLogger(loggerConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Log configuration using the new logger
	if err := logger.Info(ctx, "Application starting", map[string]interface{}{
		"port":           config.Port,
		"aws_region":     config.DynamoDB.Region,
		"dynamodb_table": config.DynamoDB.TableName,
		"log_level":      config.Logging.Level.String(),
		"log_format":     config.Logging.Format.String(),
		"service_name":   config.Logging.Service,
		"version":        config.Logging.Version,
	}); err != nil {
		return nil, fmt.Errorf("failed to log application startup: %w", err)
	}

	dynamoClient, err := db.NewDynamoDBClient(ctx, config.DynamoDB)
	if err != nil {
		logger.Error(ctx, "Failed to initialize DynamoDB client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize DynamoDB client: %w", err)
	}

	// Test DynamoDB connection
	if err := testDynamoDBConnection(ctx, dynamoClient, config.DynamoDB.TableName, logger); err != nil {
		logger.Error(ctx, "Failed to connect to DynamoDB", map[string]interface{}{
			"error":      err.Error(),
			"table_name": config.DynamoDB.TableName,
		})
		return nil, fmt.Errorf("failed to connect to DynamoDB: %w", err)
	}

	chargebackRepo := dynamoRepo.NewDynamoDBChargebackRepository(dynamoClient, config.DynamoDB.TableName)
	createChargebackUC := usecase.NewCreateChargebackUseCase(chargebackRepo)

	serverConfig := server.ServerConfig{Port: config.Port}
	httpServer := server.NewServer(serverConfig, createChargebackUC, logger)

	return &Dependencies{
		Logger:             logger,
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

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) service.LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return service.LogLevelDebug
	case "info":
		return service.LogLevelInfo
	case "warn", "warning":
		return service.LogLevelWarn
	case "error":
		return service.LogLevelError
	default:
		return service.LogLevelInfo
	}
}

// parseLogFormat converts string to LogFormat
func parseLogFormat(format string) logging.LogFormat {
	switch strings.ToLower(format) {
	case "json":
		return logging.FormatJSON
	case "text":
		return logging.FormatText
	default:
		return logging.FormatJSON
	}
}

func testDynamoDBConnection(ctx context.Context, client *dynamodb.Client, tableName string, logger service.Logger) error {
	logger.Info(ctx, "Testing DynamoDB connection", map[string]interface{}{
		"table_name": tableName,
	})

	// Try to describe the table to verify it exists and we have access
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})

	if err != nil {
		logger.Error(ctx, "DynamoDB connection test failed", map[string]interface{}{
			"error":      err.Error(),
			"table_name": tableName,
		})
		return fmt.Errorf("table '%s' not accessible: %w", tableName, err)
	}

	logger.Info(ctx, "DynamoDB connection test successful", map[string]interface{}{
		"table_name": tableName,
	})
	return nil
}

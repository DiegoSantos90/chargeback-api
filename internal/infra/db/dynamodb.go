package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBConfig holds the configuration for DynamoDB
type DynamoDBConfig struct {
	Endpoint  string
	Region    string
	TableName string
}

// NewDynamoDBClient creates a new DynamoDB client with proper credential handling
func NewDynamoDBClient(ctx context.Context, cfg DynamoDBConfig) (*dynamodb.Client, error) {
	var awsCfg aws.Config
	var err error

	// Check if we're in local development mode
	if cfg.Endpoint != "" {
		log.Printf("🔧 Local DynamoDB endpoint detected: %s", cfg.Endpoint)

		// For local development, use static credentials or default config
		if accessKey := os.Getenv("AWS_ACCESS_KEY_ID"); accessKey != "" {
			log.Println("🔑 Using AWS credentials from environment variables")
			awsCfg, err = config.LoadDefaultConfig(ctx,
				config.WithRegion(cfg.Region),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					os.Getenv("AWS_ACCESS_KEY_ID"),
					os.Getenv("AWS_SECRET_ACCESS_KEY"),
					os.Getenv("AWS_SESSION_TOKEN"),
				)),
			)
		} else {
			log.Println("🔑 Using static credentials for local DynamoDB")
			// For DynamoDB Local, use well-known static credentials
			awsCfg, err = config.LoadDefaultConfig(ctx,
				config.WithRegion(cfg.Region),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					"dummy-access-key-id",
					"dummy-secret-access-key",
					"",
				)),
			)
		}
	} else {
		log.Println("🔑 Loading AWS credentials using default credential chain")
		// For production, use the default credential chain
		// This will try: environment variables -> IAM roles -> AWS profiles -> etc.
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Log the region being used
	log.Printf("🌍 Using AWS region: %s", awsCfg.Region)

	// Create DynamoDB client options
	var clientOptions []func(*dynamodb.Options)

	// Override endpoint if specified (for DynamoDB Local)
	if cfg.Endpoint != "" {
		clientOptions = append(clientOptions, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
		log.Printf("🎯 DynamoDB client configured with custom endpoint: %s", cfg.Endpoint)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(awsCfg, clientOptions...)

	log.Printf("✅ DynamoDB client initialized successfully for table: %s", cfg.TableName)
	return client, nil
}

// LoadDynamoDBConfigFromEnv loads DynamoDB configuration from environment variables
func LoadDynamoDBConfigFromEnv() DynamoDBConfig {
	return DynamoDBConfig{
		Endpoint:  os.Getenv("DYNAMODB_ENDPOINT"),
		Region:    getEnvWithDefault("AWS_REGION", "us-east-1"),
		TableName: getEnvWithDefault("CHARGEBACK_TABLE_NAME", "chargebacks"),
	}
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

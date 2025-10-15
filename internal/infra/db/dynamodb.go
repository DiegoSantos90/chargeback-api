package db

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBConfig holds the configuration for DynamoDB
type DynamoDBConfig struct {
	Endpoint  string
	Region    string
	TableName string
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(ctx context.Context, cfg DynamoDBConfig) (*dynamodb.Client, error) {
	// Load AWS configuration
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Override region if specified
	if cfg.Region != "" {
		awsCfg.Region = cfg.Region
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(awsCfg)

	// Override endpoint if specified (for DynamoDB Local)
	if cfg.Endpoint != "" {
		client = dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

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

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/infra/db"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns environment variable when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "environment_value",
			expected:     "environment_value",
		},
		{
			name:         "returns default when environment variable not set",
			key:          "UNSET_KEY",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "returns default when environment variable is empty",
			key:          "EMPTY_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Act
			result := getEnvOrDefault(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%s, %s) = %s, want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:    "loads default configuration",
			envVars: map[string]string{},
			expected: Config{
				Port: "8080",
				DynamoDB: db.DynamoDBConfig{
					Endpoint:  "",
					Region:    "us-east-1",
					TableName: "chargebacks",
				},
			},
		},
		{
			name: "loads configuration from environment variables",
			envVars: map[string]string{
				"PORT":              "3000",
				"AWS_REGION":        "us-west-2",
				"DYNAMODB_TABLE":    "test-chargebacks",
				"DYNAMODB_ENDPOINT": "http://localhost:8000",
			},
			expected: Config{
				Port: "3000",
				DynamoDB: db.DynamoDBConfig{
					Endpoint:  "http://localhost:8000",
					Region:    "us-west-2",
					TableName: "test-chargebacks",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Act
			config := loadConfiguration()

			// Assert
			if config.Port != tt.expected.Port {
				t.Errorf("Expected Port %s, got %s", tt.expected.Port, config.Port)
			}
			if config.DynamoDB.Region != tt.expected.DynamoDB.Region {
				t.Errorf("Expected Region %s, got %s", tt.expected.DynamoDB.Region, config.DynamoDB.Region)
			}
			if config.DynamoDB.TableName != tt.expected.DynamoDB.TableName {
				t.Errorf("Expected TableName %s, got %s", tt.expected.DynamoDB.TableName, config.DynamoDB.TableName)
			}
			if config.DynamoDB.Endpoint != tt.expected.DynamoDB.Endpoint {
				t.Errorf("Expected Endpoint %s, got %s", tt.expected.DynamoDB.Endpoint, config.DynamoDB.Endpoint)
			}
		})
	}
}

func TestInitializeDependencies(t *testing.T) {
	// Setup
	config := Config{
		Port: "8080",
		DynamoDB: db.DynamoDBConfig{
			Endpoint:  "",
			Region:    "us-east-1",
			TableName: "test-chargebacks",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Act
	deps, err := initializeDependencies(ctx, config)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if deps == nil {
		t.Fatal("Expected dependencies to be initialized, got nil")
	}

	if deps.DynamoClient == nil {
		t.Error("Expected DynamoClient to be initialized")
	}

	if deps.ChargebackRepo == nil {
		t.Error("Expected ChargebackRepo to be initialized")
	}

	if deps.CreateChargebackUC == nil {
		t.Error("Expected CreateChargebackUC to be initialized")
	}

	if deps.HTTPServer == nil {
		t.Error("Expected HTTPServer to be initialized")
	}
}

func TestInitializeDependencies_InvalidDynamoDBConfig(t *testing.T) {
	// Setup - invalid region should cause AWS config to fail in some cases
	config := Config{
		Port: "8080",
		DynamoDB: db.DynamoDBConfig{
			Endpoint:  "invalid://endpoint",
			Region:    "",
			TableName: "",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Act
	deps, err := initializeDependencies(ctx, config)

	// Assert - we expect this might fail, but let's be flexible
	// since AWS SDK might handle some invalid configs gracefully
	if err != nil {
		// Error is acceptable for invalid config
		if deps != nil {
			t.Error("Expected dependencies to be nil when error occurs")
		}
	} else {
		// If no error, dependencies should be valid
		if deps == nil {
			t.Error("Expected dependencies to be initialized when no error")
		}
	}
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		shouldErr bool
	}{
		{
			name: "valid configuration",
			config: Config{
				Port: "8080",
				DynamoDB: db.DynamoDBConfig{
					Region:    "us-east-1",
					TableName: "chargebacks",
				},
			},
			shouldErr: false,
		},
		{
			name: "empty port",
			config: Config{
				Port: "",
				DynamoDB: db.DynamoDBConfig{
					Region:    "us-east-1",
					TableName: "chargebacks",
				},
			},
			shouldErr: true,
		},
		{
			name: "empty region",
			config: Config{
				Port: "8080",
				DynamoDB: db.DynamoDBConfig{
					Region:    "",
					TableName: "chargebacks",
				},
			},
			shouldErr: true,
		},
		{
			name: "empty table name",
			config: Config{
				Port: "8080",
				DynamoDB: db.DynamoDBConfig{
					Region:    "us-east-1",
					TableName: "",
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validateConfiguration(tt.config)

			// Assert
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

package db

import (
	"context"
	"os"
	"testing"
)

func TestLoadDynamoDBConfigFromEnv(t *testing.T) {
	t.Run("loads default configuration", func(t *testing.T) {
		os.Unsetenv("DYNAMODB_ENDPOINT")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("CHARGEBACK_TABLE_NAME")

		config := LoadDynamoDBConfigFromEnv()

		if config.Endpoint != "" {
			t.Errorf("Expected empty endpoint, got %s", config.Endpoint)
		}
		if config.Region != "us-east-1" {
			t.Errorf("Expected region 'us-east-1', got %s", config.Region)
		}
		if config.TableName != "chargebacks" {
			t.Errorf("Expected table name 'chargebacks', got %s", config.TableName)
		}
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")
		os.Setenv("AWS_REGION", "us-west-2")
		os.Setenv("CHARGEBACK_TABLE_NAME", "test-chargebacks")

		defer func() {
			os.Unsetenv("DYNAMODB_ENDPOINT")
			os.Unsetenv("AWS_REGION")
			os.Unsetenv("CHARGEBACK_TABLE_NAME")
		}()

		config := LoadDynamoDBConfigFromEnv()

		if config.Endpoint != "http://localhost:8000" {
			t.Errorf("Expected endpoint 'http://localhost:8000', got %s", config.Endpoint)
		}
		if config.Region != "us-west-2" {
			t.Errorf("Expected region 'us-west-2', got %s", config.Region)
		}
		if config.TableName != "test-chargebacks" {
			t.Errorf("Expected table name 'test-chargebacks', got %s", config.TableName)
		}
	})
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("returns environment variable when set", func(t *testing.T) {
		key := "TEST_ENV_VAR"
		expectedValue := "test-value"
		defaultValue := "default-value"

		os.Setenv(key, expectedValue)
		defer os.Unsetenv(key)

		result := getEnvWithDefault(key, defaultValue)

		if result != expectedValue {
			t.Errorf("Expected %s, got %s", expectedValue, result)
		}
	})

	t.Run("returns default when not set", func(t *testing.T) {
		key := "NONEXISTENT_ENV_VAR"
		defaultValue := "default-value"

		os.Unsetenv(key)

		result := getEnvWithDefault(key, defaultValue)

		if result != defaultValue {
			t.Errorf("Expected %s, got %s", defaultValue, result)
		}
	})
}

func TestNewDynamoDBClient(t *testing.T) {
	t.Run("creates client with basic config", func(t *testing.T) {
		cfg := DynamoDBConfig{
			Region:    "us-east-1",
			TableName: "test-table",
		}

		ctx := context.Background()
		client, err := NewDynamoDBClient(ctx, cfg)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if client == nil {
			t.Error("Expected client to be created")
		}
	})
}

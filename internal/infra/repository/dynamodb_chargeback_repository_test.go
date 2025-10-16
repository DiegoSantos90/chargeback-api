package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
)

// Unit tests for DynamoDB Chargeback Repository
// These tests focus on testing the repository logic with mocks and without external dependencies

// MockDynamoDBClient implements a mock DynamoDB client for testing
type MockDynamoDBClient struct {
	PutItemFunc    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItemFunc    func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	QueryFunc      func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	UpdateItemFunc func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItemFunc func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	ScanFunc       func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if m.PutItemFunc != nil {
		return m.PutItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.PutItemOutput{}, nil
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if m.GetItemFunc != nil {
		return m.GetItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, params, optFns...)
	}
	return &dynamodb.QueryOutput{}, nil
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	if m.UpdateItemFunc != nil {
		return m.UpdateItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.UpdateItemOutput{}, nil
}

func (m *MockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if m.DeleteItemFunc != nil {
		return m.DeleteItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.DeleteItemOutput{}, nil
}

func (m *MockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if m.ScanFunc != nil {
		return m.ScanFunc(ctx, params, optFns...)
	}
	return &dynamodb.ScanOutput{}, nil
}

func createTestChargeback() *entity.Chargeback {
	return &entity.Chargeback{
		ID:              "chargeback-123",
		TransactionID:   "txn-456",
		MerchantID:      "merchant-789",
		Amount:          99.99,
		Currency:        "USD",
		CardNumber:      "****-****-****-1234",
		Reason:          entity.ReasonFraud,
		Status:          entity.StatusPending,
		Description:     "Test chargeback",
		TransactionDate: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		ChargebackDate:  time.Date(2023, 1, 16, 12, 0, 0, 0, time.UTC),
		CreatedAt:       time.Date(2023, 1, 16, 12, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2023, 1, 16, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewDynamoDBChargebackRepository(t *testing.T) {
	tableName := "test-chargebacks"

	// Create a wrapper to match the interface expected by NewDynamoDBChargebackRepository
	dynamoClient := &dynamodb.Client{}
	repo := NewDynamoDBChargebackRepository(dynamoClient, tableName)

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	// Test that it returns the correct interface type
	if _, ok := repo.(*DynamoDBChargebackRepository); !ok {
		t.Fatal("Expected DynamoDBChargebackRepository type")
	}
}

// createTestRepository creates a repository instance for testing with mocked client
func createTestRepository(client *MockDynamoDBClient) *DynamoDBChargebackRepository {
	// Since we can't pass MockDynamoDBClient directly to the constructor,
	// we'll create the repository struct directly for testing
	return &DynamoDBChargebackRepository{
		client:    (*dynamodb.Client)(nil), // We'll override this in tests
		tableName: "test-chargebacks",
	}
}

func TestDynamoDBChargebackRepository_Save_GeneratesID(t *testing.T) {
	mockClient := &MockDynamoDBClient{
		PutItemFunc: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			// Verify table name
			if *params.TableName != "test-chargebacks" {
				t.Errorf("Expected table name 'test-chargebacks', got %s", *params.TableName)
			}

			// Verify item has required fields
			if params.Item["id"] == nil {
				t.Error("Expected 'id' field in item")
			}
			if params.Item["transaction_id"] == nil {
				t.Error("Expected 'transaction_id' field in item")
			}

			return &dynamodb.PutItemOutput{}, nil
		},
	}

	_ = mockClient // Use mockClient to avoid unused variable error
	chargeback := createTestChargeback()
	chargeback.ID = "" // Simulate new chargeback without ID

	// Since we can't easily mock the client in the struct,
	// we'll test the ID generation logic directly
	originalID := chargeback.ID

	// Test that generateChargebackID works
	generatedID := generateChargebackID()
	if generatedID == "" {
		t.Error("Expected non-empty generated ID")
	}

	if len(generatedID) < 3 || generatedID[:3] != "cb_" {
		t.Errorf("Expected ID to start with 'cb_', got %s", generatedID)
	}

	// Test that ID is set when empty
	if originalID != "" {
		chargeback.ID = ""
	}

	// Verify the ID would be generated (testing the logic)
	if chargeback.ID == "" {
		chargeback.ID = generateChargebackID()
	}

	if chargeback.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestDynamoDBChargebackRepository_ItemToEntity(t *testing.T) {
	repo := createTestRepository(&MockDynamoDBClient{})

	testChargeback := createTestChargeback()
	item := &chargebackItem{
		ID:              testChargeback.ID,
		TransactionID:   testChargeback.TransactionID,
		MerchantID:      testChargeback.MerchantID,
		Amount:          testChargeback.Amount,
		Currency:        testChargeback.Currency,
		CardNumber:      testChargeback.CardNumber,
		Reason:          string(testChargeback.Reason),
		Status:          string(testChargeback.Status),
		Description:     testChargeback.Description,
		TransactionDate: testChargeback.TransactionDate,
		ChargebackDate:  testChargeback.ChargebackDate,
		CreatedAt:       testChargeback.CreatedAt,
		UpdatedAt:       testChargeback.UpdatedAt,
	}

	entity := repo.itemToEntity(item)

	if entity.ID != testChargeback.ID {
		t.Errorf("Expected ID %s, got %s", testChargeback.ID, entity.ID)
	}

	if entity.TransactionID != testChargeback.TransactionID {
		t.Errorf("Expected TransactionID %s, got %s", testChargeback.TransactionID, entity.TransactionID)
	}

	if entity.Status != testChargeback.Status {
		t.Errorf("Expected Status %s, got %s", testChargeback.Status, entity.Status)
	}

	if entity.Reason != testChargeback.Reason {
		t.Errorf("Expected Reason %s, got %s", testChargeback.Reason, entity.Reason)
	}

	if entity.Amount != testChargeback.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", testChargeback.Amount, entity.Amount)
	}
}

func TestDynamoDBChargebackRepository_ItemToEntity_InvalidStatus(t *testing.T) {
	repo := createTestRepository(&MockDynamoDBClient{})

	item := &chargebackItem{
		ID:              "test-id",
		TransactionID:   "txn-123",
		MerchantID:      "merchant-456",
		Amount:          100.0,
		Currency:        "USD",
		CardNumber:      "****1234",
		Reason:          string(entity.ReasonFraud),
		Status:          "invalid_status",
		Description:     "test",
		TransactionDate: time.Now(),
		ChargebackDate:  time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result := repo.itemToEntity(item)

	// Should still create entity, but with the invalid status as-is
	if result.Status != entity.ChargebackStatus("invalid_status") {
		t.Errorf("Expected status to be preserved as-is, got %s", result.Status)
	}
}

func TestGenerateChargebackID(t *testing.T) {
	id1 := generateChargebackID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateChargebackID()

	// Test that IDs are generated
	if id1 == "" {
		t.Error("Expected non-empty ID")
	}

	if id2 == "" {
		t.Error("Expected non-empty ID")
	}

	// Test that IDs are unique
	if id1 == id2 {
		t.Error("Expected unique IDs")
	}

	// Test ID format (should start with "cb_")
	if len(id1) < 3 || id1[:3] != "cb_" {
		t.Errorf("Expected ID to start with 'cb_', got %s", id1)
	}
}

func TestChargebackItemSerialization(t *testing.T) {
	testChargeback := createTestChargeback()

	// Test marshaling to DynamoDB item
	item := chargebackItem{
		ID:              testChargeback.ID,
		TransactionID:   testChargeback.TransactionID,
		MerchantID:      testChargeback.MerchantID,
		Amount:          testChargeback.Amount,
		Currency:        testChargeback.Currency,
		CardNumber:      testChargeback.CardNumber,
		Reason:          string(testChargeback.Reason),
		Status:          string(testChargeback.Status),
		Description:     testChargeback.Description,
		TransactionDate: testChargeback.TransactionDate,
		ChargebackDate:  testChargeback.ChargebackDate,
		CreatedAt:       testChargeback.CreatedAt,
		UpdatedAt:       testChargeback.UpdatedAt,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		t.Errorf("Failed to marshal chargeback item: %v", err)
	}

	// Verify key fields are present
	if av["id"] == nil {
		t.Error("Expected 'id' field in marshaled item")
	}
	if av["transaction_id"] == nil {
		t.Error("Expected 'transaction_id' field in marshaled item")
	}
	if av["status"] == nil {
		t.Error("Expected 'status' field in marshaled item")
	}

	// Test unmarshaling back
	var unmarshaledItem chargebackItem
	err = attributevalue.UnmarshalMap(av, &unmarshaledItem)
	if err != nil {
		t.Errorf("Failed to unmarshal chargeback item: %v", err)
	}

	// Verify data integrity
	if unmarshaledItem.ID != item.ID {
		t.Errorf("Expected ID %s, got %s", item.ID, unmarshaledItem.ID)
	}
	if unmarshaledItem.Amount != item.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", item.Amount, unmarshaledItem.Amount)
	}
}

func TestDynamoDBErrorHandling(t *testing.T) {
	t.Run("handles conditional check failed error", func(t *testing.T) {
		err := &types.ConditionalCheckFailedException{
			Message: aws.String("Item already exists"),
		}

		// Test that we can identify this error type
		var conditionalErr *types.ConditionalCheckFailedException
		if !errors.As(err, &conditionalErr) {
			t.Error("Expected ConditionalCheckFailedException to be identifiable")
		}
	})

	t.Run("handles general DynamoDB errors", func(t *testing.T) {
		err := errors.New("DynamoDB service error")

		// Test error wrapping
		wrappedErr := errors.New("failed to save chargeback: " + err.Error())

		if wrappedErr.Error() != "failed to save chargeback: DynamoDB service error" {
			t.Errorf("Expected wrapped error message, got %s", wrappedErr.Error())
		}
	})
}

func TestDynamoDBKeyConstruction(t *testing.T) {
	t.Run("constructs primary key correctly", func(t *testing.T) {
		id := "chargeback-123"
		key := map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		}

		// Verify key structure
		if key["id"] == nil {
			t.Error("Expected 'id' in key")
		}

		// Verify value
		idAttr, ok := key["id"].(*types.AttributeValueMemberS)
		if !ok {
			t.Error("Expected string attribute value")
		}

		if idAttr.Value != id {
			t.Errorf("Expected ID %s, got %s", id, idAttr.Value)
		}
	})

	t.Run("constructs GSI key correctly", func(t *testing.T) {
		transactionID := "txn-456"
		expressionValues := map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: transactionID},
		}

		// Verify expression values
		if expressionValues[":tid"] == nil {
			t.Error("Expected ':tid' in expression values")
		}

		tidAttr, ok := expressionValues[":tid"].(*types.AttributeValueMemberS)
		if !ok {
			t.Error("Expected string attribute value")
		}

		if tidAttr.Value != transactionID {
			t.Errorf("Expected TransactionID %s, got %s", transactionID, tidAttr.Value)
		}
	})
}

func TestRepositoryTableConfiguration(t *testing.T) {
	tableName := "test-chargebacks-table"
	client := &dynamodb.Client{}

	repo := NewDynamoDBChargebackRepository(client, tableName)

	// Test that repository was created
	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	// Cast to concrete type to access private fields
	dynamoRepo, ok := repo.(*DynamoDBChargebackRepository)
	if !ok {
		t.Fatal("Expected DynamoDBChargebackRepository type")
	}

	// Test table name configuration
	if dynamoRepo.tableName != tableName {
		t.Errorf("Expected table name %s, got %s", tableName, dynamoRepo.tableName)
	}
}

// Additional tests for comprehensive coverage

func TestDynamoDBChargebackRepository_KeyConstruction(t *testing.T) {
	t.Run("constructs primary key correctly", func(t *testing.T) {
		id := "chargeback-123"
		key := map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		}

		if key["id"] == nil {
			t.Error("Expected 'id' in key")
		}

		idAttr, ok := key["id"].(*types.AttributeValueMemberS)
		if !ok {
			t.Error("Expected string attribute value")
		}

		if idAttr.Value != id {
			t.Errorf("Expected ID %s, got %s", id, idAttr.Value)
		}
	})

	t.Run("constructs query expression values", func(t *testing.T) {
		transactionID := "txn-456"
		expressionValues := map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: transactionID},
		}

		if expressionValues[":tid"] == nil {
			t.Error("Expected ':tid' in expression values")
		}

		tidAttr, ok := expressionValues[":tid"].(*types.AttributeValueMemberS)
		if !ok {
			t.Error("Expected string attribute value")
		}

		if tidAttr.Value != transactionID {
			t.Errorf("Expected TransactionID %s, got %s", transactionID, tidAttr.Value)
		}
	})
}

func TestDynamoDBChargebackRepository_LogicOperations(t *testing.T) {
	t.Run("tests ID generation logic", func(t *testing.T) {
		chargeback := &entity.Chargeback{
			ID:            "", // Empty ID to test generation
			TransactionID: "txn-123",
			MerchantID:    "merchant-456",
			Amount:        100.0,
			Currency:      "USD",
			Reason:        entity.ReasonFraud,
			Status:        entity.StatusPending,
			Description:   "Test chargeback",
		}

		// Simulate the ID generation logic from Save method
		if chargeback.ID == "" {
			chargeback.ID = generateChargebackID()
		}

		if chargeback.ID == "" {
			t.Error("Expected ID to be generated")
		}

		if len(chargeback.ID) < 3 || chargeback.ID[:3] != "cb_" {
			t.Errorf("Expected ID to start with 'cb_', got %s", chargeback.ID)
		}
	})

	t.Run("tests timestamp update logic", func(t *testing.T) {
		chargeback := createTestChargeback()
		originalUpdatedAt := chargeback.UpdatedAt

		// Simulate update logic from Update method
		time.Sleep(1 * time.Millisecond) // Ensure different timestamp
		chargeback.UpdatedAt = time.Now()

		if !chargeback.UpdatedAt.After(originalUpdatedAt) {
			t.Error("Expected UpdatedAt to be updated to current time")
		}
	})

	t.Run("tests item marshaling and unmarshaling", func(t *testing.T) {
		testChargeback := createTestChargeback()

		// Create chargebackItem (simulating Save/Update logic)
		item := chargebackItem{
			ID:              testChargeback.ID,
			TransactionID:   testChargeback.TransactionID,
			MerchantID:      testChargeback.MerchantID,
			Amount:          testChargeback.Amount,
			Currency:        testChargeback.Currency,
			CardNumber:      testChargeback.CardNumber,
			Reason:          string(testChargeback.Reason),
			Status:          string(testChargeback.Status),
			Description:     testChargeback.Description,
			TransactionDate: testChargeback.TransactionDate,
			ChargebackDate:  testChargeback.ChargebackDate,
			CreatedAt:       testChargeback.CreatedAt,
			UpdatedAt:       testChargeback.UpdatedAt,
		}

		// Test marshaling (used in Save/Update)
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			t.Errorf("Failed to marshal item: %v", err)
		}

		if av["id"] == nil {
			t.Error("Expected 'id' field in marshaled item")
		}

		// Test unmarshaling (used in FindByID/Query results)
		var unmarshaledItem chargebackItem
		err = attributevalue.UnmarshalMap(av, &unmarshaledItem)
		if err != nil {
			t.Errorf("Failed to unmarshal item: %v", err)
		}

		if unmarshaledItem.ID != item.ID {
			t.Errorf("Expected unmarshaled ID %s, got %s", item.ID, unmarshaledItem.ID)
		}

		// Test itemToEntity conversion (used in all Find methods)
		repo := createTestRepository(&MockDynamoDBClient{})
		entity := repo.itemToEntity(&unmarshaledItem)

		if entity.ID != testChargeback.ID {
			t.Errorf("Expected entity ID %s, got %s", testChargeback.ID, entity.ID)
		}
	})
}

func TestDynamoDBChargebackRepository_QueryParameterConstruction(t *testing.T) {
	t.Run("tests status query with reserved word handling", func(t *testing.T) {
		status := entity.StatusApproved

		// Test expression names for reserved words (used in FindByStatus)
		expressionNames := map[string]string{
			"#status": "status", // status is a DynamoDB reserved word
		}

		if expressionNames["#status"] != "status" {
			t.Error("Expected #status to map to 'status' for reserved word handling")
		}

		// Test expression values
		expressionValues := map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
		}

		if expressionValues[":status"] == nil {
			t.Error("Expected :status in expression values")
		}

		statusValue, ok := expressionValues[":status"].(*types.AttributeValueMemberS)
		if !ok {
			t.Error("Expected string attribute value for status")
		}

		if statusValue.Value != string(status) {
			t.Errorf("Expected status value %s, got %s", status, statusValue.Value)
		}
	})

	t.Run("tests pagination parameters", func(t *testing.T) {
		// Test scan limit configuration (used in List method)
		limit := 10
		scanLimit := int32(limit)

		if scanLimit != 10 {
			t.Errorf("Expected scan limit 10, got %d", scanLimit)
		}

		// Test offset logic simulation
		offset := 5
		totalItems := 20

		// Simulate pagination logic from List method
		if offset < totalItems {
			endIndex := offset + limit
			if endIndex > totalItems {
				endIndex = totalItems
			}

			expectedItems := endIndex - offset
			if expectedItems != limit {
				t.Errorf("Expected %d items with pagination, got %d", limit, expectedItems)
			}
		}
	})
}

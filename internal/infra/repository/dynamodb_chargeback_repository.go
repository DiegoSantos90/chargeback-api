package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/DiegoSantos90/chargeback-api/internal/domain/entity"
	"github.com/DiegoSantos90/chargeback-api/internal/domain/repository"
)

// DynamoDBChargebackRepository implements ChargebackRepository using DynamoDB
type DynamoDBChargebackRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBChargebackRepository creates a new DynamoDB chargeback repository
func NewDynamoDBChargebackRepository(client *dynamodb.Client, tableName string) repository.ChargebackRepository {
	return &DynamoDBChargebackRepository{
		client:    client,
		tableName: tableName,
	}
}

// chargebackItem represents the DynamoDB item structure
type chargebackItem struct {
	ID              string    `dynamodbav:"id"`
	TransactionID   string    `dynamodbav:"transaction_id"`
	MerchantID      string    `dynamodbav:"merchant_id"`
	Amount          float64   `dynamodbav:"amount"`
	Currency        string    `dynamodbav:"currency"`
	CardNumber      string    `dynamodbav:"card_number"`
	Reason          string    `dynamodbav:"reason"`
	Status          string    `dynamodbav:"status"`
	Description     string    `dynamodbav:"description"`
	TransactionDate time.Time `dynamodbav:"transaction_date"`
	ChargebackDate  time.Time `dynamodbav:"chargeback_date"`
	CreatedAt       time.Time `dynamodbav:"created_at"`
	UpdatedAt       time.Time `dynamodbav:"updated_at"`
}

// Save persists a new chargeback to DynamoDB
func (r *DynamoDBChargebackRepository) Save(ctx context.Context, chargeback *entity.Chargeback) error {
	// Generate ID if not present
	if chargeback.ID == "" {
		chargeback.ID = generateChargebackID()
	}

	item := chargebackItem{
		ID:              chargeback.ID,
		TransactionID:   chargeback.TransactionID,
		MerchantID:      chargeback.MerchantID,
		Amount:          chargeback.Amount,
		Currency:        chargeback.Currency,
		CardNumber:      chargeback.CardNumber,
		Reason:          string(chargeback.Reason),
		Status:          string(chargeback.Status),
		Description:     chargeback.Description,
		TransactionDate: chargeback.TransactionDate,
		ChargebackDate:  chargeback.ChargebackDate,
		CreatedAt:       chargeback.CreatedAt,
		UpdatedAt:       chargeback.UpdatedAt,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal chargeback: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
		// Condition to prevent overwriting existing items
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})

	if err != nil {
		return fmt.Errorf("failed to save chargeback: %w", err)
	}

	return nil
}

// FindByID retrieves a chargeback by its unique identifier
func (r *DynamoDBChargebackRepository) FindByID(ctx context.Context, id string) (*entity.Chargeback, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get chargeback: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Not found
	}

	var item chargebackItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chargeback: %w", err)
	}

	return r.itemToEntity(&item), nil
}

// FindByTransactionID retrieves a chargeback by transaction ID
func (r *DynamoDBChargebackRepository) FindByTransactionID(ctx context.Context, transactionID string) (*entity.Chargeback, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("transaction-id-index"), // GSI on transaction_id
		KeyConditionExpression: aws.String("transaction_id = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: transactionID},
		},
		Limit: aws.Int32(1), // We expect only one chargeback per transaction
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query chargeback by transaction ID: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil // Not found
	}

	var item chargebackItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chargeback: %w", err)
	}

	return r.itemToEntity(&item), nil
}

// FindByMerchantID retrieves all chargebacks for a specific merchant
func (r *DynamoDBChargebackRepository) FindByMerchantID(ctx context.Context, merchantID string) ([]*entity.Chargeback, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("merchant-id-index"), // GSI on merchant_id
		KeyConditionExpression: aws.String("merchant_id = :mid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":mid": &types.AttributeValueMemberS{Value: merchantID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query chargebacks by merchant ID: %w", err)
	}

	chargebacks := make([]*entity.Chargeback, 0, len(result.Items))
	for _, item := range result.Items {
		var chargebackItem chargebackItem
		if err := attributevalue.UnmarshalMap(item, &chargebackItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chargeback: %w", err)
		}
		chargebacks = append(chargebacks, r.itemToEntity(&chargebackItem))
	}

	return chargebacks, nil
}

// Update updates an existing chargeback in DynamoDB
func (r *DynamoDBChargebackRepository) Update(ctx context.Context, chargeback *entity.Chargeback) error {
	chargeback.UpdatedAt = time.Now()

	item := chargebackItem{
		ID:              chargeback.ID,
		TransactionID:   chargeback.TransactionID,
		MerchantID:      chargeback.MerchantID,
		Amount:          chargeback.Amount,
		Currency:        chargeback.Currency,
		CardNumber:      chargeback.CardNumber,
		Reason:          string(chargeback.Reason),
		Status:          string(chargeback.Status),
		Description:     chargeback.Description,
		TransactionDate: chargeback.TransactionDate,
		ChargebackDate:  chargeback.ChargebackDate,
		CreatedAt:       chargeback.CreatedAt,
		UpdatedAt:       chargeback.UpdatedAt,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal chargeback: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
		// Condition to ensure the item exists
		ConditionExpression: aws.String("attribute_exists(id)"),
	})

	if err != nil {
		return fmt.Errorf("failed to update chargeback: %w", err)
	}

	return nil
}

// Delete removes a chargeback from DynamoDB
func (r *DynamoDBChargebackRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		// Condition to ensure the item exists
		ConditionExpression: aws.String("attribute_exists(id)"),
	})

	if err != nil {
		return fmt.Errorf("failed to delete chargeback: %w", err)
	}

	return nil
}

// FindByStatus retrieves chargebacks by their status
func (r *DynamoDBChargebackRepository) FindByStatus(ctx context.Context, status entity.ChargebackStatus) ([]*entity.Chargeback, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("status-index"), // GSI on status
		KeyConditionExpression: aws.String("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status", // status is a reserved word
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query chargebacks by status: %w", err)
	}

	chargebacks := make([]*entity.Chargeback, 0, len(result.Items))
	for _, item := range result.Items {
		var chargebackItem chargebackItem
		if err := attributevalue.UnmarshalMap(item, &chargebackItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chargeback: %w", err)
		}
		chargebacks = append(chargebacks, r.itemToEntity(&chargebackItem))
	}

	return chargebacks, nil
}

// List retrieves chargebacks with pagination support
func (r *DynamoDBChargebackRepository) List(ctx context.Context, offset, limit int) ([]*entity.Chargeback, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Limit:     aws.Int32(int32(limit)),
	}

	// For offset, we need to scan and skip items (not efficient for large offsets)
	// In production, consider using pagination tokens instead
	if offset > 0 {
		// This is a simplified implementation
		// For better performance, implement cursor-based pagination
		var scannedItems []map[string]types.AttributeValue
		var lastEvaluatedKey map[string]types.AttributeValue

		for len(scannedItems) < offset+limit {
			if lastEvaluatedKey != nil {
				input.ExclusiveStartKey = lastEvaluatedKey
			}

			result, err := r.client.Scan(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to scan chargebacks: %w", err)
			}

			scannedItems = append(scannedItems, result.Items...)
			lastEvaluatedKey = result.LastEvaluatedKey

			if lastEvaluatedKey == nil {
				break // No more items
			}
		}

		// Take only the items we need
		if offset >= len(scannedItems) {
			return []*entity.Chargeback{}, nil
		}

		endIndex := offset + limit
		if endIndex > len(scannedItems) {
			endIndex = len(scannedItems)
		}

		items := scannedItems[offset:endIndex]
		chargebacks := make([]*entity.Chargeback, 0, len(items))

		for _, item := range items {
			var chargebackItem chargebackItem
			if err := attributevalue.UnmarshalMap(item, &chargebackItem); err != nil {
				return nil, fmt.Errorf("failed to unmarshal chargeback: %w", err)
			}
			chargebacks = append(chargebacks, r.itemToEntity(&chargebackItem))
		}

		return chargebacks, nil
	}

	// Simple case: no offset
	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan chargebacks: %w", err)
	}

	chargebacks := make([]*entity.Chargeback, 0, len(result.Items))
	for _, item := range result.Items {
		var chargebackItem chargebackItem
		if err := attributevalue.UnmarshalMap(item, &chargebackItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chargeback: %w", err)
		}
		chargebacks = append(chargebacks, r.itemToEntity(&chargebackItem))
	}

	return chargebacks, nil
}

// itemToEntity converts a DynamoDB item to a domain entity
func (r *DynamoDBChargebackRepository) itemToEntity(item *chargebackItem) *entity.Chargeback {
	return &entity.Chargeback{
		ID:              item.ID,
		TransactionID:   item.TransactionID,
		MerchantID:      item.MerchantID,
		Amount:          item.Amount,
		Currency:        item.Currency,
		CardNumber:      item.CardNumber,
		Reason:          entity.ChargebackReason(item.Reason),
		Status:          entity.ChargebackStatus(item.Status),
		Description:     item.Description,
		TransactionDate: item.TransactionDate,
		ChargebackDate:  item.ChargebackDate,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

// generateChargebackID generates a unique ID for a chargeback
func generateChargebackID() string {
	return fmt.Sprintf("cb_%d", time.Now().UnixNano())
}

package entity

import (
	"strings"
	"testing"
	"time"
)

func TestCreateChargebackRequest_Validate(t *testing.T) {
	validRequest := CreateChargebackRequest{
		TransactionID:   "txn-12345",
		MerchantID:      "merchant-67890",
		Amount:          99.99,
		Currency:        "USD",
		CardNumber:      "1234567890123456",
		Reason:          ReasonFraud,
		Description:     "Suspicious transaction",
		TransactionDate: time.Now().Add(-24 * time.Hour),
	}

	tests := []struct {
		name      string
		request   CreateChargebackRequest
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid request",
			request:   validRequest,
			shouldErr: false,
		},
		{
			name: "empty transaction ID",
			request: CreateChargebackRequest{
				TransactionID:   "",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "1234567890123456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "transaction ID is required",
		},
		{
			name: "whitespace only transaction ID",
			request: CreateChargebackRequest{
				TransactionID:   "   ",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "1234567890123456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "transaction ID is required",
		},
		{
			name: "empty merchant ID",
			request: CreateChargebackRequest{
				TransactionID:   "txn-12345",
				MerchantID:      "",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "1234567890123456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "merchant ID is required",
		},
		{
			name: "zero amount",
			request: CreateChargebackRequest{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          0,
				Currency:        "USD",
				CardNumber:      "1234567890123456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "amount must be greater than zero",
		},
		{
			name: "negative amount",
			request: CreateChargebackRequest{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          -50.00,
				Currency:        "USD",
				CardNumber:      "1234567890123456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "amount must be greater than zero",
		},
		{
			name: "empty currency",
			request: CreateChargebackRequest{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "",
				CardNumber:      "1234567890123456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "currency is required",
		},
		{
			name: "empty card number",
			request: CreateChargebackRequest{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "card number is required",
		},
		{
			name: "invalid reason",
			request: CreateChargebackRequest{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "1234567890123456",
				Reason:          "invalid_reason",
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			shouldErr: true,
			errMsg:    "invalid chargeback reason",
		},
		{
			name: "zero transaction date",
			request: CreateChargebackRequest{
				TransactionID: "txn-12345",
				MerchantID:    "merchant-67890",
				Amount:        99.99,
				Currency:      "USD",
				CardNumber:    "1234567890123456",
				Reason:        ReasonFraud,
				// TransactionDate not set (zero value)
			},
			shouldErr: true,
			errMsg:    "transaction date is required",
		},
		{
			name: "multiple validation errors",
			request: CreateChargebackRequest{
				TransactionID: "",
				MerchantID:    "",
				Amount:        0,
				Currency:      "",
				CardNumber:    "",
				Reason:        "invalid",
				// TransactionDate not set
			},
			shouldErr: true,
			errMsg:    "validation errors:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestNewChargeback(t *testing.T) {
	validRequest := CreateChargebackRequest{
		TransactionID:   "txn-12345",
		MerchantID:      "merchant-67890",
		Amount:          99.99,
		Currency:        "USD",
		CardNumber:      "1234567890123456",
		Reason:          ReasonFraud,
		Description:     "Suspicious transaction",
		TransactionDate: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	t.Run("creates valid chargeback", func(t *testing.T) {
		chargeback, err := NewChargeback(validRequest)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify all fields are set correctly
		if chargeback.TransactionID != validRequest.TransactionID {
			t.Errorf("Expected TransactionID %s, got %s", validRequest.TransactionID, chargeback.TransactionID)
		}

		if chargeback.MerchantID != validRequest.MerchantID {
			t.Errorf("Expected MerchantID %s, got %s", validRequest.MerchantID, chargeback.MerchantID)
		}

		if chargeback.Amount != validRequest.Amount {
			t.Errorf("Expected Amount %f, got %f", validRequest.Amount, chargeback.Amount)
		}

		if chargeback.Currency != validRequest.Currency {
			t.Errorf("Expected Currency %s, got %s", validRequest.Currency, chargeback.Currency)
		}

		if chargeback.Reason != validRequest.Reason {
			t.Errorf("Expected Reason %s, got %s", validRequest.Reason, chargeback.Reason)
		}

		if chargeback.Description != validRequest.Description {
			t.Errorf("Expected Description %s, got %s", validRequest.Description, chargeback.Description)
		}

		if !chargeback.TransactionDate.Equal(validRequest.TransactionDate) {
			t.Errorf("Expected TransactionDate %v, got %v", validRequest.TransactionDate, chargeback.TransactionDate)
		}

		// Verify defaults
		if chargeback.Status != StatusPending {
			t.Errorf("Expected Status %s, got %s", StatusPending, chargeback.Status)
		}

		// Verify card number is masked
		if !strings.Contains(chargeback.CardNumber, "*") {
			t.Error("Expected card number to be masked")
		}

		if !strings.HasSuffix(chargeback.CardNumber, "3456") {
			t.Errorf("Expected card number to end with '3456', got %s", chargeback.CardNumber)
		}

		// Verify timestamps are set
		if chargeback.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}

		if chargeback.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}

		if chargeback.ChargebackDate.IsZero() {
			t.Error("Expected ChargebackDate to be set")
		}

		// Verify ID is not set (should be set by repository)
		if chargeback.ID != "" {
			t.Error("Expected ID to be empty")
		}
	})

	t.Run("fails with invalid request", func(t *testing.T) {
		invalidRequest := CreateChargebackRequest{
			// Missing required fields
		}

		chargeback, err := NewChargeback(invalidRequest)

		if err == nil {
			t.Error("Expected error but got none")
		}

		if chargeback != nil {
			t.Error("Expected chargeback to be nil when error occurs")
		}
	})
}

func TestChargeback_Approve(t *testing.T) {
	chargeback := &Chargeback{
		Status:    StatusPending,
		UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	t.Run("approves pending chargeback", func(t *testing.T) {
		err := chargeback.Approve()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if chargeback.Status != StatusApproved {
			t.Errorf("Expected Status %s, got %s", StatusApproved, chargeback.Status)
		}

		// Verify UpdatedAt was changed
		if chargeback.UpdatedAt.Equal(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) {
			t.Error("Expected UpdatedAt to be updated")
		}
	})

	t.Run("fails to approve already approved chargeback", func(t *testing.T) {
		approvedChargeback := &Chargeback{Status: StatusApproved}
		err := approvedChargeback.Approve()

		if err == nil {
			t.Error("Expected error but got none")
		}

		if !strings.Contains(err.Error(), "only pending chargebacks can be approved") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("fails to approve rejected chargeback", func(t *testing.T) {
		rejectedChargeback := &Chargeback{Status: StatusRejected}
		err := rejectedChargeback.Approve()

		if err == nil {
			t.Error("Expected error but got none")
		}

		if !strings.Contains(err.Error(), "only pending chargebacks can be approved") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}

func TestChargeback_Reject(t *testing.T) {
	chargeback := &Chargeback{
		Status:    StatusPending,
		UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	t.Run("rejects pending chargeback", func(t *testing.T) {
		err := chargeback.Reject()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if chargeback.Status != StatusRejected {
			t.Errorf("Expected Status %s, got %s", StatusRejected, chargeback.Status)
		}

		// Verify UpdatedAt was changed
		if chargeback.UpdatedAt.Equal(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)) {
			t.Error("Expected UpdatedAt to be updated")
		}
	})

	t.Run("fails to reject already approved chargeback", func(t *testing.T) {
		approvedChargeback := &Chargeback{Status: StatusApproved}
		err := approvedChargeback.Reject()

		if err == nil {
			t.Error("Expected error but got none")
		}

		if !strings.Contains(err.Error(), "only pending chargebacks can be rejected") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("fails to reject already rejected chargeback", func(t *testing.T) {
		rejectedChargeback := &Chargeback{Status: StatusRejected}
		err := rejectedChargeback.Reject()

		if err == nil {
			t.Error("Expected error but got none")
		}

		if !strings.Contains(err.Error(), "only pending chargebacks can be rejected") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}

func TestChargeback_IsValid(t *testing.T) {
	validChargeback := &Chargeback{
		TransactionID:   "txn-12345",
		MerchantID:      "merchant-67890",
		Amount:          99.99,
		Currency:        "USD",
		CardNumber:      "****3456",
		Reason:          ReasonFraud,
		TransactionDate: time.Now().Add(-24 * time.Hour),
		ChargebackDate:  time.Now(),
	}

	tests := []struct {
		name       string
		chargeback *Chargeback
		expected   bool
	}{
		{
			name:       "valid chargeback",
			chargeback: validChargeback,
			expected:   true,
		},
		{
			name: "missing transaction ID",
			chargeback: &Chargeback{
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "****3456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
				ChargebackDate:  time.Now(),
			},
			expected: false,
		},
		{
			name: "missing merchant ID",
			chargeback: &Chargeback{
				TransactionID:   "txn-12345",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "****3456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
				ChargebackDate:  time.Now(),
			},
			expected: false,
		},
		{
			name: "zero amount",
			chargeback: &Chargeback{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          0,
				Currency:        "USD",
				CardNumber:      "****3456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
				ChargebackDate:  time.Now(),
			},
			expected: false,
		},
		{
			name: "missing currency",
			chargeback: &Chargeback{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				CardNumber:      "****3456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
				ChargebackDate:  time.Now(),
			},
			expected: false,
		},
		{
			name: "missing card number",
			chargeback: &Chargeback{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
				ChargebackDate:  time.Now(),
			},
			expected: false,
		},
		{
			name: "missing reason",
			chargeback: &Chargeback{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "****3456",
				TransactionDate: time.Now().Add(-24 * time.Hour),
				ChargebackDate:  time.Now(),
			},
			expected: false,
		},
		{
			name: "zero transaction date",
			chargeback: &Chargeback{
				TransactionID:  "txn-12345",
				MerchantID:     "merchant-67890",
				Amount:         99.99,
				Currency:       "USD",
				CardNumber:     "****3456",
				Reason:         ReasonFraud,
				ChargebackDate: time.Now(),
			},
			expected: false,
		},
		{
			name: "zero chargeback date",
			chargeback: &Chargeback{
				TransactionID:   "txn-12345",
				MerchantID:      "merchant-67890",
				Amount:          99.99,
				Currency:        "USD",
				CardNumber:      "****3456",
				Reason:          ReasonFraud,
				TransactionDate: time.Now().Add(-24 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chargeback.IsValid()

			if result != tt.expected {
				t.Errorf("Expected IsValid() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsValidReason(t *testing.T) {
	tests := []struct {
		name     string
		reason   ChargebackReason
		expected bool
	}{
		{
			name:     "fraud reason",
			reason:   ReasonFraud,
			expected: true,
		},
		{
			name:     "authorization error reason",
			reason:   ReasonAuthorizationError,
			expected: true,
		},
		{
			name:     "processing error reason",
			reason:   ReasonProcessingError,
			expected: true,
		},
		{
			name:     "consumer dispute reason",
			reason:   ReasonConsumerDispute,
			expected: true,
		},
		{
			name:     "invalid reason",
			reason:   "invalid_reason",
			expected: false,
		},
		{
			name:     "empty reason",
			reason:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidReason(tt.reason)

			if result != tt.expected {
				t.Errorf("Expected isValidReason(%s) to return %v, got %v", tt.reason, tt.expected, result)
			}
		})
	}
}

func TestMaskCardNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "16 digit card number",
			input:    "1234567890123456",
			expected: "************3456",
		},
		{
			name:     "15 digit card number",
			input:    "123456789012345",
			expected: "***********2345",
		},
		{
			name:     "card number with spaces",
			input:    "1234 5678 9012 3456",
			expected: "************3456",
		},
		{
			name:     "card number with dashes",
			input:    "1234-5678-9012-3456",
			expected: "************3456",
		},
		{
			name:     "card number with mixed separators",
			input:    "1234 5678-9012 3456",
			expected: "************3456",
		},
		{
			name:     "short card number",
			input:    "123",
			expected: "****",
		},
		{
			name:     "exactly 4 digits",
			input:    "1234",
			expected: "1234",
		},
		{
			name:     "empty card number",
			input:    "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskCardNumber(tt.input)

			if result != tt.expected {
				t.Errorf("Expected maskCardNumber(%s) to return %s, got %s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestChargebackReasonConstants(t *testing.T) {
	// Test that all reason constants are properly defined
	expectedReasons := map[ChargebackReason]string{
		ReasonFraud:              "fraud",
		ReasonAuthorizationError: "authorization_error",
		ReasonProcessingError:    "processing_error",
		ReasonConsumerDispute:    "consumer_dispute",
	}

	for reason, expectedValue := range expectedReasons {
		if string(reason) != expectedValue {
			t.Errorf("Expected reason %s to have value %s, got %s", reason, expectedValue, string(reason))
		}
	}
}

func TestChargebackStatusConstants(t *testing.T) {
	// Test that all status constants are properly defined
	expectedStatuses := map[ChargebackStatus]string{
		StatusPending:  "pending",
		StatusApproved: "approved",
		StatusRejected: "rejected",
	}

	for status, expectedValue := range expectedStatuses {
		if string(status) != expectedValue {
			t.Errorf("Expected status %s to have value %s, got %s", status, expectedValue, string(status))
		}
	}
}

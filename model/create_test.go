package model

import (
	"context"
	"errors"
	"github.com/Ilios-LLC/magicmodel-go/model/mocks"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"strings"
	"testing"
)

// TestUser is a test struct that embeds the Model
type TestUser struct {
	Model
	Name    string
	Email   string
	Age     int
	IsAdmin bool
}

func TestOperator_Create(t *testing.T) {
	// Define test cases
	tests := []struct {
		name          string
		setupMock     func(mock *mocks.MockDynamoDBAPI)
		input         *TestUser
		presetID      string
		expectError   bool
		errorContains string
	}{
		{
			name: "Success - Create new item",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.PutItemFunc = func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
					return &dynamodb.PutItemOutput{}, nil
				}
			},
			input: &TestUser{
				Name:    "John Doe",
				Email:   "john@example.com",
				Age:     30,
				IsAdmin: true,
			},
			expectError: false,
		},
		{
			name: "Error - Item already exists",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				// No need to set up mock as we won't reach the PutItem call
			},
			input: &TestUser{
				Name:    "John Doe",
				Email:   "john@example.com",
				Age:     30,
				IsAdmin: true,
			},
			presetID:      "existing-id",
			expectError:   true,
			errorContains: "item already exists",
		},
		{
			name: "Error - DynamoDB PutItem fails",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.PutItemFunc = func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
					return nil, errors.New("dynamodb error")
				}
			},
			input: &TestUser{
				Name:    "John Doe",
				Email:   "john@example.com",
				Age:     30,
				IsAdmin: true,
			},
			expectError:   true,
			errorContains: "dynamodb error",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock
			mockDB := &mocks.MockDynamoDBAPI{}
			if tc.setupMock != nil {
				tc.setupMock(mockDB)
			}

			// Create operator with mock
			op := NewMagicModelOperatorWithClient(mockDB, "test-table")

			// Set ID if needed for the test
			if tc.presetID != "" {
				tc.input.ID = tc.presetID
			}

			// Call the Create method
			result := op.Create(tc.input)

			// Check for errors
			if tc.expectError {
				if result.Err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tc.errorContains != "" && !containsString(result.Err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, result.Err.Error())
				}
			} else {
				if result.Err != nil {
					t.Errorf("Expected no error but got: %v", result.Err)
					return
				}

				// Verify fields were set correctly
				if tc.input.ID == "" {
					t.Errorf("ID was not set")
				}
				if tc.input.Type != "test_user" {
					t.Errorf("Type was not set correctly, got: %s", tc.input.Type)
				}
				if tc.input.CreatedAt.IsZero() {
					t.Errorf("CreatedAt was not set")
				}
				if tc.input.UpdatedAt.IsZero() {
					t.Errorf("UpdatedAt was not set")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	// Use Go's built-in strings.Contains function
	return strings.Contains(s, substr)
}

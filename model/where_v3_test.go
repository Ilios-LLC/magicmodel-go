package model

import (
	"context"
	"errors"
	"github.com/Ilios-LLC/magicmodel-go/model/mocks"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"testing"
	"time"
)

func TestOperator_WhereV3(t *testing.T) {
	// Create test users
	user1 := &TestUser{
		Model: Model{
			ID:        "1",
			Type:      "test_user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "John Doe",
		Email:   "john@example.com",
		Age:     30,
		IsAdmin: true,
	}

	user2 := &TestUser{
		Model: Model{
			ID:        "2",
			Type:      "test_user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:    "Jane Smith",
		Email:   "jane@example.com",
		Age:     25,
		IsAdmin: false,
	}

	// Define test cases
	tests := []struct {
		name          string
		setupMock     func(mock *mocks.MockDynamoDBAPI)
		input         interface{}
		key           string
		value         interface{}
		isChain       bool
		initialItems  interface{}
		expectError   bool
		errorContains string
		expectedItems int
	}{
		{
			name: "Success - Query from DynamoDB",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.QueryFunc = func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
					items := []map[string]types.AttributeValue{}

					if params.ExpressionAttributeValues != nil {
						val := params.ExpressionAttributeValues[":v"]
						var isAdmin bool
						_ = attributevalue.Unmarshal(val, &isAdmin)

						if isAdmin {
							item1, _ := attributevalue.MarshalMap(user1)
							items = append(items, item1)
						} else {
							item2, _ := attributevalue.MarshalMap(user2)
							items = append(items, item2)
						}
					}

					return &dynamodb.QueryOutput{
						Items: items,
					}, nil
				}
			},
			input:         &[]TestUser{},
			key:           "IsAdmin",
			value:         true,
			isChain:       false,
			expectError:   false,
			expectedItems: 1, // Only user1 is admin
		},
		{
			name: "Success - Filter existing items",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				// No need to set up mock as we won't reach the Query call
			},
			input:         &[]TestUser{*user1, *user2},
			key:           "IsAdmin",
			value:         true,
			isChain:       false,
			expectError:   false,
			expectedItems: 1, // Only user1 is admin
		},
		{
			name: "Success - Filter with nested field",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				// No need to set up mock as we won't reach the Query call
			},
			input:         &[]TestUser{*user1, *user2},
			key:           "Model.ID",
			value:         "1",
			isChain:       false,
			expectError:   false,
			expectedItems: 1, // Only user1 has ID "1"
		},
		{
			name: "Error - DynamoDB Query fails",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.QueryFunc = func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
					return nil, errors.New("dynamodb error")
				}
			},
			input:         &[]TestUser{},
			key:           "IsAdmin",
			value:         true,
			isChain:       false,
			expectError:   true,
			errorContains: "dynamodb error",
		},
		{
			name: "Error - Invalid input",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				// No need to set up mock as we won't reach the Query call
			},
			input:         &TestUser{}, // Not a slice
			key:           "IsAdmin",
			value:         true,
			isChain:       false,
			expectError:   true,
			errorContains: "the WhereV3 operation encountered an error: expected a pointer to a slice, got pointer to struct",
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

			// Call the WhereV3 method
			result := op.WhereV3(tc.isChain, tc.input, tc.key, tc.value)

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

				// Check if the chain state was set correctly
				if result.IsWhereChain != tc.isChain {
					t.Errorf("Expected IsWhereChain to be %v, got %v", tc.isChain, result.IsWhereChain)
				}

				// Check if the items were filtered correctly
				val := reflect.ValueOf(tc.input).Elem()
				if val.Kind() == reflect.Slice {
					if val.Len() != tc.expectedItems {
						t.Errorf("Expected %d items, got %d", tc.expectedItems, val.Len())
					}
				}
			}
		})
	}
}

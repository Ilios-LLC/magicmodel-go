package model

import (
	"context"
	"errors"
	"github.com/Ilios-LLC/magicmodel-go/model/mocks"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"testing"
)

func TestNewMagicModelOperatorWithClient(t *testing.T) {
	// Create mock
	mockDB := &mocks.MockDynamoDBAPI{}
	tableName := "test-table"

	// Create operator with mock
	op := NewMagicModelOperatorWithClient(mockDB, tableName)

	// Check if the operator was created correctly
	if op == nil {
		t.Errorf("Expected operator to be created, got nil")
		return
	}

	if op.db != mockDB {
		t.Errorf("Expected db to be set to mock")
	}

	if op.tableName != tableName {
		t.Errorf("Expected tableName to be '%s', got '%s'", tableName, op.tableName)
	}

	if op.Err != nil {
		t.Errorf("Expected Err to be nil, got %v", op.Err)
	}

	// Check if global variables were set correctly
	if svc != mockDB {
		t.Errorf("Expected global svc to be set to mock")
	}

	if dynamoDBTableName != tableName {
		t.Errorf("Expected global dynamoDBTableName to be '%s', got '%s'", tableName, dynamoDBTableName)
	}
}

func TestOperator_createDynamoDBTable(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock *mocks.MockDynamoDBAPI)
		expectError   bool
		errorContains string
	}{
		{
			name: "Success - Table created",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.CreateTableFunc = func(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
					return &dynamodb.CreateTableOutput{}, nil
				}
				mock.DescribeTableFunc = func(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
					// Include TableStatus field which is needed by the TableExistsWaiter
					return &dynamodb.DescribeTableOutput{
						Table: &types.TableDescription{
							TableStatus: types.TableStatusActive,
						},
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "Success - Table already exists",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.CreateTableFunc = func(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
					return nil, &types.ResourceInUseException{}
				}
			},
			expectError: false,
		},
		{
			name: "Error - CreateTable fails",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.CreateTableFunc = func(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
					return nil, errors.New("create table error")
				}
			},
			expectError:   true,
			errorContains: "create table error",
		},
		{
			name: "Error - DescribeTable fails",
			setupMock: func(mock *mocks.MockDynamoDBAPI) {
				mock.CreateTableFunc = func(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
					return &dynamodb.CreateTableOutput{}, nil
				}
				mock.DescribeTableFunc = func(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
					return nil, errors.New("error while waiting for table to be created: exceeded max wait time for TableExists waiter")
				}
			},
			expectError:   true,
			errorContains: "error while waiting for table to be created: exceeded max wait time for TableExists waiter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock
			mockDB := &mocks.MockDynamoDBAPI{}
			if tc.setupMock != nil {
				tc.setupMock(mockDB)
			}

			// Create operator with mock
			op := NewMagicModelOperatorWithClient(mockDB, "test-table")

			// Call the createDynamoDBTable method
			err := op.createDynamoDBTable(context.Background())

			// Check for errors
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tc.errorContains != "" && !containsString(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

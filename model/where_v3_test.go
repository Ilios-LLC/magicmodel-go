package model

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Ilios-LLC/magicmodel-go/mocks"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOperator_WhereV3(t *testing.T) {
	now := time.Now()

	user1 := &TestUser{
		Model: Model{ID: "1", Type: "test_user", CreatedAt: now, UpdatedAt: now},
		Name:  "John Doe", Email: "john@example.com", Age: 30, IsAdmin: true,
	}
	user2 := &TestUser{
		Model: Model{ID: "2", Type: "test_user", CreatedAt: now, UpdatedAt: now},
		Name:  "Jane Smith", Email: "jane@example.com", Age: 25, IsAdmin: false,
	}

	tests := []struct {
		name          string
		setupMock     func(*mocks.DynamoDBAPI)
		input         interface{}
		key           string
		value         interface{}
		isChain       bool
		expectError   bool
		errorContains string
		expectedItems int
	}{
		{
			name: "success",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1},
				}, nil)
			},
			input:         &[]TestUser{},
			key:           "IsAdmin",
			value:         true,
			expectedItems: 1,
		},
		{
			name:          "success_filter_existing_items",
			setupMock:     nil,
			input:         &[]TestUser{*user1, *user2},
			key:           "IsAdmin",
			value:         true,
			expectedItems: 1,
		},
		{
			name:          "success_filter_nested_field",
			setupMock:     nil,
			input:         &[]TestUser{*user1, *user2},
			key:           "Model.ID",
			value:         "1",
			expectedItems: 1,
		},
		{
			name: "fail_query_fails",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("dynamodb error"))
			},
			input:         &[]TestUser{},
			key:           "IsAdmin",
			value:         true,
			expectError:   true,
			errorContains: "dynamodb error",
		},
		{
			name:          "fail_invalid_input",
			setupMock:     nil,
			input:         &TestUser{},
			key:           "IsAdmin",
			value:         true,
			expectError:   true,
			errorContains: "expected a pointer to a slice",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mocks.NewDynamoDBAPI(t)
			if tc.setupMock != nil {
				tc.setupMock(mockDB)
			}

			op := NewMagicModelOperatorWithClient(mockDB, "test-table")
			result := op.WhereV3(tc.isChain, tc.input, tc.key, tc.value)

			if tc.expectError {
				require.Error(t, result.Err)
				require.Contains(t, result.Err.Error(), tc.errorContains)
				return
			}

			require.NoError(t, result.Err)
			require.Equal(t, tc.isChain, result.IsWhereChain)

			sliceVal := reflect.ValueOf(tc.input).Elem()
			require.Equal(t, tc.expectedItems, sliceVal.Len())
		})
	}
}

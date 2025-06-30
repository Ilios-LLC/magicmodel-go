package model

import (
	"errors"
	"testing"
	"time"

	"github.com/Ilios-LLC/magicmodel-go/mocks"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOperator_WhereV4(t *testing.T) {
	now := time.Now()

	user1 := &TestUser{
		Model: Model{ID: "1", Type: "test_user", CreatedAt: now, UpdatedAt: now},
		Name:  "John Doe", Email: "john@example.com", Age: 30, IsAdmin: true,
	}
	user2 := &TestUser{
		Model: Model{ID: "2", Type: "test_user", CreatedAt: now, UpdatedAt: now},
		Name:  "Jane Smith", Email: "jane@example.com", Age: 25, IsAdmin: false,
	}
	user3 := &TestUser{
		Model: Model{ID: "3", Type: "test_user", CreatedAt: now, UpdatedAt: now},
		Name:  "Bob Johnson", Email: "bob@example.com", Age: 35, IsAdmin: true,
	}

	tests := []struct {
		name          string
		setupMock     func(*mocks.DynamoDBAPI)
		operations    func(*Operator, *[]TestUser) *Operator
		expectError   bool
		errorContains string
		expectedItems int
	}{
		{
			name: "single_condition_single_value",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1},
				}, nil)
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(false, result, "IsAdmin", true)
			},
			expectedItems: 1,
		},
		{
			name: "single_condition_multiple_values_array",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				item3, _ := attributevalue.MarshalMap(user3)
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1, item3},
				}, nil)
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(false, result, "Age", []int{30, 35})
			},
			expectedItems: 2,
		},
		{
			name: "chained_conditions_single_values",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1},
				}, nil)
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(true, result, "IsAdmin", true).
					WhereV4(false, result, "Age", 30)
			},
			expectedItems: 1,
		},
		{
			name: "chained_conditions_mixed_single_and_array",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				item3, _ := attributevalue.MarshalMap(user3)
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1, item3},
				}, nil)
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(true, result, "IsAdmin", true).
					WhereV4(false, result, "Age", []int{30, 35})
			},
			expectedItems: 2,
		},
		{
			name: "three_chained_conditions",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1},
				}, nil)
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(true, result, "IsAdmin", true).
					WhereV4(true, result, "Age", 30).
					WhereV4(false, result, "Name", "John Doe")
			},
			expectedItems: 1,
		},
		{
			name: "multiple_separate_chains",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				item1, _ := attributevalue.MarshalMap(user1)
				item2, _ := attributevalue.MarshalMap(user2)
				// First call should return admin users
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item1},
				}, nil).Once()
				// Second call should return non-admin users
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{item2},
				}, nil).Once()
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				// First chain
				op = op.WhereV4(false, result, "IsAdmin", true)
				if op.Err != nil {
					return op
				}
				// Second chain (should reset state)
				return op.WhereV4(false, result, "IsAdmin", false)
			},
			expectedItems: 1, // Only the last query result should be in the slice
		},
		{
			name: "query_fails",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				dbMock.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("dynamodb error"))
			},
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(false, result, "IsAdmin", true)
			},
			expectError:   true,
			errorContains: "dynamodb error",
		},
		{
			name: "invalid_input_not_slice",
			operations: func(op *Operator, result *[]TestUser) *Operator {
				return op.WhereV4(false, &TestUser{}, "IsAdmin", true)
			},
			expectError:   true,
			errorContains: "expected a pointer to a slice",
		},
		{
			name: "error_propagation",
			operations: func(op *Operator, result *[]TestUser) *Operator {
				op.Err = errors.New("existing error")
				return op.WhereV4(false, result, "IsAdmin", true)
			},
			expectError:   true,
			errorContains: "existing error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mocks.NewDynamoDBAPI(t)
			if tc.setupMock != nil {
				tc.setupMock(mockDB)
			}

			op := NewMagicModelOperatorWithClient(mockDB, "test-table")
			result := []TestUser{}
			finalOp := tc.operations(op, &result)

			if tc.expectError {
				require.Error(t, finalOp.Err)
				require.Contains(t, finalOp.Err.Error(), tc.errorContains)
				return
			}

			require.NoError(t, finalOp.Err)
			require.Equal(t, tc.expectedItems, len(result))
			
			// Verify state is reset after chain completion
			require.False(t, finalOp.IsWhereV4Chain)
			require.Empty(t, finalOp.PendingConditions)
		})
	}
}

func TestNormalizeFieldValues(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
	}{
		{
			name:     "single_string",
			input:    "test",
			expected: []interface{}{"test"},
		},
		{
			name:     "single_int",
			input:    42,
			expected: []interface{}{42},
		},
		{
			name:     "single_bool",
			input:    true,
			expected: []interface{}{true},
		},
		{
			name:     "string_slice",
			input:    []string{"a", "b", "c"},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "int_slice",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "interface_slice",
			input:    []interface{}{"mixed", 42, true},
			expected: []interface{}{"mixed", 42, true},
		},
		{
			name:     "empty_slice",
			input:    []string{},
			expected: []interface{}{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeFieldValues(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildWhereV4Expression(t *testing.T) {
	tests := []struct {
		name       string
		typeName   string
		conditions []WhereV4Condition
		expectErr  bool
	}{
		{
			name:     "single_condition_single_value",
			typeName: "test_user",
			conditions: []WhereV4Condition{
				{FieldName: "IsAdmin", FieldValues: []interface{}{true}},
			},
		},
		{
			name:     "single_condition_multiple_values",
			typeName: "test_user",
			conditions: []WhereV4Condition{
				{FieldName: "Age", FieldValues: []interface{}{30, 35, 40}},
			},
		},
		{
			name:     "multiple_conditions",
			typeName: "test_user",
			conditions: []WhereV4Condition{
				{FieldName: "IsAdmin", FieldValues: []interface{}{true}},
				{FieldName: "Age", FieldValues: []interface{}{30, 35}},
			},
		},
		{
			name:       "empty_conditions",
			typeName:   "test_user",
			conditions: []WhereV4Condition{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := buildWhereV4Expression(tc.typeName, tc.conditions)
			
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, expr.KeyCondition())
			
			// Only check for filter if we have conditions
			if len(tc.conditions) > 0 {
				require.NotNil(t, expr.Filter())
			}
		})
	}
}
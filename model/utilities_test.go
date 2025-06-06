package model

import (
	"errors"
	"github.com/Ilios-LLC/magicmodel-go/mocks"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
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

func TestSetField(t *testing.T) {
	type TestStruct struct {
		Name    string
		Age     int
		IsAdmin bool
	}

	tests := []struct {
		name          string
		item          interface{}
		fieldName     string
		value         interface{}
		expectError   bool
		errorContains string
		expected      TestStruct
	}{
		{
			name:      "set_string_field",
			item:      &TestStruct{},
			fieldName: "Name",
			value:     "John Doe",
			expected:  TestStruct{Name: "John Doe"},
		},
		{
			name:      "set_int_field",
			item:      &TestStruct{},
			fieldName: "Age",
			value:     30,
			expected:  TestStruct{Age: 30},
		},
		{
			name:      "set_bool_field",
			item:      &TestStruct{},
			fieldName: "IsAdmin",
			value:     true,
			expected:  TestStruct{IsAdmin: true},
		},
		{
			name:          "fail_non_pointer_item",
			item:          TestStruct{},
			fieldName:     "Name",
			value:         "John Doe",
			expectError:   true,
			errorContains: "must be a pointer",
		},
		{
			name:          "fail_field_does_not_exist",
			item:          &TestStruct{},
			fieldName:     "NonExistentField",
			value:         "value",
			expectError:   true,
			errorContains: "does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := SetField(tc.item, tc.fieldName, tc.value)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				return
			}

			require.NoError(t, err)

			result := tc.item.(*TestStruct)
			assert.Equal(t, tc.expected, *result)
		})
	}
}

type my_struct struct {
	Model
	Name string
}

func TestParseModelName(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple_named_struct",
			input:    &TestUser{},
			expected: "test_user",
		},
		{
			name:     "unnamed_struct",
			input:    &struct{ Model }{},
			expected: "unnamed_struct",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := ParseModelName(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	type NestedStruct struct {
		Value string
	}

	type TestStruct struct {
		Name   string
		Nested NestedStruct
		Ptr    *NestedStruct
	}

	testStruct := TestStruct{
		Name: "Test",
		Nested: NestedStruct{
			Value: "NestedValue",
		},
		Ptr: &NestedStruct{
			Value: "PtrValue",
		},
	}

	tests := []struct {
		name          string
		value         reflect.Value
		fieldPath     string
		expectedValue interface{}
		expectedFound bool
	}{
		{
			name:          "simple_field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Name",
			expectedValue: "Test",
			expectedFound: true,
		},
		{
			name:          "nested_field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Nested.Value",
			expectedValue: "NestedValue",
			expectedFound: true,
		},
		{
			name:          "pointer_field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Ptr.Value",
			expectedValue: "PtrValue",
			expectedFound: true,
		},
		{
			name:          "non_existent_field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "NonExistent",
			expectedFound: false,
		},
		{
			name:          "non_existent_nested_field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Name.NonExistent",
			expectedFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, found := GetFieldValue(tc.value, tc.fieldPath)

			assert.Equal(t, tc.expectedFound, found, "found mismatch")

			if found {
				require.NotNil(t, val)
				assert.Equal(t, tc.expectedValue, val.Interface())
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		operation     string
		structName    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid",
			input:       &TestUser{},
			operation:   "Test",
			structName:  "TestUser",
			expectError: false,
		},
		{
			name:          "fail_non_pointer_input",
			input:         TestUser{},
			operation:     "Test",
			structName:    "TestUser",
			expectError:   true,
			errorContains: "expected a non-nil pointer",
		},
		{
			name:          "fail_nil_input",
			input:         nil,
			operation:     "Test",
			structName:    "TestUser",
			expectError:   true,
			errorContains: "expected a non-nil pointer",
		},
		{
			name:          "fail_pointer_to_non_struct",
			input:         new(string),
			operation:     "Test",
			structName:    "string",
			expectError:   true,
			errorContains: "expected a pointer to a struct",
		},
		{
			name: "fail_struct_without_model",
			input: &struct {
				Name string
			}{},
			operation:     "Test",
			structName:    "struct",
			expectError:   true,
			errorContains: "must embed model.Model",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateInput(tc.input, tc.operation, tc.structName)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateInputSlice(t *testing.T) {
	type NotAModel struct {
		Name string
	}

	type ValidModel struct {
		Model
		Name string
	}

	tests := []struct {
		name          string
		input         interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:          "non_pointer_input",
			input:         []ValidModel{},
			expectError:   true,
			errorContains: "non-nil pointer",
		},
		{
			name:          "pointer_to_non_slice",
			input:         new(string),
			expectError:   true,
			errorContains: "pointer to a slice",
		},
		{
			name:          "slice_of_non_structs",
			input:         &[]string{"bad"},
			expectError:   true,
			errorContains: "elements must be structs",
		},
		{
			name:          "structs_without_model",
			input:         &[]NotAModel{{Name: "Nope"}},
			expectError:   true,
			errorContains: "must embed model.Model",
		},
		{
			name:        "valid_struct_with_model",
			input:       &[]ValidModel{{Name: "Yes"}},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInputSlice(tc.input, "TestOp", "Whatever")
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	truePtr := new(bool)
	*truePtr = true

	tests := []struct {
		name     string
		value    reflect.Value
		compare  interface{}
		expected bool
	}{
		{"bool_pointer_true", reflect.ValueOf(truePtr), true, true},
		{"direct_equal", reflect.ValueOf("hello"), "hello", true},
		{"different_values", reflect.ValueOf(123), 456, false},
		{"fallback_string_match", reflect.ValueOf(123), "123", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compareValues(tc.value, tc.compare)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFilterSliceByField(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Entry struct {
		Model
		Name  string
		Inner Inner
	}

	entries := []Entry{
		{Name: "Alpha", Inner: Inner{Value: "One"}},
		{Name: "Beta", Inner: Inner{Value: "Two"}},
		{Name: "Gamma", Inner: Inner{Value: "One"}},
	}

	refVal := reflect.ValueOf(entries)

	tests := []struct {
		name       string
		field      string
		value      interface{}
		wantLength int
	}{
		{"direct_field_match", "Name", "Beta", 1},
		{"nested_field_match", "Inner.Value", "One", 2},
		{"no_match", "Name", "Zeta", 0},
		{"invalid_field", "Unknown", "X", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := filterSliceByField(refVal, tc.field, tc.value)
			assert.Equal(t, tc.wantLength, result.Len())
		})
	}
}

func TestBuildWhereExpression(t *testing.T) {
	expr, err := buildWhereExpression("test_user", "Name", "John")
	require.NoError(t, err)

	// Check expression pieces are non-nil
	require.NotNil(t, expr.KeyCondition())
	require.NotNil(t, expr.Filter())

	// Check attribute names mapping
	names := expr.Names()
	assert.Contains(t, names, "#0") // "Type"
	assert.Contains(t, names, "#1") // "DeletedAt"
	assert.Contains(t, names, "#2") // "Name"

	// Check that the placeholders map to the correct field names
	assert.Equal(t, "Type", names["#2"])
	assert.Equal(t, "DeletedAt", names["#1"])
	assert.Equal(t, "Name", names["#0"])

	// Check values map
	values := expr.Values()
	assert.Contains(t, values, ":0") // "test_user"
	assert.Contains(t, values, ":2") // "John"

	expectedVal0, err := attributevalue.Marshal("John")
	require.NoError(t, err)
	assert.Equal(t, expectedVal0, values[":0"])

	expectedVal2, err := attributevalue.Marshal("test_user")
	require.NoError(t, err)
	assert.Equal(t, expectedVal2, values[":2"])

}

func TestExecuteWhereQuery_Error(t *testing.T) {
	mockSvc := mocks.NewDynamoDBAPI(t)
	svc = mockSvc
	dynamoDBTableName = "test-table"

	expr := expression.Expression{
		// mock empty expression parts
	}

	mockSvc.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("query failed"))

	op := &Operator{}
	op.executeWhereQuery(expr, &[]TestUser{})

	assert.Error(t, op.Err)
	assert.Contains(t, op.Err.Error(), "Where operation")
}

package model

import (
	"reflect"
	"testing"
)

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
			name:      "Set string field",
			item:      &TestStruct{},
			fieldName: "Name",
			value:     "John Doe",
			expected:  TestStruct{Name: "John Doe"},
		},
		{
			name:      "Set int field",
			item:      &TestStruct{},
			fieldName: "Age",
			value:     30,
			expected:  TestStruct{Age: 30},
		},
		{
			name:      "Set bool field",
			item:      &TestStruct{},
			fieldName: "IsAdmin",
			value:     true,
			expected:  TestStruct{IsAdmin: true},
		},
		{
			name:          "Error - Non-pointer item",
			item:          TestStruct{},
			fieldName:     "Name",
			value:         "John Doe",
			expectError:   true,
			errorContains: "must be a pointer",
		},
		{
			name:          "Error - Field doesn't exist",
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
					return
				}

				// Check if the field was set correctly
				result := tc.item.(*TestStruct)
				switch tc.fieldName {
				case "Name":
					if result.Name != tc.expected.Name {
						t.Errorf("Expected Name to be '%s', got '%s'", tc.expected.Name, result.Name)
					}
				case "Age":
					if result.Age != tc.expected.Age {
						t.Errorf("Expected Age to be %d, got %d", tc.expected.Age, result.Age)
					}
				case "IsAdmin":
					if result.IsAdmin != tc.expected.IsAdmin {
						t.Errorf("Expected IsAdmin to be %v, got %v", tc.expected.IsAdmin, result.IsAdmin)
					}
				}
			}
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
			name:     "Simple struct",
			input:    &TestUser{},
			expected: "test_user",
		},
		{
			name:     "Unnamed struct",
			input:    &struct{ Model }{},
			expected: "unnamed_struct",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := parseModelName(tc.input)
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
			name:          "Simple field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Name",
			expectedValue: "Test",
			expectedFound: true,
		},
		{
			name:          "Nested field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Nested.Value",
			expectedValue: "NestedValue",
			expectedFound: true,
		},
		{
			name:          "Pointer field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Ptr.Value",
			expectedValue: "PtrValue",
			expectedFound: true,
		},
		{
			name:          "Non-existent field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "NonExistent",
			expectedFound: false,
		},
		{
			name:          "Non-existent nested field",
			value:         reflect.ValueOf(testStruct),
			fieldPath:     "Name.NonExistent",
			expectedFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, found := getFieldValue(tc.value, tc.fieldPath)

			if found != tc.expectedFound {
				t.Errorf("Expected found to be %v, got %v", tc.expectedFound, found)
				return
			}

			if found {
				if value.Interface() != tc.expectedValue {
					t.Errorf("Expected value to be '%v', got '%v'", tc.expectedValue, value.Interface())
				}
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
			name:        "Valid input",
			input:       &TestUser{},
			operation:   "Test",
			structName:  "TestUser",
			expectError: false,
		},
		{
			name:          "Error - Non-pointer input",
			input:         TestUser{},
			operation:     "Test",
			structName:    "TestUser",
			expectError:   true,
			errorContains: "expected a non-nil pointer",
		},
		{
			name:          "Error - Nil pointer",
			input:         nil,
			operation:     "Test",
			structName:    "TestUser",
			expectError:   true,
			errorContains: "expected a non-nil pointer",
		},
		{
			name:          "Error - Pointer to non-struct",
			input:         new(string),
			operation:     "Test",
			structName:    "string",
			expectError:   true,
			errorContains: "expected a pointer to a struct",
		},
		{
			name: "Error - Struct without Model",
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
			err := validateInput(tc.input, tc.operation, tc.structName)

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

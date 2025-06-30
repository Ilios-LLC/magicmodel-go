package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stoewer/go-strcase"
	"reflect"
	"strings"
)

func SetField(item interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(item)

	// Check if item is a pointer
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot assign to the item passed, item must be a pointer in order to assign")
	}

	// Get the element the pointer refers to
	v = v.Elem()

	if !v.CanAddr() {
		return fmt.Errorf("cannot assign to the item passed, item must be a pointer to an addressable value")
	}

	fieldNames := map[string]int{}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		jname := typeField.Name
		fieldNames[jname] = i
	}

	fieldNum, ok := fieldNames[fieldName]
	if !ok {
		return fmt.Errorf("field %s does not exist within the provided item", fieldName)
	}
	fieldVal := v.Field(fieldNum)
	fieldVal.Set(reflect.ValueOf(value))
	return nil
}

func ParseModelName(q interface{}) (string, error) {
	t := reflect.TypeOf(q)

	// Unwrap pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Unwrap slice if needed
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected a struct or slice of structs, got %s", t.Kind())
	}

	if t.Name() == "" {
		return "unnamed_struct", fmt.Errorf("cannot use an unnamed struct")
	}

	return strcase.SnakeCase(t.Name()), nil
}

func GetFieldValue(value reflect.Value, fieldPath string) (reflect.Value, bool) {
	fields := strings.Split(fieldPath, ".")

	for _, field := range fields {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}

		value = value.FieldByName(field)
		if !value.IsValid() {
			return reflect.Value{}, false
		}
	}

	return value, true
}

func ValidateInput(q interface{}, operation, structName string) error {
	val := reflect.ValueOf(q)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("the %s operation encountered an error: expected a non-nil pointer to a struct, got %T", operation, q)
	}

	checkPayload := val.Elem()
	if checkPayload.Kind() != reflect.Struct {
		return fmt.Errorf("the %s operation encountered an error: expected a pointer to a struct, got pointer to %s", operation, checkPayload.Kind())
	}

	modelType := reflect.TypeOf((*Model)(nil)).Elem()
	hasModel := false

	for i := 0; i < checkPayload.NumField(); i++ {
		field := checkPayload.Type().Field(i)
		if field.Anonymous && field.Type == modelType {
			hasModel = true
			break
		}
	}

	if !hasModel {
		return fmt.Errorf(`the %s operation encountered an error: struct %s must embed model.Model (e.g., model.Model `, operation, structName+"`yaml:\",inline\"`"+`)`)
	}
	return nil
}

func validateInputSlice(q interface{}, operation, structName string) error {
	val := reflect.ValueOf(q)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("the %s operation encountered an error: expected a non-nil pointer to a slice, got %T", operation, q)
	}

	sliceVal := val.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("the %s operation encountered an error: expected a pointer to a slice, got pointer to %s", operation, sliceVal.Kind())
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem() // unwrap *T if slice is []*T
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("the %s operation encountered an error: slice elements must be structs, got %s", operation, elemType.Kind())
	}

	modelType := reflect.TypeOf((*Model)(nil)).Elem()
	hasModel := false

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if field.Anonymous && field.Type == modelType {
			hasModel = true
			break
		}
	}

	if !hasModel {
		return fmt.Errorf(
			"the %s operation encountered an error: struct %s must embed model.Model (e.g., model.Model `yaml:\",inline\"`)",
			operation, structName)
	}

	return nil
}

// compareValues compares two values and returns true if they match
// Handles various types of comparisons including pointers, direct equality, and string representation
func compareValues(fieldValue reflect.Value, compareValue interface{}) bool {
	// Handle nil pointer values for booleans
	if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() && fieldValue.Elem().Kind() == reflect.Bool {
		// Compare with the dereferenced value
		return compareValue == fieldValue.Elem().Bool()
	}

	// Try direct comparison
	if reflect.DeepEqual(fieldValue.Interface(), compareValue) {
		return true
	}

	// String comparison as fallback
	fieldStr := fmt.Sprintf("%v", fieldValue.Interface())
	compareStr := fmt.Sprintf("%v", compareValue)
	return fieldStr == compareStr
}

// filterSliceByField filters a slice based on a field name and value
// Returns a new slice containing only the elements that match
func filterSliceByField(slice reflect.Value, fieldName string, compareValue interface{}) reflect.Value {
	// Create a new slice of the same type
	newSlice := reflect.New(slice.Type()).Elem()

	// Iterate through each element
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}

		var fieldValue reflect.Value
		var found bool

		// Handle nested fields vs direct fields
		if strings.Contains(fieldName, ".") {
			fieldValue, found = GetFieldValue(elem, fieldName)
		} else {
			fieldValue = elem.FieldByName(fieldName)
			found = fieldValue.IsValid()
		}

		// Skip if field not found
		if !found {
			continue
		}

		// Add to new slice if values match
		if compareValues(fieldValue, compareValue) {
			newSlice = reflect.Append(newSlice, elem)
		}
	}

	return newSlice
}

// buildWhereExpression builds the DynamoDB expression for a where query
func buildWhereExpression(typeName, fieldName string, fieldValue interface{}) (expression.Expression, error) {
	// Create key condition for the Type
	keyCondition := expression.Key("Type").Equal(expression.Value(typeName))

	// Create filter condition for the field
	fieldCondition := expression.Name(fieldName).Equal(expression.Value(fieldValue))

	// Add soft delete conditions
	softDeleteCond := expression.Not(expression.Name("DeletedAt").AttributeExists())
	softDeleteCond2 := expression.Not(expression.Name("DeletedAt").NotEqual(expression.Value(nil)))

	// Build the complete expression
	return expression.NewBuilder().
		WithKeyCondition(keyCondition).
		WithFilter(fieldCondition.And(softDeleteCond.Or(softDeleteCond2))).
		Build()
}

// buildWhereV4Expression builds a comprehensive DynamoDB expression for multiple where conditions
func buildWhereV4Expression(typeName string, conditions []WhereV4Condition) (expression.Expression, error) {
	// Create key condition for the Type
	keyCondition := expression.Key("Type").Equal(expression.Value(typeName))

	// Add soft delete conditions
	softDeleteCond := expression.Not(expression.Name("DeletedAt").AttributeExists())
	softDeleteCond2 := expression.Not(expression.Name("DeletedAt").NotEqual(expression.Value(nil)))

	// Start with soft delete filter
	finalFilter := softDeleteCond.Or(softDeleteCond2)

	// Build field filter conditions if any exist
	if len(conditions) > 0 {
		var fieldFilterCondition expression.ConditionBuilder

		for i, condition := range conditions {
			var conditionExpr expression.ConditionBuilder

			if len(condition.FieldValues) == 1 {
				// Single value - use equality
				conditionExpr = expression.Name(condition.FieldName).Equal(expression.Value(condition.FieldValues[0]))
			} else {
				// Multiple values - use IN operator
				values := make([]expression.OperandBuilder, len(condition.FieldValues))
				for j, val := range condition.FieldValues {
					values[j] = expression.Value(val)
				}
				conditionExpr = expression.Name(condition.FieldName).In(values[0], values[1:]...)
			}

			if i == 0 {
				fieldFilterCondition = conditionExpr
			} else {
				fieldFilterCondition = fieldFilterCondition.And(conditionExpr)
			}
		}

		// Combine field conditions with soft delete filter
		finalFilter = fieldFilterCondition.And(finalFilter)
	}

	// Build the complete expression
	return expression.NewBuilder().
		WithKeyCondition(keyCondition).
		WithFilter(finalFilter).
		Build()
}

// executeWhereQuery executes a DynamoDB query with the given expression
func (o *Operator) executeWhereQuery(expr expression.Expression, result interface{}) *Operator {
	response, err := svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(dynamoDBTableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", err)
		return o
	}

	err = attributevalue.UnmarshalListOfMaps(response.Items, result)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", err)
	}

	return o
}

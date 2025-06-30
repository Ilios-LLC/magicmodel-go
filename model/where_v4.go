package model

import (
	"fmt"
	"reflect"
)

// WhereV4 filters a collection based on a field name and value(s) with deferred execution
// Supports both single values and arrays for OR conditions within the same field
// Can be chained with other WhereV4 calls when isChain is true
func (o *Operator) WhereV4(isChain bool, q interface{}, fieldName string, fieldValue interface{}) *Operator {
	// Return early if there's already an error
	if o.Err != nil {
		return o
	}

	// Parse model name and validate input
	name, err := ParseModelName(q)
	if err != nil {
		o.Err = err
		return o
	}

	err = validateInputSlice(q, "WhereV4", name)
	if err != nil {
		o.Err = err
		return o
	}

	// Reset state if this is the start of a new chain
	if !o.IsWhereV4Chain {
		o.PendingConditions = nil
	}

	// Convert single values to arrays for consistent handling
	fieldValues := normalizeFieldValues(fieldValue)

	// Add condition to pending list
	o.PendingConditions = append(o.PendingConditions, WhereV4Condition{
		FieldName:   fieldName,
		FieldValues: fieldValues,
	})

	// Set chain state
	o.IsWhereV4Chain = isChain

	// If this is the end of the chain, execute the query
	if !isChain {
		o = o.executeWhereV4Query(name, q)
		// Reset state after execution for next use
		o.PendingConditions = nil
		o.IsWhereV4Chain = false
	}

	return o
}

// normalizeFieldValues converts single values to arrays and validates arrays
func normalizeFieldValues(fieldValue interface{}) []interface{} {
	val := reflect.ValueOf(fieldValue)
	
	// If it's already a slice, convert to []interface{}
	if val.Kind() == reflect.Slice {
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result
	}
	
	// Single value, wrap in slice
	return []interface{}{fieldValue}
}

// executeWhereV4Query builds and executes a single DynamoDB query with all pending conditions
func (o *Operator) executeWhereV4Query(typeName string, result interface{}) *Operator {
	if len(o.PendingConditions) == 0 {
		o.Err = fmt.Errorf("no conditions to execute in WhereV4")
		return o
	}

	// Build the comprehensive expression
	expr, err := buildWhereV4Expression(typeName, o.PendingConditions)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during WhereV4 operation: %v", err)
		return o
	}

	// Execute the query using the existing helper
	return o.executeWhereQuery(expr, result)
}
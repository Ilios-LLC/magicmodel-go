package model

import (
	"fmt"
	"reflect"
)

// WhereV2 filters a collection based on a field name and value
// This is a simpler version of WhereV3 that only supports string comparisons
// Can be chained with other Where calls when isChain is true
func (o *Operator) WhereV2(isChain bool, q interface{}, fieldName string, fieldValue interface{}) *Operator {
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

	err = validateInputSlice(q, "WhereV2", name)
	if err != nil {
		o.Err = err
		return o
	}

	// Get the slice value
	val := reflect.ValueOf(q)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	// Ensure we're working with a slice
	if val.Kind() != reflect.Slice {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", "q is not a slice")
		return o
	}

	// Handle in-memory filtering for non-empty slices
	if val.Len() > 0 {
		// Filter the slice
		filteredSlice := filterSliceByField(val, fieldName, fieldValue)

		// Update the original slice with filtered results
		reflect.ValueOf(q).Elem().Set(filteredSlice)

		// Update chain state
		o.IsWhereChain = isChain
		return o
	}

	// If we're in a chain but have no items to filter, just return
	if o.IsWhereChain {
		o.IsWhereChain = isChain
		return o
	}

	// Build query expression
	expr, err := buildWhereExpression(name, fieldName, fieldValue)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", err)
		return o
	}

	// Execute the query
	o = o.executeWhereQuery(expr, q)

	// Update chain state
	o.IsWhereChain = isChain
	return o
}

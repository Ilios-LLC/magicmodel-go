package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"reflect"
	"strings"
)

func (o *Operator) WhereV3(isChain bool, q interface{}, k string, v interface{}) *Operator {
	if o.Err != nil {
		return o
	}

	name := parseModelName(q)
	err := validateInput(q, "WhereV3", name)
	if err != nil {
		o.Err = err
		return o
	}

	val := reflect.ValueOf(q)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	isSlice := val.Kind() == reflect.Slice

	// If we have items to filter (either from a previous query or a new one)
	if isSlice && val.Len() > 0 {
		newQ := reflect.New(reflect.TypeOf(q).Elem()).Elem()
		for i := 0; i < val.Len(); i++ {
			e := val.Index(i)
			if e.Kind() == reflect.Pointer {
				e = e.Elem()
			}

			var matches bool
			if strings.Contains(k, ".") {
				// Handle nested fields
				value, found := getFieldValue(e, k)
				if !found {
					continue
				}

				// Handle nil pointer values for booleans
				if value.Kind() == reflect.Ptr && !value.IsNil() && value.Elem().Kind() == reflect.Bool {
					// Compare with the dereferenced value
					if v == value.Elem().Bool() {
						matches = true
					}
				} else if reflect.DeepEqual(value.Interface(), v) {
					// Direct comparison
					matches = true
				} else {
					// String comparison as fallback
					fieldValue := fmt.Sprintf("%v", value.Interface())
					vValue := fmt.Sprintf("%v", v)
					if fieldValue == vValue {
						matches = true
					}
				}
			} else {
				// Handle direct fields
				field := e.FieldByName(k)
				if !field.IsValid() {
					continue // Skip if the field doesn't exist
				}

				// Handle nil pointer values for booleans
				if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Bool {
					// Compare with the dereferenced value
					if v == field.Elem().Bool() {
						matches = true
					}
				} else if reflect.DeepEqual(field.Interface(), v) {
					// Direct comparison
					matches = true
				} else {
					// String comparison as fallback
					fieldValue := fmt.Sprintf("%v", field.Interface())
					vValue := fmt.Sprintf("%v", v)
					if fieldValue == vValue {
						matches = true
					}
				}
			}

			if matches {
				newQ = reflect.Append(newQ, e)
			}
		}

		// Set the filtered results
		reflect.ValueOf(q).Elem().Set(newQ)

		// Update chain state based on isChain parameter
		o.IsWhereChain = isChain

		return o
	}

	// If we're in a chain but have no items to filter, just return
	// This ensures that chained calls with no results behave as AND operations
	if o.IsWhereChain {
		// Update chain state based on isChain parameter
		o.IsWhereChain = isChain
		return o
	}

	// If we're not in a chain or have no items, perform a database query
	cond := expression.Key("Type").Equal(expression.Value(name))
	cond2 := expression.Name(k).Equal(expression.Value(v))
	softDeleteCond := expression.Not(expression.Name("DeletedAt").AttributeExists())
	sofDeleteCond2 := expression.Not(expression.Name("DeletedAt").NotEqual(expression.Value(nil)))
	expr, err := expression.NewBuilder().WithKeyCondition(cond).WithFilter(cond2.And(softDeleteCond.Or(sofDeleteCond2))).Build()
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", err)
		return o
	}

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

	err = attributevalue.UnmarshalListOfMaps(response.Items, q)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", err)
		return o
	}

	// Update chain state based on isChain parameter
	o.IsWhereChain = isChain

	return o
}

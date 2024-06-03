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

func (o *Operator) WhereV2(q interface{}, k string, v interface{}) *Operator {
	if o.Err != nil {
		return o
	}
	name := parseModelName(q)
	val := reflect.ValueOf(q)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	isSlice := val.Kind() == reflect.Slice
	if isSlice {
		if val.Len() > 0 {
			newQ := reflect.New(reflect.TypeOf(q).Elem()).Elem()
			for i := 0; i < val.Len(); i++ {
				e := val.Index(i)
				if e.Kind() == reflect.Pointer {
					e = e.Elem()
				}
				if strings.Contains(k, ".") {
					value, found := getFieldValue(e, k)
					if !found {
						continue
					}
					fieldValue := fmt.Sprintf("%v", value.Interface())
					if fieldValue == v.(string) {
						newQ = reflect.Append(newQ, e)
					}
				} else {
					field := e.FieldByName(k)
					if !field.IsValid() {
						continue // Skip if the field doesn't exist
					}
					fieldValue := fmt.Sprintf("%v", field.Interface())
					if fieldValue == v.(string) {
						newQ = reflect.Append(newQ, e)
					}
				}
			}
			reflect.ValueOf(q).Elem().Set(newQ)
			o.isWhereChain = true
			return o
		}
	} else {
		o.Err = fmt.Errorf("encountered an error during Where operation: %v", "q is not a slice")
		return o
	}

	if o.isWhereChain {
		return o
	}
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
	o.isWhereChain = true
	return o
}

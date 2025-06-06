package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
)

func (o *Operator) Update(q interface{}, k string, v interface{}) *Operator {
	if o.Err != nil {
		return o
	}

	name, err := parseModelName(q)
	if err != nil {
		o.Err = err
		return o
	}

	err = validateInput(q, "Update", name)
	if err != nil {
		o.Err = err
		return o
	}

	payload := reflect.ValueOf(q).Elem()
	update := expression.Set(expression.Name(k), expression.Value(v))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Update operation: %v", err)
		return o
	}

	payload.FieldByName(k).Set(reflect.ValueOf(v))

	key := map[string]types.AttributeValue{
		"ID":   &types.AttributeValueMemberS{Value: payload.FieldByName("ID").String()},
		"Type": &types.AttributeValueMemberS{Value: payload.FieldByName("Type").String()},
	}

	_, err = svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 aws.String(dynamoDBTableName),
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Update operation: %v", err)
		return o
	}
	return o
}

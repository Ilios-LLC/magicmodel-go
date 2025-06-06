package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"time"
)

func (o *Operator) SoftDelete(q interface{}) *Operator {
	if o.Err != nil {
		return o
	}

	name := parseModelName(q)
	err := validateInput(q, "SoftDelete", name)
	if err != nil {
		o.Err = err
		return o
	}

	t := time.Now()
	payload := reflect.ValueOf(q).Elem()
	update := expression.Set(expression.Name("DeletedAt"), expression.Value(t))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during SoftDelete operation: %v", err)
		return o
	}

	//payload.FieldByName("DeletedAt").Set(reflect.ValueOf(t))
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
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during SoftDelete operation: %v", err)
		return o
	}
	return o
}

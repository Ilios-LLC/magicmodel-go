package model

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
	"reflect"
)

func (o *Operator) Update(q interface{}, k string, v interface{}) *Operator {
	payload := reflect.ValueOf(q).Elem()
	update := expression.Set(expression.Name(k), expression.Value(v))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		panic(err)
	}

	payload.FieldByName(k).Set(reflect.ValueOf(v))

	key := map[string]types.AttributeValue{
		"Id":   &types.AttributeValueMemberS{Value: payload.FieldByName("ID").String()},
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
		log.Err(err)
	}

	return o
}

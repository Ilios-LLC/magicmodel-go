package model

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
	"reflect"
)

func (o *Operator) Delete(q interface{}) *Operator {
	payload := reflect.ValueOf(q).Elem()
	_, err := svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(dynamoDBTableName), Key: map[string]types.AttributeValue{
			"ID":   &types.AttributeValueMemberS{Value: payload.FieldByName("ID").String()},
			"Type": &types.AttributeValueMemberS{Value: payload.FieldByName("Type").String()},
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete item")
	}
	return o
}

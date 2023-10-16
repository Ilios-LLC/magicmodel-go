package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
)

func (o *Operator) Delete(q interface{}) *Operator {
	if o.Err != nil {
		return o
	}
	payload := reflect.ValueOf(q).Elem()
	_, err := svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(dynamoDBTableName), Key: map[string]types.AttributeValue{
			"ID":   &types.AttributeValueMemberS{Value: payload.FieldByName("ID").String()},
			"Type": &types.AttributeValueMemberS{Value: payload.FieldByName("Type").String()},
		},
	})
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Delete operation: %v", err)
		return o
	}
	return o
}

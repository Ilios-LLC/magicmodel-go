package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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

	result, err := svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 aws.String(dynamoDBTableName),
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	var attributeMap map[string]map[string]interface{}
	err = attributevalue.UnmarshalMap(result.Attributes, &attributeMap)
	if err != nil {
		log.Printf("Couldn't unmarshall update response. Here's why: %v\n", err)
	}

	bs, _ := json.Marshal(attributeMap)

	log.Debug().Str("result", fmt.Sprintf(string(bs))).Msg("UpdateItem result")

	if err != nil {
		log.Err(err)
	}

	return o
}

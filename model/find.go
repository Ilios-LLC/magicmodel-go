package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

func (o *Operator) Find(q interface{}, id string) {
	name := parseModelName(q)
	payload := map[string]types.AttributeValue{
		"Type": &types.AttributeValueMemberS{Value: name},
		"ID":   &types.AttributeValueMemberS{Value: id},
	}

	out, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(dynamoDBTableName),
		Key:       payload,
	})

	if err != nil {
		panic(err)
	}

	if out.Item == nil {
		log.Error().Msg("Item not found")
		return
	}

	err = attributevalue.UnmarshalMap(out.Item, q)
	if err != nil {
		panic(err)
	}

	fmt.Println("thing")
}

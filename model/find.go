package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (o *Operator) Find(q interface{}, id string) *Operator {
	if o.Err != nil {
		return o
	}

	name, err := ParseModelName(q)
	if err != nil {
		o.Err = err
		return o
	}
	err = ValidateInput(q, "Find", name)
	if err != nil {
		o.Err = err
		return o
	}

	payload := map[string]types.AttributeValue{
		"Type": &types.AttributeValueMemberS{Value: name},
		"ID":   &types.AttributeValueMemberS{Value: id},
	}

	out, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(dynamoDBTableName),
		Key:       payload,
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Find operation: %v", err)
		return o
	}

	if out.Item == nil {
		o.Err = fmt.Errorf("encountered an error during Find operation: item not found")
		return o
	}

	err = attributevalue.UnmarshalMap(out.Item, q)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Find operation: %v", err)
		return o
	}
	return o
}

package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog/log"
)

func (o *Operator) All(q interface{}) *Operator {
	name := parseModelName(q)

	cond := expression.Key("Type").Equal(expression.Value(name))
	expr, err := expression.NewBuilder().WithKeyCondition(cond).Build()

	log.Debug().Msg(fmt.Sprintf("%v", expr.Names()))
	log.Debug().Msg(fmt.Sprintf("%v", expr.Values()))
	log.Debug().Msg(fmt.Sprintf("%v", expr.KeyCondition()))

	response, err := svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(dynamoDBTableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		log.Printf("Couldn't query for movies released in %v. Here's why: %v\n", "", err)
		return o
	}

	err = attributevalue.UnmarshalListOfMaps(response.Items, q)
	if err != nil {
		log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
		return o
	}

	return o
}
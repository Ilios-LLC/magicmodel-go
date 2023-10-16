package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func (o *Operator) Where(q interface{}, k string, v interface{}) *Operator {
	if o.Err != nil {
		return o
	}
	name := parseModelName(q)

	cond := expression.Key("Type").Equal(expression.Value(name))
	cond2 := expression.Name(k).Equal(expression.Value(v))
	expr, err := expression.NewBuilder().WithKeyCondition(cond).WithFilter(cond2).Build()
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

	return o
}

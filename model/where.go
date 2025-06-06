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
	name, err := ParseModelName(q)
	if err != nil {
		o.Err = err
		return o
	}

	err = validateInputSlice(q, "Where", name)
	if err != nil {
		o.Err = err
		return o
	}

	cond := expression.Key("Type").Equal(expression.Value(name))
	cond2 := expression.Name(k).Equal(expression.Value(v))
	softDeleteCond := expression.Not(expression.Name("DeletedAt").AttributeExists())
	sofDeleteCond2 := expression.Not(expression.Name("DeletedAt").NotEqual(expression.Value(nil)))
	expr, err := expression.NewBuilder().WithKeyCondition(cond).WithFilter(cond2.And(softDeleteCond.Or(sofDeleteCond2))).Build()
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

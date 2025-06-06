package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"reflect"
	"time"
)

func (o *Operator) Create(q interface{}) *Operator {
	if o.Err != nil {
		return o
	}

	name, err := ParseModelName(q)
	if err != nil {
		o.Err = err
		return o
	}

	err = ValidateInput(q, "Create", name)
	if err != nil {
		o.Err = err
		return o
	}

	payload := reflect.ValueOf(q).Elem()

	if payload.FieldByName("ID").String() != "" {
		o.Err = fmt.Errorf("encountered an error during Create operations: item already exists. try the update method instead")
		return o
	}

	t := time.Now()

	payload.FieldByName("Type").SetString(name)
	payload.FieldByName("ID").SetString(uuid.New().String())
	payload.FieldByName("CreatedAt").Set(reflect.ValueOf(t))
	payload.FieldByName("UpdatedAt").Set(reflect.ValueOf(t))

	av, err := attributevalue.MarshalMap(q)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Create operations: %v", err)
		return o
	}

	// Use the operator's db field if available, otherwise fall back to the global svc
	dbClient := o.db
	tableName := o.tableName
	if dbClient == nil {
		dbClient = svc
		tableName = dynamoDBTableName
	}

	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Create operations: %v", err)
		return o
	}

	return o
}

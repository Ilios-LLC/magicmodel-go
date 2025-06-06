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

func (o *Operator) Save(q interface{}) *Operator {
	if o.Err != nil {
		return o
	}

	name := parseModelName(q)
	err := validateInput(q, "Save", name)
	if err != nil {
		o.Err = err
		return o
	}

	payload := reflect.ValueOf(q).Elem()

	id := payload.FieldByName("ID").String()
	if id == "" {
		t := time.Now()
		payload.FieldByName("Type").SetString(name)
		payload.FieldByName("ID").SetString(uuid.New().String())
		payload.FieldByName("CreatedAt").Set(reflect.ValueOf(t))
		payload.FieldByName("UpdatedAt").Set(reflect.ValueOf(t))
	}

	av, err := attributevalue.MarshalMap(q)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Save operation: %v", err)
		return o
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(dynamoDBTableName),
		Item:      av,
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Save operation: %v", err)
		return o
	}

	return o
}

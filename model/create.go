package model

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"reflect"
	"time"
)

func (o *Operator) Create(q interface{}) *Operator {
	log.Debug().Msg("Create")
	payload := reflect.ValueOf(q).Elem()

	if payload.FieldByName("ID").String() != "" {
		log.Debug().Msg("Item already exists, skipping")
		return o
	}

	name := parseModelName(q)
	t := time.Now()

	log.Debug().Msg("PK is empty, generating")
	log.Debug().Msg(payload.FieldByName("CreatedAt").String())
	payload.FieldByName("Type").SetString(name)
	payload.FieldByName("ID").SetString(uuid.New().String())
	payload.FieldByName("CreatedAt").Set(reflect.ValueOf(t))
	payload.FieldByName("UpdatedAt").Set(reflect.ValueOf(t))

	av, err := attributevalue.MarshalMap(q)
	if err != nil {
		log.Err(err).Msg("failed to marshal map")
		return o
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(dynamoDBTableName),
		Item:      av,
	})

	if err != nil {
		panic(err)
	}

	return o
}

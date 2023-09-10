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
	payload := map[string]types.AttributeValue{
		"Type": &types.AttributeValueMemberS{Value: "movie"},
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

	//thing := make(map[string]string)
	//err = attributevalue.UnmarshalMap(out.Item, &thing)
	//if err != nil {
	//	panic(err)
	//}

	//val := reflect.ValueOf(q).Elem()
	//for i := 0; i < val.NumField(); i++ {
	//	fieldValue := val.Field(i)
	//	fieldType := val.Type().Field(i).Type
	//	if fieldType.String() == "magicmodel_go.Model" {
	//		createdAt, err := time.Parse(timeLayout, thing["CreatedAt"])
	//		if err != nil {
	//			panic(err)
	//		}
	//
	//		updatedAt, err := time.Parse(timeLayout, thing["UpdatedAt"])
	//		if err != nil {
	//			panic(err)
	//		}
	//
	//		fieldValue.FieldByName("PK").SetString(thing["PK"])
	//		fieldValue.FieldByName("SK").SetString(thing["SK"])
	//		fieldValue.FieldByName("CreatedAt").Set(reflect.ValueOf(createdAt))
	//		fieldValue.FieldByName("UpdatedAt").Set(reflect.ValueOf(updatedAt))
	//	}
	//}

	err = attributevalue.UnmarshalMap(out.Item, q)
	if err != nil {
		panic(err)
	}

	fmt.Println("thing")
}

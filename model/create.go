package model

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsTypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/google/uuid"
	"os"
	"strings"

	//"github.com/rs/zerolog/log"
	"reflect"
	"time"
)

func (o *Operator) Create(q interface{}) *Operator {
	if o.Err != nil {
		return o
	}
	payload := reflect.ValueOf(q).Elem()

	if payload.FieldByName("ID").String() != "" {
		o.Err = fmt.Errorf("encountered an error during Create operation: item already exists. Try the update method instead")
		return o
	}

	name := parseModelName(q)
	t := time.Now()

	payload.FieldByName("Type").SetString(name)
	payload.FieldByName("ID").SetString(uuid.New().String())
	payload.FieldByName("CreatedAt").Set(reflect.ValueOf(t))
	payload.FieldByName("UpdatedAt").Set(reflect.ValueOf(t))

	//val := reflect.ValueOf(q)
	switch payload.Kind() {
	case reflect.Struct:
		//for i := 0; i < payload.Len(); i++ {
		//	otherVal := payload.Index(i)
		types := payload.Type()

		for i2 := 0; i2 < payload.NumField(); i2++ {
			tag := payload.Type().Field(i2).Tag.Get("magicmodel")

			fmt.Println(types.Field(i2).Index[0], types.Field(i2).Name, payload.Field(i2)) // 2 Secret thisIsASecret
			fmt.Println(payload.String())                                                  //<main.MyStruct Value>
			fmt.Println(payload.FieldByName(payload.Type().Field(i2).Name).String())       //thisIsASecret
			if strings.Contains(tag, "sensitive") {
				// first make sure we got a kms key id
				if o.KmsKeyId == nil {
					o.Err = fmt.Errorf("encountered an error during All operation: field tagged with `sensitive` but no KMS key ID was provided")
					return o
				}
				// encrypt the field
				kmsClient := kms.NewFromConfig(o.AwsCfg)
				str := payload.FieldByName(payload.Type().Field(i2).Name).String()
				result, err := kmsClient.Encrypt(context.Background(), &kms.EncryptInput{Plaintext: []byte(str), EncryptionAlgorithm: kmsTypes.EncryptionAlgorithmSpecSymmetricDefault, KeyId: o.KmsKeyId})
				if err != nil {
					fmt.Println("Got error decrypting data: ", err)
					os.Exit(1)
				}

				output := base64.StdEncoding.EncodeToString(result.CiphertextBlob)
				payload.FieldByName(payload.Type().Field(i2).Name).SetString(output)
			}
		}
	}

	av, err := attributevalue.MarshalMap(q)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Create operations: %v", err)
		return o
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(dynamoDBTableName),
		Item:      av,
	})

	if err != nil {
		o.Err = fmt.Errorf("encountered an error during Create operations: %v", err)
		return o
	}

	return o
}

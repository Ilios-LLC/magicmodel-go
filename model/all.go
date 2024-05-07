package model

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsTypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"os"
	"reflect"
	"strings"
)

func (o *Operator) All(q interface{}) *Operator {
	if o.Err != nil {
		return o
	}
	name := parseModelName(q)
	// get the fields where `magicmodel:"sensitive"` is set
	//var sensitiveFields []string
	//val := reflect.ValueOf(q)
	//for i := 0; i < val.NumField(); i++ {
	//	tag := val.Type().Field(i).Tag.Get("magicmodel")
	//	// Field name: val.Type().Field(i).Name
	//	if strings.Contains(tag, "sensitive") {
	//		sensitiveFields = append(sensitiveFields, val.Type().Field(i).Name)
	//	}
	//	fmt.Printf("Field: %s, Tag: %s\n", val.Type().Field(i).Name, tag)
	//	// Implement your validation logic here based on the tag value.
	//}

	cond := expression.Key("Type").Equal(expression.Value(name))
	softDeleteCond := expression.Not(expression.Name("DeletedAt").AttributeExists())
	sofDeleteCond2 := expression.Not(expression.Name("DeletedAt").NotEqual(expression.Value(nil)))
	expr, err := expression.NewBuilder().WithKeyCondition(cond).WithFilter(softDeleteCond.Or(sofDeleteCond2)).Build()
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during All operations: %v", err)
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
		o.Err = fmt.Errorf("encountered an error during All operations: %v", err)
		return o
	}

	err = attributevalue.UnmarshalListOfMaps(response.Items, q)
	if err != nil {
		o.Err = fmt.Errorf("encountered an error during All operations: %v", err)
		return o
	}

	val := reflect.ValueOf(q)
	switch val.Kind() {
	case reflect.Pointer:
		for i := 0; i < reflect.Indirect(val).Len(); i++ {
			otherVal := reflect.Indirect(val).Index(i)
			types := otherVal.Type()

			for i2 := 0; i2 < otherVal.NumField(); i2++ {
				tag := otherVal.Type().Field(i2).Tag.Get("magicmodel")

				fmt.Println(types.Field(i2).Index[0], types.Field(i2).Name, otherVal.Field(i2))              // 2 Secret thisIsASecret
				fmt.Println(otherVal.String())                                                               //<main.MyStruct Value>
				fmt.Println(reflect.Indirect(otherVal).FieldByName(otherVal.Type().Field(i2).Name).String()) //thisIsASecret
				if strings.Contains(tag, "sensitive") {
					// first make sure we got a kms key id
					if o.KmsKeyId == nil {
						o.Err = fmt.Errorf("encountered an error during All operation: field tagged with `sensitive` but no KMS key ID was provided")
						return o
					} // Decrypt the data
					kmsClient := kms.NewFromConfig(o.AwsCfg)

					str := reflect.Indirect(otherVal).FieldByName(otherVal.Type().Field(i2).Name).String()
					uDec, _ := base64.URLEncoding.DecodeString(str)
					fmt.Println(string(uDec))
					result, err := kmsClient.Decrypt(context.Background(), &kms.DecryptInput{CiphertextBlob: uDec, EncryptionAlgorithm: kmsTypes.EncryptionAlgorithmSpecSymmetricDefault, KeyId: o.KmsKeyId})
					if err != nil {
						fmt.Println("Got error decrypting data: ", err)
						os.Exit(1)
					}

					//uEnc := base64.URLEncoding.EncodeToString(result.Plaintext)
					//fmt.Println(uEnc)
					//uDec, _ := base64.URLEncoding.DecodeString(uEnc)
					//fmt.Println(string(uDec))
					output := string(result.Plaintext)
					reflect.Indirect(otherVal).FieldByName(otherVal.Type().Field(i2).Name).SetString(output)
				}
				//fmt.Printf("Field: %s, Tag: %s\n", val.Type().Field(i).Name, tag)
				// Implement your validation logic here based on the tag value.
			}
		}
	}
	//for idx := range reflect.ValueOf(q).Elem().Interface().([]interface{}) {
	//
	//}

	return o
}
